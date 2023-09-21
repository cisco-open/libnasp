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
	"strconv"
	"strings"

	"github.com/pierrec/lz4/v4"

	gobytes "bytes"

	"emperror.dev/errors"
	"github.com/klauspost/compress/zstd"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/compression/gzip"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
)

type Reader struct {
	chunkReader *bytes.ChunkReader

	gzipDecoder   *gzip.Reader
	zstdDecoder   *zstd.Decoder
	snappyDecoder *gobytes.Reader
	lz4Decoder    *lz4.Reader

	decoderType int8
}

func (r *Reader) Read(p []byte) (n int, err error) {
	switch r.decoderType {
	case 0:
		return r.chunkReader.Read(p)
	case 1:
		return r.gzipDecoder.Read(p)
	case 2:
		return r.snappyDecoder.Read(p)
	case 3:
		return r.lz4Decoder.Read(p)
	case 4:
		return r.zstdDecoder.Read(p)
	default:
		return 0, errors.New(strings.Join([]string{"unknown decoder type id:", strconv.Itoa(int(r.decoderType))}, " "))
	}
}

func (r *Reader) ReadByte() (byte, error) {
	switch r.decoderType {
	case 0:
		return r.chunkReader.ReadByte()
	case 1:
		return r.gzipDecoder.ReadByte()
	case 2:
		return r.snappyDecoder.ReadByte()
	case 3:
		return 0, errors.New("ByteReader.ReadByte() not supported by lz4")
	case 4:
		return 0, errors.New("ByteReader.ReadByte() not supported by zstd")
	default:
		return 0, errors.New(strings.Join([]string{"unknown decoder type id:", strconv.Itoa(int(r.decoderType))}, " "))
	}
}

func (r *Reader) Close() error {
	switch r.decoderType {
	case 0:
		return nil
	case 1:
		return r.gzipDecoder.Close()
	case 2:
		return nil
	case 3:
		return nil
	case 4:
		return nil
	default:
		return errors.New(strings.Join([]string{"unknown decoder type id:", strconv.Itoa(int(r.decoderType))}, " "))
	}
}

func NewReaderWithChunkReader(r *bytes.ChunkReader) Reader {
	return Reader{chunkReader: r, decoderType: 0}
}

func NewReaderWithGzipDecoder(decoder *gzip.Reader) Reader {
	return Reader{gzipDecoder: decoder, decoderType: 1}
}

func NewReaderWithSnappyDecoder(decoder *gobytes.Reader) Reader {
	return Reader{snappyDecoder: decoder, decoderType: 2}
}

func NewReaderWithWithLz4Decoder(decoder *lz4.Reader) Reader {
	return Reader{lz4Decoder: decoder, decoderType: 3}
}

func NewReaderWithZstdDecoder(decoder *zstd.Decoder) Reader {
	return Reader{zstdDecoder: decoder, decoderType: 4}
}
