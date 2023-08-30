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
	"encoding/binary"
	"io"
	"math"
	"strconv"
	"strings"

	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"emperror.dev/errors"
)

func ReadFloat64(r *bytes.Reader, out *float64) error {
	var buf [8]byte

	n, err := r.Read(buf[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.WrapIf(err, strings.Join([]string{"expected 8 bytes, but got", strconv.Itoa(n)}, " "))
		}
		return err
	}

	if n != 8 {
		return errors.New(strings.Join([]string{"expected 8 bytes, but got", strconv.Itoa(n)}, " "))
	}

	*out = math.Float64frombits(binary.BigEndian.Uint64(buf[:]))

	return nil
}

func WriteFloat64(w *typesbytes.SliceWriter, data float64) error {
	var buf [8]byte

	binary.BigEndian.PutUint64(buf[:], math.Float64bits(data))
	n, err := w.Write(buf[:])

	if err != nil {
		return errors.WrapIf(err, "couldn't write serialized float64")
	}
	if n != 8 {
		return errors.New(strings.Join([]string{"expected to write 8 bytes of serialized float64, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}
