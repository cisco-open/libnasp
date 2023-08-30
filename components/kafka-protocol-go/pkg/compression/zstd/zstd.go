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

package zstd

import (
	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"emperror.dev/errors"
	"github.com/klauspost/compress/zstd"
)

func NewReader(r *typesbytes.ChunkReader) (*zstd.Decoder, error) {
	zstdr, err := zstd.NewReader(r, zstd.WithDecoderConcurrency(1), zstd.WithDecoderLowmem(false))
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't create zstd reader")
	}
	return zstdr, nil
}

func NewWriter(w *typesbytes.SliceWriter) (*zstd.Encoder, error) {
	zstdrw, err := zstd.NewWriter(w, zstd.WithEncoderConcurrency(1), zstd.WithLowerEncoderMem(false))
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't create zstd writer")
	}

	return zstdrw, nil
}
