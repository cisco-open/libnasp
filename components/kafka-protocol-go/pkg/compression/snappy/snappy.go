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

package snappy

import (
	"bytes"
	"strconv"
	"strings"

	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"emperror.dev/errors"
	xerialsnappy "github.com/eapache/go-xerial-snappy"
)

func NewReader(r *typesbytes.ChunkReader) (*bytes.Reader, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't read data to be decompressed")
	}

	decodedBytes, err := xerialsnappy.Decode(buf.Bytes())
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't decompress bytes with snappy")
	}

	return bytes.NewReader(decodedBytes), nil
}

type Writer struct {
	buf *bytes.Buffer

	wr *typesbytes.SliceWriter
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	n = len(p)

	for len(p) > 0 {
		off, err := w.buf.Write(p)
		if err != nil {
			return 0, errors.WrapIf(err, "couldn't write to internal buffer")
		}

		p = p[off:]
	}

	return n, nil
}

func (w *Writer) Close() error {
	return w.flush()
}

func (w *Writer) flush() error {
	compressed := xerialsnappy.EncodeStream(nil, w.buf.Bytes())
	n, err := w.wr.Write(compressed)
	if err != nil {
		return errors.WrapIf(err, "couldn't write to underlying Writer")
	}
	if n != len(compressed) {
		return errors.New(strings.Join([]string{"expected to write", strconv.Itoa(len(compressed)), "bytes underlying Writer, but only wrote", strconv.Itoa(n)}, " "))
	}

	w.buf.Reset()
	return nil
}

func NewWriter(w *typesbytes.SliceWriter) (*Writer, error) {
	return &Writer{
		wr:  w,
		buf: bytes.NewBuffer(make([]byte, 0, 32*1024)),
	}, nil
}
