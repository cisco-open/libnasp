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

package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"sync/atomic"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"wwwin-github.cisco.com/eti/nasp/pkg/wasm/sdk"
)

const ModuleName = "dibbler_stream"

type Stream interface {
	io.ReadWriteCloser
	Fd() int
}

type stream struct {
	openedStreams map[uint32]Stream
	handlers      map[string]Handler
	lastFD        uint32

	r wazero.Runtime
}

type Handler interface {
	Open(name string, flag int, perm uint32) (Stream, error)
	io.Closer
}

type FDProvider interface {
	NextFD() uint32
}

func New(r wazero.Runtime) *stream {
	s := &stream{
		openedStreams: make(map[uint32]Stream),
		handlers:      make(map[string]Handler),
		lastFD:        3,
		r:             r,
	}

	s.handlers["file"] = &LocalFSHandler{}
	s.handlers["http"] = &HTTPClientFS{
		FDProvider: s,
	}

	return s
}

func (s *stream) AddHandler(scheme string, handler Handler) *stream {
	s.handlers[scheme] = handler

	return s
}

func (s *stream) Instantiate(ctx context.Context, ns wazero.Namespace) (api.Closer, error) {
	return s.r.NewModuleBuilder(ModuleName).ExportFunctions(map[string]interface{}{
		"open":  s.Open,
		"close": s.Close,
		"read":  s.Read,
		"write": s.Write,
		"seek":  s.Seek,
	}).Instantiate(ctx, ns)
}

func (s *stream) NextFD() uint32 {
	if s.lastFD == math.MaxUint32 {
		return 0
	}
	return atomic.AddUint32(&s.lastFD, 1)
}

func (s *stream) Open(ctx context.Context, m api.Module, pathPtr, pathLen, resultFd, flag, perm uint32) sdk.Errno {
	p, ok := m.Memory().Read(ctx, pathPtr, pathLen)
	if !ok {
		return sdk.ErrnoFault
	}
	path := string(p)

	u, err := url.Parse(path)
	if err != nil {
		return sdk.ErrnoInval
	}

	if u.Path == "" {
		return sdk.ErrnoInval
	}

	var handler Handler
	if handler, ok = s.handlers[u.Scheme]; !ok {
		return sdk.ErrnoInval
	}

	f, err := handler.Open(path, int(flag), perm)
	if err != nil {
		return sdk.ErrnoNoent
	}

	if !m.Memory().WriteUint32Le(ctx, resultFd, uint32(f.Fd())) {
		return sdk.ErrnoFault
	}

	s.openedStreams[uint32(f.Fd())] = f

	return sdk.ErrnoSuccess
}

func (s *stream) Close(ctx context.Context, mod api.Module, fd uint32) uint32 {
	var stream Stream
	var ok bool

	if stream, ok = s.openedStreams[fd]; !ok {
		return sdk.ErrnoSuccess
	}

	if closer, ok := stream.(io.Closer); ok {
		err := closer.Close()
		if err != nil {
			return sdk.ErrnoFault
		}
	}

	delete(s.openedStreams, fd)

	return sdk.ErrnoSuccess
}

func (s *stream) Read(ctx context.Context, mod api.Module, fd, iovs, iovsCount, resultSize uint32) sdk.Errno {
	var nread uint32

	var stream Stream
	var ok bool

	if stream, ok = s.openedStreams[fd]; !ok {
		return sdk.ErrnoInval
	}

	for i := uint32(0); i < iovsCount; i++ {
		iovPtr := iovs + i*8
		offset, ok := mod.Memory().ReadUint32Le(ctx, iovPtr)
		if !ok {
			return sdk.ErrnoFault
		}
		l, ok := mod.Memory().ReadUint32Le(ctx, iovPtr+4)
		if !ok {
			return sdk.ErrnoFault
		}
		b, ok := mod.Memory().Read(ctx, offset, l)
		if !ok {
			return sdk.ErrnoFault
		}
		n, err := stream.Read(b)
		nread += uint32(n)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return sdk.ErrnoIo
		}
	}

	if !mod.Memory().WriteUint32Le(ctx, resultSize, nread) {
		return sdk.ErrnoFault
	}

	return sdk.ErrnoSuccess
}

func (s *stream) Write(ctx context.Context, mod api.Module, fd, iovs, iovsCount, resultSize uint32) sdk.Errno {
	var nwritten uint32

	var stream Stream
	var ok bool

	if stream, ok = s.openedStreams[fd]; !ok {
		return sdk.ErrnoInval
	}

	for i := uint32(0); i < iovsCount; i++ {
		iovPtr := iovs + i*8
		offset, ok := mod.Memory().ReadUint32Le(ctx, iovPtr)
		if !ok {
			return sdk.ErrnoFault
		}
		l, ok := mod.Memory().ReadUint32Le(ctx, iovPtr+4)
		if !ok {
			return sdk.ErrnoFault
		}
		b, ok := mod.Memory().Read(ctx, offset, l)
		if !ok {
			return sdk.ErrnoFault
		}
		n, err := stream.Write(b)
		if err != nil {
			fmt.Println(err.Error())
			return sdk.ErrnoIo
		}
		nwritten += uint32(n)
	}

	if !mod.Memory().WriteUint32Le(ctx, resultSize, nwritten) {
		return sdk.ErrnoFault
	}

	return sdk.ErrnoSuccess
}

func (s *stream) Seek(ctx context.Context, mod api.Module, fd uint32, offset int64, whence uint32, resultNewoffset uint32) sdk.Errno {
	var f Stream
	var ok bool

	if f, ok = s.openedStreams[fd]; !ok {
		return sdk.ErrnoInval
	}

	var seeker io.Seeker
	if seeker, ok = f.(io.Seeker); !ok {
		return sdk.ErrnoBadf
	}

	if whence > io.SeekEnd {
		return sdk.ErrnoInval
	}

	newOffset, err := seeker.Seek(offset, int(whence))
	if err != nil {
		return sdk.ErrnoIo
	}

	if !mod.Memory().WriteUint32Le(ctx, resultNewoffset, uint32(newOffset)) {
		return sdk.ErrnoFault
	}

	return sdk.ErrnoSuccess
}
