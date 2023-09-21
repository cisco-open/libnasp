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

package varint

import (
	"bytes"
	"encoding/binary"
	"math/bits"
	"strconv"
	"strings"

	"emperror.dev/errors"

	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/serialization"
)

func Uint32Size(value uint32) int {
	if value < (1 << 7) {
		return 1
	}

	if value < (1 << 14) {
		return 2
	}

	if value < (1 << 21) {
		return 3
	}

	if value < (1 << 28) {
		return 4
	}

	return 5
}

func Int32Size(value int32) int {
	ux := (value << 1) ^ (value >> 31)

	return Uint32Size(uint32(ux))
}

func Uint64Size(value uint64) int {
	if value < (1 << 7) {
		return 1
	}

	if value < (1 << 14) {
		return 2
	}

	if value < (1 << 21) {
		return 3
	}

	if value < (1 << 28) {
		return 4
	}

	if value < (1 << 35) {
		return 5
	}

	if value < (1 << 42) {
		return 6
	}

	if value < (1 << 49) {
		return 7
	}

	if value < (1 << 56) {
		return 9
	}

	if value < (1 << 63) {
		return 9
	}

	return 10
}

func Int64Size(value int64) int {
	ux := (value << 1) ^ (value >> 63)

	return Uint64Size(uint64(ux))
}

// PutInt32 writes an integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value requires more than 5 bytes
func PutInt32(buf []byte, data int32) (int, error) {
	if Int32Size(data) > binary.MaxVarintLen32 {
		return -1, errors.New("varint is too long, the most significant bit in the 5th byte is set")
	}

	length := binary.PutVarint(buf, int64(data))

	return length, nil
}

// PutUint32 writes an integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value requires more than 5 bytes
func PutUint32(buf []byte, data uint32) (int, error) {
	if Uint32Size(data) > binary.MaxVarintLen32 {
		return -1, errors.New("varint is too long, the most significant bit in the 5th byte is set")
	}

	length := binary.PutUvarint(buf, uint64(data))

	return length, nil
}

// AppendInt32 appends an integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value requires more than 5 bytes
func AppendInt32(buf []byte, data int32) ([]byte, error) {
	if Int32Size(data) > binary.MaxVarintLen32 {
		return nil, errors.New("varint is too long, the most significant bit in the 5th byte is set")
	}

	return binary.AppendVarint(buf, int64(data)), nil
}

// WriteUint32 writes an unsigned integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value requires more than 5 bytes
func WriteUint32(w *typesbytes.SliceWriter, data uint32) error {
	if Uint32Size(data) > binary.MaxVarintLen32 {
		return errors.New("varint is too long, the most significant bit in the 5th byte is set")
	}
	var val [5]byte
	length := binary.PutUvarint(val[:], uint64(data))

	n, err := w.Write(val[:length])
	if err != nil {
		return errors.WrapIf(err, "couldn't write varint to byte buffer")
	}
	if n != length {
		errMsg := strings.Join([]string{"expected to write", strconv.Itoa(length), "bytes, but wrote", strconv.Itoa(n), "bytes"}, " ")
		return errors.New(errMsg)
	}

	return nil
}

