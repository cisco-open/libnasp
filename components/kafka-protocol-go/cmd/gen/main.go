//  Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        https://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig/v3"

	goimports "golang.org/x/tools/imports"

	"emperror.dev/errors"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/assets"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/internal/protocol"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/internal/util"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/fields"
)

const (
	moduleName = "github.com/cisco-open/libnasp/components/kafka-protocol-go"

	requestSpecType  = "request"
	responseSpecType = "response"
	headerSpecType   = "header"
	dataSpecType     = "data"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("missing <repo-root-dir> option")
		fmt.Println("usage: kafka-protocol-go <repo-root-dir")
		os.Exit(1)
	}
	err := generate(args[1])
	if err != nil {
		panic(err)
	}
}

type templateData struct {
	Package     string
	Imports     []string
	Name        string
	ApiKey      int
	Constraints structConstraints
	Fields      []fieldTemplateData
}

type fieldTemplateData struct {
	Name            string
	PropertyName    string
	Type            string
	Description     string
	IsArrayOfStruct bool
	IsArray         bool
	IsStruct        bool

	FieldType        string
	FieldContext     fields.Context
	ArrayElementType string
	Ignorable        bool
}

type messageFactoryTemplateData struct {
	Package           string
	Imports           []string
	RequestsTypeSpec  []messageTypeSpecTemplateData
	ResponsesTypeSpec []messageTypeSpecTemplateData
}

type messageTypeSpecTemplateData struct {
	ApiKey int
	Pkg    string
	Name   string
}

func generate(repoRootDirPath string) error {
	utilFuncMap := template.FuncMap{
		"CamelCase":               util.CamelCase,
		"PascalCase":              util.PascalCase,
		"EscapeGoReservedKeyword": util.EscapeGoReservedKeyword,
	}

	tmpl, err := template.New("MessageData").
		Funcs(sprig.TxtFuncMap()).
		Funcs(utilFuncMap).
		Parse(string(assets.StructProcessorTemplate))
	if err != nil {
		return errors.WrapIff(err, "couldn't parse struct processor template")
	}

	tmplFactory, err := template.New("MessageFactory").
		Funcs(sprig.TxtFuncMap()).
		Parse(string(assets.MessageFactoryTemplate))
	if err != nil {
		return errors.WrapIff(err, "couldn't parse message factory template")
	}

	var requestsTypeSpec, responsesTypeSpec []messageTypeSpecTemplateData

	targetSrcDir := "pkg/protocol/messages"
	sourceCode := make(map[string][]byte)

	kafkaMessageSpecs, err := protocol.LoadProtocolSpec(assets.KafkaProtocolSpec)
	if err != nil {
		return errors.WrapIff(err, "couldn't load kafka protocol spec")
	}

	for _, spec := range kafkaMessageSpecs {
		switch spec.Type {
		case headerSpecType, requestSpecType, responseSpecType:
			pkg, structName, err := generateMessageProcessors(tmpl, targetSrcDir, spec, sourceCode)
			if err != nil {
				return errors.WrapIff(err, "generating message processors for %s failed", spec.Name)
			}

			if spec.Type == requestSpecType {
				requestsTypeSpec = append(requestsTypeSpec, messageTypeSpecTemplateData{
					ApiKey: spec.APIKey,
					Pkg:    pkg,
					Name:   structName,
				})
			}
			if spec.Type == responseSpecType {
				responsesTypeSpec = append(responsesTypeSpec, messageTypeSpecTemplateData{
					ApiKey: spec.APIKey,
					Pkg:    pkg,
					Name:   structName,
				})
			}
		case dataSpecType:
		default:
			return errors.Errorf("unsupported kafka protocol message type %q", spec.Type)
		}
	}
	err = generateMessageFactoryCode(tmplFactory, targetSrcDir, requestsTypeSpec, responsesTypeSpec, sourceCode)
	if err != nil {
		return errors.WrapIf(err, "generating message factories failed")
	}

	// write to file
	for filePath, src := range sourceCode {
		filePath = path.Join(repoRootDirPath, filePath)
		err := os.MkdirAll(path.Dir(filePath), os.ModePerm)
		if err != nil {
			return err
		}

		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}

		_, err = f.Write(src)
		if err != nil {
			errClose := f.Close()
			return errors.Combine(
				errors.WrapIff(err, "couldn't write message processor to file %s", f.Name()),
				errClose)
		}

		errClose := f.Close()
		if errClose != nil {
			return errClose
		}
	}

	return nil
}

