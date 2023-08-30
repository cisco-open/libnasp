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

package bytes

import (
	"io"
	"strconv"
	"strings"

	"emperror.dev/errors"
)

// SliceWriter writes bytes to buf up to buf's capacity. In case more bytes are written than buf's capacity
// buf is not resized but rather an error is returned
type SliceWriter struct {
	buf []byte
}

// NewSliceWriter creates a new SliceWriter instance which writes the given b byte slice. It is expected
// that the provided b byte slice has the desired capacity (which won't grow) but a length zero.
func NewSliceWriter(b []byte) SliceWriter {
	return SliceWriter{
		buf: b,
	}
}

func (w *SliceWriter) Write(p []byte) (n int, err error) {
	if len(p) > cap(w.buf)-len(w.buf) {
		return 0, errors.New("couldn't append the provided byte slice to underlying byte stream without exceeding maximum capacity")
	}

	w.buf = append(w.buf, p...)

	return len(p), nil
}

func (w *SliceWriter) WriteAt(p []byte, off int64) (n int, err error) {
	if len(p) == 0 {
		return
	}

	if int64(len(p)) > int64(cap(w.buf))-off {
		return 0, errors.New(strings.Join([]string{"couldn't write the provided byte slice to underlying byte stream at offset", strconv.Itoa(int(off)), "without exceeding maximum capacity"}, " "))
	}

	n = copy(w.buf[off:], p)

	return n, nil
}

func (w *SliceWriter) InsertAt(p []byte, off int64) (n int, err error) {
	if len(p) == 0 {
		return
	}

	if len(p) > cap(w.buf)-len(w.buf) {
		return 0, errors.New(strings.Join([]string{"couldn't insert the provided byte slice to underlying byte stream at offset", strconv.Itoa(int(off)), "without exceeding maximum capacity"}, " "))
	}

	// move w.buf's len to make room for p
	for i := 0; i < len(p); i++ {
		w.buf = append(w.buf, 0)
	}

	moveTo := off + int64(len(p))
	m := int64(len(w.buf)) - moveTo
	n = copy(w.buf[moveTo:], w.buf[off:])
	if int64(n) != m {
		return 0, errors.New(strings.Join([]string{"was able to move only", strconv.Itoa(n), "bytes to make room for insertion, expected to move", strconv.Itoa(int(m))}, " "))
	}

	n = copy(w.buf[off:], p)

	return n, nil
}

func (w *SliceWriter) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errors.New("bytes.SliceWriter.Seek: invalid whence")
	case io.SeekStart:
	case io.SeekCurrent:
		offset += int64(len(w.buf))
	}

	if offset < 0 || offset > int64(cap(w.buf)) {
		return 0, errors.New(strings.Join([]string{"bytes.SliceWriter.Seek: position", strconv.Itoa(int(offset)), "is out of bounds: [0,", strconv.Itoa(cap(w.buf)), "]"}, " "))
	}
	w.buf = w.buf[:offset]

	return offset, nil
}

func (w *SliceWriter) Len() int {
	return len(w.buf)
}

func (w *SliceWriter) Slice(lower, upper int) ([]byte, error) {
	if lower < 0 {
		return nil, errors.New("slice lower bound is negative")
	}
	if upper > cap(w.buf) {
		return nil, errors.New(strings.Join([]string{"slice upper bound", strconv.Itoa(upper), "is out of range [:", strconv.Itoa(cap(w.buf)), "]"}, " "))
	}

	return w.buf[lower:upper], nil
}

func (w *SliceWriter) Bytes() []byte {
	return w.buf
}