// WriteInt32 writes an integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value requires more than 5 bytes
func WriteInt32(w *serialization.Writer, data int32) error {
	if Int32Size(data) > binary.MaxVarintLen32 {
		return errors.New("varint is too long, the most significant bit in the 5th byte is set")
	}

	var buf [5]byte
	length := binary.PutVarint(buf[:], int64(data))

	n, err := w.Write(buf[:length])
	if err != nil {
		return errors.WrapIf(err, "couldn't write serialized varint")
	}
	if n != length {
		return errors.New(strings.Join([]string{"expected to write", strconv.Itoa(length), "bytes of serialized varint, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}

// ReadUint32 reads an unsigned integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value does not terminate after 5 bytes have been read
func ReadUint32(r *bytes.Reader) (uint32, error) {
	val, err := binary.ReadUvarint(r)
	if err != nil {
		return 0, err
	}

	if bits.Len64(val) > 32 {
		return 0, errors.New(strings.Join([]string{"unsigned varint is too long, the most significant bit in the 5th byte is set, converted value:", strconv.Itoa(int(val))}, " "))
	}
	return uint32(val), nil
}

// ReadUint32_ reads an unsigned integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value does not terminate after 5 bytes have been read
func ReadUint32_(r *serialization.Reader) (uint32, error) {
	val, err := binary.ReadUvarint(r)
	if err != nil {
		return 0, err
	}

	if bits.Len64(val) > 32 {
		return 0, errors.New(strings.Join([]string{"unsigned varint is too long, the most significant bit in the 5th byte is set, converted value:", strconv.Itoa(int(val))}, " "))
	}
	return uint32(val), nil
}

// ReadInt32 reads an integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value does not terminate after 5 bytes have been read
func ReadInt32(r *bytes.Reader) (int32, error) {
	uval, err := ReadUint32(r)
	if err != nil {
		return 0, err
	}

	if bits.Len64(uint64(uval)) > 32 {
		return 0, errors.New(strings.Join([]string{"unsigned varint is too long, the most significant bit in the 5th byte is set, converted value:", strconv.Itoa(int(uval))}, " "))
	}

	val := int64(uval >> 1)
	if uval&1 != 0 {
		val = ^val
	}

	return int32(val), nil
}

// ReadInt32_ reads an integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value does not terminate after 5 bytes have been read
func ReadInt32_(r *serialization.Reader) (int32, error) {
	uval, err := ReadUint32_(r)
	if err != nil {
		return 0, err
	}

	if bits.Len64(uint64(uval)) > 32 {
		return 0, errors.New(strings.Join([]string{"unsigned varint is too long, the most significant bit in the 5th byte is set, converted value:", strconv.Itoa(int(uval))}, " "))
	}

	val := int64(uval >> 1)
	if uval&1 != 0 {
		val = ^val
	}

	return int32(val), nil
}

// PutInt64 writes an integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value requires more than 10 bytes
func PutInt64(buf []byte, data int64) (int, error) {
	if Int64Size(data) > binary.MaxVarintLen64 {
		return -1, errors.New("varint is too long, the most significant bit in the 10th byte is set")
	}

	length := binary.PutVarint(buf, data)
	return length, nil
}

// AppendInt64 appends an integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value requires more than 10 bytes
func AppendInt64(buf []byte, data int64) ([]byte, error) {
	if Int64Size(data) > binary.MaxVarintLen64 {
		return nil, errors.New("varint is too long, the most significant bit in the 10th byte is set")
	}

	return binary.AppendVarint(buf, data), nil
}

// ReadInt64 reads an 64-bit integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value does not terminate after 10 bytes have been read
func ReadInt64(r *bytes.Reader) (int64, error) {
	val, err := binary.ReadVarint(r)
	if err != nil {
		return 0, err
	}

	return val, nil
}

// ReadInt64_ reads an 64-bit integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value does not terminate after 10 bytes have been read
func ReadInt64_(r *serialization.Reader) (int64, error) {
	val, err := binary.ReadVarint(r)
	if err != nil {
		return 0, err
	}

	return val, nil
}

// WriteInt64 writes an integer stored in variable-length format using unsigned decoding from
// http://code.google.com/apis/protocolbuffers/docs/encoding.html
// Returns an error if the variable-length value requires more than 10 bytes
func WriteInt64(w *serialization.Writer, data int64) error {
	var buf [10]byte
	length := binary.PutVarint(buf[:], data)

	n, err := w.Write(buf[:length])
	if err != nil {
		return errors.WrapIf(err, "couldn't write serialized varint")
	}
	if n != length {
		return errors.New(strings.Join([]string{"expected to write", strconv.Itoa(length), "bytes of serialized varint, but wrote", strconv.Itoa(n), "bytes"}, " "))
	}

	return nil
}