func generateMessageProcessors(tmpl *template.Template, baseDir string, spec *protocol.MessageSpec, sourceCode map[string][]byte) (string, string, error) {
	structName := spec.Name
	pkg := ""
	idx := strings.LastIndexFunc(structName, unicode.IsUpper)
	if idx != -1 {
		pkg = strings.ToLower(structName[:idx])
		baseDir = path.Join(baseDir, pkg)
		structName = structName[idx:]
	}

	messageContext, err := contextFromMessageSpec(spec)
	if err != nil {
		return "", "", errors.WrapIff(err, "couldn't get context for %s", spec.Name)
	}

	commonStructsDir := path.Join(baseDir, "common")
	for _, commonStruct := range spec.CommonStructs {
		constraints, err := structConstraintsFromStructSpec(commonStruct)
		if err != nil {
			return "", "", errors.WrapIff(err, "couldn't get version constraints for %s", commonStruct.Name)
		}

		// disregard the version of common structs from spec as that is controlled by the referencing field of the parent struct
		constraints.LowestSupportedVersion = -1
		constraints.HighestSupportedVersion = -1

		// flexible versions range is inherited from parent message spec
		constraints.LowestSupportedFlexVersion = messageContext.LowestSupportedFlexVersion
		constraints.HighestSupportedFlexVersion = messageContext.HighestSupportedFlexVersion

		err = generateStructProcessor(tmpl, "", commonStructsDir, commonStruct.Name, -1, messageContext, constraints, commonStruct.Fields, sourceCode)
		if err != nil {
			return "", "", errors.WrapIff(err, "couldn't generate code for %s", spec.Name)
		}
	}

	constraints, err := structConstraintsFromMessageSpec(*spec)
	if err != nil {
		return "", "", errors.WrapIff(err, "couldn't get version constraints for %s", spec.Name)
	}

	apiKey := -1
	if spec.Type == requestSpecType || spec.Type == responseSpecType {
		apiKey = spec.APIKey
	}
	err = generateStructProcessor(tmpl, commonStructsDir, baseDir, structName, apiKey, messageContext, constraints, spec.Fields, sourceCode)
	if err != nil {
		return "", "", errors.WrapIff(err, "couldn't generate code for %s", spec.Name)
	}

	return pkg, structName, nil
}
func generateStructProcessor(tmpl *template.Template,
	commonStructsDir, baseDir, structName string, apiKey int,
	context *fields.Context,
	constraints structConstraints,
	fieldSpecs []protocol.MessageFieldSpec,
	sourceCode map[string][]byte) error {
	fields := make([]fieldTemplateData, 0, len(fieldSpecs)+1)
	for i := range fieldSpecs {
		fieldSpec := fieldSpecs[i]

		subPkg := "common"
		if len(fieldSpec.Fields) > 0 {
			subPkg = strings.ToLower(structName)
			fieldStructTypeName, err := getStructTypeName(fieldSpec.Type)
			if err != nil {
				return errors.WrapIff(err, "couldn't determine the struct type name for field %s", fieldSpec.Name)
			}

			fieldStructConstraints, err := structConstraintsFromFieldSpec(constraints, fieldSpec)
			if err != nil {
				return errors.WrapIff(err, "couldn't get version constraints for %s", fieldSpec.Name)
			}
			fieldContext, err := contextFromFieldSpec(context, &fieldSpec)
			if err != nil {
				return errors.WrapIff(err, "couldn't get context for %s", fieldSpec.Name)
			}

			err = generateStructProcessor(tmpl, commonStructsDir, path.Join(baseDir, subPkg), fieldStructTypeName, -1, fieldContext, fieldStructConstraints, fieldSpec.Fields, sourceCode)
			if err != nil {
				return errors.WrapIff(err, "couldn't generate code for %s", fieldStructTypeName)
			}
		}

		fieldGoType, err := goType(subPkg, fieldSpec.Type)
		if err != nil {
			return errors.WrapIff(err, "couldn't determine the Go type for field %s", fieldSpec.Name)
		}

		fieldName := util.EscapeGoReservedKeyword(util.CamelCase(fieldSpec.Name))
		propertyName := util.PascalCase(fieldSpec.Name)

		elementType := ""
		arrayOfStruct, err := isArrayOfStruct(fieldSpec.Type)
		if err != nil {
			return errors.WrapIff(err, "couldn't determine if the %q field is of array of struct type", fieldSpec.Name)
		}

		fieldType := " fields." + util.PascalCase(fieldSpec.Type)
		if arrayOfStruct {
			elementType = arrayElementType(fieldSpec.Type)
			elementGoType, err := goType(subPkg, elementType)
			if err != nil {
				return errors.WrapIff(err, "couldn't determine the Go type for %s", elementType)
			}
			fieldType = " fields.ArrayOfStruct[" + elementGoType + "," + "*" + elementGoType + "]"
		} else if isArray(fieldSpec.Type) {
			elementType = arrayElementType(fieldSpec.Type)
			elementGoType, err := goType(subPkg, elementType)
			if err != nil {
				return errors.WrapIff(err, "couldn't determine the Go type for %s", elementType)
			}
			elementType = " fields." + util.PascalCase(elementType)
			fieldType = " fields.Array[" + elementGoType + ", *" + elementType + "]"
		}

		fieldContext, err := contextFromFieldSpec(context, &fieldSpec)
		if err != nil {
			return errors.WrapIff(err, "couldn't get context for %s", fieldSpec.Name)
		}

		fields = append(fields, fieldTemplateData{
			Name:             fieldName,
			PropertyName:     propertyName,
			Type:             fieldGoType,
			Description:      fieldSpec.About,
			IsArray:          isArray(fieldSpec.Type),
			IsArrayOfStruct:  arrayOfStruct,
			IsStruct:         isStruct(fieldSpec.Type),
			FieldType:        fieldType,
			FieldContext:     *fieldContext,
			ArrayElementType: elementType,
			Ignorable:        fieldSpec.Ignorable,
		})
	}

	err := generateStructProcessorCode(tmpl, commonStructsDir, baseDir, structName, apiKey, constraints, fields, sourceCode)
	if err != nil {
		return errors.WrapIff(err, "couldn't generate code for %s", structName)
	}

	return nil
}

