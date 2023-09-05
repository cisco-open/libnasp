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
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/pools"

	"emperror.dev/errors"
)

// NewChunkReader returns a ChunkReader that reads from r
// but stops with EOF after n bytes or when the last byte is read from r in case r has less the n bytes.
func NewChunkReader(r *bytes.Reader, n int64) ChunkReader {
	if int64(r.Len()) < n {
		n = int64(r.Len())
	}

	return ChunkReader{
		br:   r,
		size: n,
		base: r.Size() - int64(r.Len()),
	}
}

// NewSubChunkReader returns a ChunkReader that reads from r
// but stops with EOF after n bytes or when the last byte is read from r in case r has less the n bytes.
func NewSubChunkReader(r *ChunkReader, n int64) ChunkReader {
	if int64(r.Len()) < n {
		n = int64(r.Len())
	}

	return ChunkReader{
		cr:   r,
		size: n,
		base: r.Size() - int64(r.Len()),
	}
}

// A ChunkReader reads from bytes.Reader's unread portion up to size amount of bytes.
// Read returns EOF when size is reached or when the underlying reader returns EOF.
type ChunkReader struct {
	br *bytes.Reader
	cr *ChunkReader

	base int64
	size int64
}

// Len returns the number of bytes of the unread portion of the
// slice.
func (c *ChunkReader) Len() int {
	var size int64
	if c.cr != nil {
		size = c.cr.Size()
	} else {
		size = c.br.Size()
	}

	var l int
	if c.cr != nil {
		l = c.cr.Len()
	} else {
		l = c.br.Len()
	}

	baseCrtPos := size - int64(l) // base reader's current read pos

	return int(c.base + c.size - baseCrtPos)
}

// Size returns the original length of the underlying byte slice.
// Size is the number of bytes available for reading via ReadAt.
// The result is unaffected by any method calls except Reset.
func (c *ChunkReader) Size() int64 { return c.size }

// Read implements the io.Reader interface.
func (c *ChunkReader) Read(b []byte) (n int, err error) {
	remainder := c.Len()
	if remainder <= 0 {
		return 0, io.EOF
	}
	if len(b) > remainder {
		b = b[0:remainder]
	}

	if c.cr != nil {
		return c.cr.Read(b)
	} else {
		return c.br.Read(b)
	}
}

// ReadAt implements the io.ReaderAt interface.
func (c *ChunkReader) ReadAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("bytes.ChunkReader.ReadAt: negative offset")
	}

	if off >= c.size {
		return 0, errors.WrapIf(io.EOF, strings.Join([]string{"bytes.ChunkReader.ReadAt: offset", strconv.Itoa(int(off)), "is beyond byte slice length", strconv.Itoa(int(c.size))}, " "))
	}

	if int64(len(b)) > c.size-off {
		return 0, errors.WrapIf(io.EOF, strings.Join([]string{"bytes.ChunkReader.ReadAt: byte slice doesn't have enough data beyond offset", strconv.Itoa(int(off)), ", expected", strconv.Itoa(len(b)), ", got", strconv.Itoa(int(c.size - off))}, " "))
	}

	if c.cr != nil {
		return c.cr.ReadAt(b, c.base+off)
	} else {
		return c.br.ReadAt(b, c.base+off)
	}
}

// ReadByte implements the io.ByteReader interface.
func (c *ChunkReader) ReadByte() (byte, error) {
	if c.Len() <= 0 {
		return 0, io.EOF
	}

	if c.cr != nil {
		return c.cr.ReadByte()
	} else {
		return c.br.ReadByte()
	}
}

// UnreadByte complements ReadByte in implementing the io.ByteScanner interface.
func (c *ChunkReader) UnreadByte() error {
	var size int64
	if c.cr != nil {
		size = c.cr.Size()
	} else {
		size = c.br.Size()
	}

	var l int
	if c.cr != nil {
		l = c.cr.Len()
	} else {
		l = c.br.Len()
	}

	if size-int64(l) <= c.base {
		return errors.New("bytes.ChunkReader.UnreadByte: at beginning of slice")
	}

	if c.cr != nil {
		return c.cr.UnreadByte()
	} else {
		return c.br.UnreadByte()
	}
}

// ReadRune implements the io.RuneReader interface.
func (c *ChunkReader) ReadRune() (ch rune, size int, err error) {
	if c.Len() <= 0 {
		return 0, 0, io.EOF
	}

	if c.cr != nil {
		ch, size, err = c.cr.ReadRune()
	} else {
		ch, size, err = c.br.ReadRune()
	}

	if err != nil {
		return
	}
	if c.Len() < 0 { // the reading exceeded c.size
		//nolint:errcheck
		if c.cr != nil {
			c.cr.UnreadRune()
		} else {
			c.br.UnreadRune()
		}
		return 0, 0, io.EOF
	}

	return
}

// UnreadRune complements ReadRune in implementing the io.RuneScanner interface.
func (c *ChunkReader) UnreadRune() error {
	var size int64
	if c.cr != nil {
		size = c.cr.Size()
	} else {
		size = c.br.Size()
	}

	var l int
	if c.cr != nil {
		l = c.cr.Len()
	} else {
		l = c.br.Len()
	}

	if size-int64(l) <= c.base {
		return errors.New("bytes.ChunkReader.UnreadRune: at beginning of slice")
	}

	if c.cr != nil {
		return c.cr.UnreadRune()
	} else {
		return c.br.UnreadRune()
	}
}

// Seek implements the io.Seeker interface.
func (c *ChunkReader) Seek(offset int64, whence int) (int64, error) {
	var abs, rebasedOffset int64

	switch whence {
	case io.SeekStart:
		abs = offset
		rebasedOffset = c.base + abs
	case io.SeekCurrent:
		current := c.size - int64(c.Len())
		abs = current + offset
		rebasedOffset = offset
	case io.SeekEnd:
		abs = c.size + offset
		if c.cr != nil {
			rebasedOffset = c.base + c.size - c.cr.Size() + offset
		} else {
			rebasedOffset = c.base + c.size - c.br.Size() + offset
		}
	default:
		return 0, errors.New("bytes.ChunkReader.Seek: invalid whence")
	}

	if abs < 0 || abs > c.size {
		return 0, errors.New(strings.Join([]string{"bytes.ChunkReader.Seek: position", strconv.Itoa(int(abs)), "is out of bounds: [0,", strconv.Itoa(int(c.size)), "]"}, " "))
	}

	var err error
	if c.cr != nil {
		_, err = c.cr.Seek(rebasedOffset, whence)
	} else {
		_, err = c.br.Seek(rebasedOffset, whence)
	}
	if err != nil {
		return 0, err
	}

	return abs, nil
}

// WriteTo implements the io.WriterTo interface.
func (c *ChunkReader) WriteTo(w io.Writer) (n int64, err error) {
	if c.Len() <= 0 {
		return 0, nil
	}

	b := pools.GetByteSlice(c.Len(), c.Len())
	defer pools.ReleaseByteSlice(b)

	if c.cr != nil {
		_, err = c.cr.Read(b)
	} else {
		_, err = c.br.Read(b)
	}
	if err != nil {
		return 0, err
	}

	m, err := w.Write(b)
	n = int64(m)
	if m != len(b) && err == nil {
		err = io.ErrShortWrite
	}

	return
}
