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
	"strconv"
	"strings"

	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/serialization"

	"emperror.dev/errors"
)

type Readers interface {
	*bytes.Reader | *typesbytes.ChunkReader | *serialization.Reader
	Read([]byte) (int, error)
}

func WriteInt32(w *typesbytes.SliceWriter, data int32) error {
	var buf [4]byte

	binary.BigEndian.PutUint32(buf[:], uint32(data))
	n, err := w.Write(buf[:])

	if err != nil {
		return errors.WrapIf(err, "couldn't write serialized int32")
	}
	if n != 4 {
		return errors.New(strings.Join([]string{"expected to write 4 bytes of serialized int32, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

func ReadInt32(r *bytes.Reader, out *int32) error {
	var buf [4]byte

	n, err := r.Read(buf[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.WrapIf(err, strings.Join([]string{"expected 4 bytes, but got", strconv.Itoa(n)}, " "))
		}
		return err
	}

	if n != 4 {
		return errors.New(strings.Join([]string{"expected 4 bytes, but got", strconv.Itoa(n)}, " "))
	}

	*out = int32(binary.BigEndian.Uint32(buf[:]))

	return nil
}

func ReadInt32_(r *typesbytes.ChunkReader, out *int32) error {
	var buf [4]byte

	n, err := r.Read(buf[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.WrapIf(err, strings.Join([]string{"expected 4 bytes, but got", strconv.Itoa(n)}, " "))
		}
		return err
	}

	if n != 4 {
		return errors.New(strings.Join([]string{"expected 4 bytes, but got", strconv.Itoa(n)}, " "))
	}

	*out = int32(binary.BigEndian.Uint32(buf[:]))

	return nil
}

func ReadInt16(r *bytes.Reader, out *int16) error {
	var buf [2]byte

	n, err := r.Read(buf[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.WrapIf(err, strings.Join([]string{"expected 2 bytes, but got", strconv.Itoa(n)}, " "))
		}
		return err
	}

	if n != 2 {
		return errors.New(strings.Join([]string{"expected 2 bytes, but got", strconv.Itoa(n)}, " "))
	}

	*out = int16(binary.BigEndian.Uint16(buf[:]))

	return nil
}

func ReadInt16_(r *typesbytes.ChunkReader, out *int16) error {
	var buf [2]byte

	n, err := r.Read(buf[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.WrapIf(err, strings.Join([]string{"expected 2 bytes, but got", strconv.Itoa(n)}, " "))
		}
		return err
	}

	if n != 2 {
		return errors.New(strings.Join([]string{"expected 2 bytes, but got", strconv.Itoa(n)}, " "))
	}

	*out = int16(binary.BigEndian.Uint16(buf[:]))

	return nil
}

func WriteInt16(w *typesbytes.SliceWriter, data int16) error {
	var buf [2]byte

	binary.BigEndian.PutUint16(buf[:], uint16(data))
	n, err := w.Write(buf[:])

	if err != nil {
		return errors.WrapIf(err, "couldn't write serialized int16")
	}
	if n != 2 {
		return errors.New(strings.Join([]string{"expected to write 2 bytes of serialized int16, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

func ReadUint16(r *bytes.Reader, out *uint16) error {
	var buf [2]byte

	n, err := r.Read(buf[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.WrapIf(err, strings.Join([]string{"expected 2 bytes, but got", strconv.Itoa(n)}, " "))
		}
		return err
	}

	if n != 2 {
		return errors.New(strings.Join([]string{"expected 2 bytes, but got", strconv.Itoa(n)}, " "))
	}

	*out = binary.BigEndian.Uint16(buf[:])

	return nil
}

func WriteUint16(w *typesbytes.SliceWriter, data uint16) error {
	var buf [2]byte

	binary.BigEndian.PutUint16(buf[:], data)
	n, err := w.Write(buf[:])

	if err != nil {
		return errors.WrapIf(err, "couldn't write serialized uint16")
	}
	if n != 2 {
		return errors.New(strings.Join([]string{"expected to write 2 bytes of serialized uint16, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

func WriteInt64(w *typesbytes.SliceWriter, data int64) error {
	var buf [8]byte

	binary.BigEndian.PutUint64(buf[:], uint64(data))
	n, err := w.Write(buf[:])

	if err != nil {
		return errors.WrapIf(err, "couldn't write serialized int64")
	}
	if n != 8 {
		return errors.New(strings.Join([]string{"expected to write 8 bytes of serialized int64, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

func ReadInt64(r *bytes.Reader, out *int64) error {
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

	*out = int64(binary.BigEndian.Uint64(buf[:]))

	return nil
}

func ReadInt64_(r *typesbytes.ChunkReader, out *int64) error {
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

	*out = int64(binary.BigEndian.Uint64(buf[:]))

	return nil
}

func WriteInt8(w *typesbytes.SliceWriter, data int8) error {
	var buf [1]byte
	buf[0] = byte(data)

	n, err := w.Write(buf[:])

	if err != nil {
		return errors.WrapIf(err, "couldn't write serialized int8")
	}
	if n != 1 {
		return errors.New(strings.Join([]string{"expected to write 1 byte of serialized int8, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

func ReadInt8(r *bytes.Reader, out *int8) error {
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

	*out = int8(buf[0])

	return nil
}

func ReadInt8_(r *serialization.Reader, out *int8) error {
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

	*out = int8(buf[0])

	return nil
}

func ReadInt8__(r *typesbytes.ChunkReader, out *int8) error {
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

	*out = int8(buf[0])

	return nil
}