func generateMessageFactoryCode(tmpl *template.Template, baseDir string, requestsTypeSpec, responsesTypeSpec []messageTypeSpecTemplateData, sourceCode map[string][]byte) error {
	key := path.Join(baseDir, "/", "requests.go")
	if _, ok := sourceCode[key]; ok {
		return errors.Errorf("type conflict, %q already exists", key)
	}

	pkg := path.Base(baseDir)
	importsSet := make(map[string]struct{})
	importsSet["\"emperror.dev/errors\""] = struct{}{}
	importsSet["\"strconv\""] = struct{}{}
	importsSet["\"strings\""] = struct{}{}
	importsSet["\""+moduleName+"/pkg/protocol\""] = struct{}{}

	for _, spec := range requestsTypeSpec {
		importsSet[fmt.Sprintf("%q", path.Join(moduleName, baseDir, spec.Pkg))] = struct{}{}
	}
	for _, spec := range responsesTypeSpec {
		importsSet[fmt.Sprintf("%q", path.Join(moduleName, baseDir, spec.Pkg))] = struct{}{}
	}
	messageFactories := messageFactoryTemplateData{
		Package:           pkg,
		RequestsTypeSpec:  requestsTypeSpec,
		ResponsesTypeSpec: responsesTypeSpec,
	}

	for importPath := range importsSet {
		messageFactories.Imports = append(messageFactories.Imports, importPath)
	}
	sort.Strings(messageFactories.Imports)

	sort.Slice(messageFactories.RequestsTypeSpec, func(i, j int) bool {
		return messageFactories.RequestsTypeSpec[i].ApiKey < messageFactories.RequestsTypeSpec[j].ApiKey
	})
	sort.Slice(messageFactories.ResponsesTypeSpec, func(i, j int) bool {
		return messageFactories.ResponsesTypeSpec[i].ApiKey < messageFactories.ResponsesTypeSpec[j].ApiKey
	})

	var raw bytes.Buffer
	if err := tmpl.Execute(&raw, messageFactories); err != nil {
		return errors.WrapIff(err, "unable to render template %s", tmpl.Name())
	}

	src, err := goimports.Process("", raw.Bytes(), nil)
	if err != nil {
		return errors.WrapIff(err, "couldn't format generated source code: %s", raw.String())
	}

	sourceCode[key] = src
	return nil
}

