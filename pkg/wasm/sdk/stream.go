// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

//go:build wasi
// +build wasi

package sdk

import (
	"errors"
	"io"
	"unsafe"
)

type Stream struct {
	Fd uint32
}

func OpenStream(path string, params ...uint32) (*Stream, error) {
	flag := uint32(0)
	perm := uint32(0)

	if len(params) > 0 {
		flag = params[0]
	}
	if len(params) > 1 {
		perm = params[1]
	}

	ptr, size := stringToPtr(path)
	fd := uint32(0)
	rc := _stream_open(ptr, size, uint32Pointer(&fd), flag, perm)
	if rc > 0 {
		return nil, errors.New(ErrnoString(rc))
	}

	return &Stream{
		Fd: fd,
	}, nil
}

func (s *Stream) Read(p []byte) (n int, err error) {
	return stream_read(s.Fd, p)
}

func (s *Stream) Write(p []byte) (n int, err error) {
	return stream_write(s.Fd, p)
}

func (s *Stream) WriteString(str string) (n int, err error) {
	return s.Write([]byte(str))
}

func (s *Stream) Close() error {
	return stream_close(s.Fd)
}

func (s *Stream) Seek(offset int64, whence int) (int64, error) {
	newOffset, err := stream_seek(s.Fd, offset, uint32(whence))

	return int64(newOffset), err
}

type IOVS struct {
	buf uintptr
	len uint32
}

func stream_read(fd uint32, p []byte) (n int, err error) {
	resultSize := uint32(0)

	iovs := IOVS{
		buf: uintptr(unsafe.Pointer(&p[0])),
		len: uint32(len(p)),
	}

	if rc := _stream_read(fd, uint32(uintptr(unsafe.Pointer(&iovs))), 1, uint32Pointer(&resultSize)); rc != 0 {
		return 0, errors.New(ErrnoString(rc))
	}

	if resultSize == 0 {
		return 0, io.EOF
	}

	return int(resultSize), nil
}

func stream_write(fd uint32, p []byte) (n int, err error) {
	resultSize := uint32(0)

	iovs := IOVS{
		buf: uintptr(unsafe.Pointer(&p[0])),
		len: uint32(len(p)),
	}

	if rc := _stream_write(fd, uint32(uintptr(unsafe.Pointer(&iovs))), 1, uint32Pointer(&resultSize)); rc != 0 {
		return 0, errors.New(ErrnoString(rc))
	}

	return int(resultSize), nil
}

func stream_close(fd uint32) error {
	if rc := _stream_close(fd); rc != 0 {
		return errors.New(ErrnoString(rc))
	}

	return nil
}

func stream_seek(fd uint32, offset int64, whence uint32) (uint32, error) {
	resultNewoffset := uint32(0)

	if rc := _stream_seek(fd, offset, whence, uint32Pointer(&resultNewoffset)); rc != 0 {
		return 0, errors.New(ErrnoString(rc))
	}

	return resultNewoffset, nil
}

//go:wasm-module dibbler_stream
//export open
func _stream_open(ptr, size, fd, flag, perm uint32) uint32

//go:wasm-module dibbler_stream
//export read
func _stream_read(fd, iovs, iovsCount, resultSize uint32) uint32

//go:wasm-module dibbler_stream
//export write
func _stream_write(fd, iovs, iovsCount, resultSize uint32) uint32

//go:wasm-module dibbler_stream
//export close
func _stream_close(fd uint32) uint32

//go:wasm-module dibbler_stream
//export seek
func _stream_seek(fd uint32, offset int64, whence uint32, resultNewoffset uint32) uint32

// stringToPtr returns a pointer and size pair for the given string in a way
// compatible with WebAssembly numeric types.
func stringToPtr(s string) (uint32, uint32) {
	buf := []byte(s)
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return uint32(unsafePtr), uint32(len(buf))
}

// uint32Pointer returns a pointer to an uint32 value
func uint32Pointer(n *uint32) uint32 {
	return uint32(uintptr(unsafe.Pointer(n)))
}
