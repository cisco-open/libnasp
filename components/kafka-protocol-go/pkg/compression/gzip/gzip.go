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

package gzip

import (
	"compress/gzip"
	"io"

	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"emperror.dev/errors"
)

type Reader struct {
	rd *gzip.Reader

	buf  [16 * 1024]byte
	r, w int
}

func (r *Reader) Read(p []byte) (int, error) {
	n := 0
	if len(p) == 0 {
		return 0, nil
	}

	if r.r < r.w {
		m := copy(p, r.buf[r.r:r.w])
		r.r += m
		if r.r >= r.w {
			r.r = 0
			r.w = 0
		}

		if m == len(p) {
			return m, nil
		}

		n += m
		if m > 0 {
			p = p[m:]
		}
	}

	if len(p) >= len(r.buf) {
		m, err := r.rd.Read(p)
		return n + m, err
	}

	err := r.fillInternalBuf()
	if err != nil {
		return 0, err
	}

	if r.r >= r.w {
		return n, io.EOF
	}

	m := copy(p, r.buf[r.r:r.w])
	r.r += m
	if r.r >= r.w {
		r.r = 0
		r.w = 0
	}

	return n + m, nil
}

func (r *Reader) ReadByte() (byte, error) {
	if r.r >= r.w {
		r.r = 0
		r.w = 0

		err := r.fillInternalBuf()
		if err != nil {
			return 0, err
		}
	}

	if r.r >= r.w {
		return 0, io.EOF
	}

	b := r.buf[r.r]
	r.r++

	if r.r >= r.w {
		r.r = 0
		r.w = 0
	}

	return b, nil
}

func (r *Reader) Close() error {
	return r.rd.Close()
}

func (r *Reader) fillInternalBuf() error {
	for {
		if r.w == len(r.buf) {
			// internal buffer is full
			return nil
		}
		m, err := r.rd.Read(r.buf[r.w:])
		r.w += m

		if err != nil {
			if errors.Is(err, io.EOF) {
				// no more data available to read into internal buffer
				return nil
			}
			return errors.WrapIf(err, "couldn't read from gzip reader")
		}
	}
}

func NewReader(r *typesbytes.ChunkReader) (*Reader, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.WrapIf(err, "couldn't create gzip reader")
	}

	gz.Multistream(false)

	return &Reader{
		rd: gz,
	}, nil
}

func NewWriter(w *typesbytes.SliceWriter) *gzip.Writer {
	gz := gzip.NewWriter(w)

	return gz
}