func generateStructProcessorCode(tmpl *template.Template, commonStructsDir, baseDir, structName string, apiKey int, constraints structConstraints, fields []fieldTemplateData, sourceCode map[string][]byte) error {
	key := path.Join(baseDir, strings.ToLower(structName)+".go")
	if _, ok := sourceCode[key]; ok {
		return errors.Errorf("type conflict, %q already exists", key)
	}

	pkg := path.Base(baseDir)
	importsSet := make(map[string]struct{})
	importsSet["\"bytes\""] = struct{}{}
	importsSet["\"strconv\""] = struct{}{}
	importsSet["\"strings\""] = struct{}{}
	importsSet["\"emperror.dev/errors\""] = struct{}{}
	importsSet["\""+moduleName+"/pkg/pools\""] = struct{}{}
	importsSet["\""+moduleName+"/pkg/protocol/types/fields\""] = struct{}{}
	importsSet["\""+moduleName+"/pkg/protocol/types/varint\""] = struct{}{}
	importsSet["typesbytes \""+moduleName+"/pkg/protocol/types/bytes\""] = struct{}{}

	for _, f := range fields {
		if !strings.ContainsRune(f.Type, '.') {
			continue
		}
		p := strings.Split(f.Type, ".")[0]
		if strings.HasPrefix(p, "[]") {
			p = strings.TrimSpace(strings.TrimPrefix(p, "[]"))
			if len(p) == 0 {
				return errors.Errorf("couldn't determine package name from %s", f.Type)
			}
		}
		if strings.HasPrefix(p, "*") {
			p = strings.TrimSpace(strings.TrimPrefix(p, "*"))
			if len(p) == 0 {
				return errors.Errorf("couldn't determine package name from %s", f.Type)
			}
		}

		var importPath string

		switch p {
		case "common":
			if commonStructsDir == "" {
				return errors.New("missing common structs dir")
			}
			importPath = fmt.Sprintf("%q", path.Join(moduleName, commonStructsDir))
		case "fields":
			importPath = "\"" + moduleName + "/pkg/protocol/types/fields\""
		default:
			importPath = fmt.Sprintf("%q", path.Join(moduleName, baseDir, p))
		}

		importsSet[importPath] = struct{}{}
	}

	imports := make([]string, 0, len(importsSet))
	for k := range importsSet {
		imports = append(imports, k)
	}
	sort.Strings(imports)

	data := templateData{
		Package:     pkg,
		Imports:     imports,
		Name:        structName,
		ApiKey:      apiKey,
		Constraints: constraints,
		Fields:      fields,
	}

	src, err := generateProcessorCode(tmpl, data)
	if err != nil {
		return errors.WrapIff(err, "couldn't generate code for %s", structName)
	}
	sourceCode[key] = src

	return nil
}

func generateProcessorCode(tmpl *template.Template, data templateData) ([]byte, error) {
	var raw bytes.Buffer
	if err := tmpl.Execute(&raw, data); err != nil {
		return nil, errors.WrapIff(err, "unable to render template %s", tmpl.Name())
	}

	src, err := goimports.Process("", raw.Bytes(), nil)
	if err != nil {
		return nil, errors.WrapIff(err, "couldn't format generated source code: %s", raw.String())
	}
	return src, nil
}

func goType(pkg, t string) (string, error) {
	switch t {
	case Bool:
		return "bool", nil
	case Int8:
		return "int8", nil
	case Int16:
		return "int16", nil
	case Uint16:
		return "uint16", nil
	case Int32:
		return "int32", nil
	case Int64:
		return "int64", nil
	case Float64:
		return "float64", nil
	case String:
		return "fields.NullableString", nil
	case Uuid:
		return "fields.UUID", nil
	case Bytes:
		return "[]byte", nil
	case Records:
		return "fields.RecordBatches", nil
	default:
		if strings.HasPrefix(t, "[]") {
			elementType := strings.TrimSpace(strings.TrimPrefix(t, "[]"))
			if len(elementType) == 0 {
				return "", errors.Errorf("couldn't determine array element type from %s", t)
			}
			elementGoType, err := goType(pkg, elementType)
			if err != nil {
				return "", errors.Errorf("couldn't determine Go type for array element type %s", elementType)
			}
			return "[]" + elementGoType, nil
		}

		if len(t) > 0 && unicode.IsUpper(rune(t[0])) {
			return pkg + "." + t, nil
		}

		return "", errors.Errorf("unsupported field type %s", t)
	}
}

func getStructTypeName(t string) (string, error) {
	if isArray(t) {
		elementType := strings.TrimSpace(strings.TrimPrefix(t, "[]"))
		if len(elementType) == 0 {
			return "", errors.Errorf("couldn't determine array element type from %s", t)
		}

		structTypeName, err := getStructTypeName(elementType)
		if err != nil {
			return "", errors.Errorf("couldn't determine struct type name from %s", elementType)
		}

		return structTypeName, nil
	}

	if isStruct(t) {
		return t, nil
	}

	return "", errors.Errorf("%s is not struct type", t)
}

func isArrayOfStruct(t string) (bool, error) {
	if !isArray(t) {
		return false, nil
	}

	elementType := strings.TrimSpace(strings.TrimPrefix(t, "[]"))
	if len(elementType) == 0 {
		return false, errors.Errorf("couldn't determine array element type from %s", t)
	}

	return isStruct(elementType), nil
}

func isArray(t string) bool {
	return strings.HasPrefix(t, "[]")
}

func isStruct(t string) bool {
	return len(t) > 0 && unicode.IsUpper(rune(t[0]))
}

func arrayElementType(t string) string {
	return strings.TrimSpace(strings.TrimPrefix(t, "[]"))
}
