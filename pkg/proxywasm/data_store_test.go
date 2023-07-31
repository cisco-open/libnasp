// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxywasm_test

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cisco-open/nasp/pkg/proxywasm"
)

type testHTTPGetter struct {
	content *bytes.Buffer
}

func (g *testHTTPGetter) Get(url string) (resp *http.Response, err error) {
	return &http.Response{
		Body: io.NopCloser(g.content),
	}, nil
}

func TestURLDataSourceTest(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	param := []byte("wasm content")
	ds := proxywasm.NewURLDataSource("", proxywasm.URLDataSourceWithHTTPGetter(&testHTTPGetter{
		content: bytes.NewBuffer(param),
	}))
	content, err := ds.Get()

	assert.NoError(err)
	assert.Equal(content, param)
}

func TestBytesDataSource(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	param := []byte("wasm content")
	ds := proxywasm.NewBytesDataSource(param)
	content, err := ds.Get()

	assert.NoError(err)
	assert.Equal(content, param)
}

func TestFileDataSource(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	param := []byte("wasm content")
	ds := proxywasm.NewFileDataSource(os.DirFS("."), "testdata/filter.wasm")
	content, err := ds.Get()

	assert.NoError(err)
	assert.Equal(content, param)
}
