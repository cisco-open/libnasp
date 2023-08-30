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

package serialization

import (
	"compress/gzip"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/compression/snappy"
	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
)

type Writer struct {
	sliceWriter *bytes.SliceWriter

	gzipEncoder   *gzip.Writer
	zstdEncoder   *zstd.Encoder
	snappyEncoder *snappy.Writer
	lz4Encoder    *lz4.Writer

	encoderType int8
}

func (w *Writer) Write(p []byte) (n int, err error) {
	switch w.encoderType {
	case 0:
		return w.sliceWriter.Write(p)
	case 1:
		return w.gzipEncoder.Write(p)
	case 2:
		return w.snappyEncoder.Write(p)
	case 3:
		return w.lz4Encoder.Write(p)
	case 4:
		return w.zstdEncoder.Write(p)
	default:
		return 0, errors.New(strings.Join([]string{"unknown encoder type id:", strconv.Itoa(int(w.encoderType))}, " "))
	}
}

func (w *Writer) Close() error {
	switch w.encoderType {
	case 0:
		return nil
	case 1:
		return w.gzipEncoder.Close()
	case 2:
		return w.snappyEncoder.Close()
	case 3:
		return w.lz4Encoder.Close()
	case 4:
		return w.zstdEncoder.Close()
	default:
		return errors.New(strings.Join([]string{"unknown encoder type id:", strconv.Itoa(int(w.encoderType))}, " "))
	}
}

func NewWriterWithSliceWriter(wr *bytes.SliceWriter) Writer {
	return Writer{sliceWriter: wr, encoderType: 0}
}

func NewWriterWithGzipEncoder(encoder *gzip.Writer) Writer {
	return Writer{gzipEncoder: encoder, encoderType: 1}
}

func NewWriterWithSnappyEncoder(encoder *snappy.Writer) Writer {
	return Writer{snappyEncoder: encoder, encoderType: 2}
}

func NewWriterWithLz4Encoder(encoder *lz4.Writer) Writer {
	return Writer{lz4Encoder: encoder, encoderType: 3}
}

func NewWriterWithZstdEncoder(encoder *zstd.Encoder) Writer {
	return Writer{zstdEncoder: encoder, encoderType: 4}
}
