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

package types

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"emperror.dev/errors"
)

func ReadBool(r *bytes.Reader, out *bool) error {
	var buf [1]byte
	n, err := r.Read(buf[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.WrapIf(err, strings.Join([]string{"expected 1 byte, but got", strconv.Itoa(n)}, " "))
		}
		return err
	}

	if n != 1 {
		return errors.New(strings.Join([]string{"expected 1 byte, but got", strconv.Itoa(n)}, " "))
	}

	*out = buf[0] != 0

	return nil
}

func WriteBool(w *typesbytes.SliceWriter, data bool) error {
	var buf [1]byte

	if data {
		buf[0] = 1
	} else {
		buf[0] = 0
	}

	n, err := w.Write(buf[:])

	if err != nil {
		return errors.WrapIf(err, "couldn't write serialized bool")
	}
	if n != 1 {
		return errors.New(strings.Join([]string{"expected to write 1 byte of serialized bool, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}
