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

package lz4

import (
	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"emperror.dev/errors"
	"github.com/pierrec/lz4/v4"
)

func NewReader(r *typesbytes.ChunkReader) (*lz4.Reader, error) {
	lzr := lz4.NewReader(r)
	err := lzr.Apply(lz4.ConcurrencyOption(1))
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't set lz4  concurrency option")
	}

	return lzr, nil
}

func NewWriter(w *typesbytes.SliceWriter) (*lz4.Writer, error) {
	lz4w := lz4.NewWriter(w)
	err := lz4w.Apply(lz4.ConcurrencyOption(1))
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't set lz4 concurrency option")
	}

	return lz4w, nil
}
