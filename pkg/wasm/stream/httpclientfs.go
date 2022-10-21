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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HTTPClientFS struct {
	FDProvider FDProvider
}

type URL struct {
	req  *http.Request
	resp *http.Response
	buf  *bytes.Buffer

	fd int
}

func (fs *HTTPClientFS) Open(name string, flag int, perm uint32) (Stream, error) {
	req, err := http.NewRequest(http.MethodGet, name, nil)
	if err != nil {
		return nil, err
	}

	return &URL{
		req: req,
		fd:  int(fs.FDProvider.NextFD()),
	}, nil
}

func (h *HTTPClientFS) Close() error {
	return nil
}

func (f *URL) Fd() int {
	return f.fd
}

func (f *URL) Write(p []byte) (n int, err error) {
	pieces := strings.SplitN(string(p), ":", 2)
	if pieces[0] == "set-header" {
		fmt.Printf("set-header: '%s'\n", p)
		header := strings.SplitN(pieces[1], ":", 2)
		f.req.Header.Add(header[0], header[1])
	}

	return len(p), nil
}

func (f *URL) Read(p []byte) (int, error) {
	if f.resp == nil {
		buf := new(bytes.Buffer)
		resp, err := http.DefaultClient.Do(f.req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		f.resp = resp

		buf.WriteString(fmt.Sprintf("HTTP/%d.%d %03d %s\r\n", f.resp.ProtoMajor, f.resp.ProtoMinor, f.resp.StatusCode, http.StatusText(f.resp.StatusCode)))
		_ = f.resp.Header.Write(buf)
		buf.WriteString("\n")

		b, _ := io.ReadAll(f.resp.Body)
		buf.Write(b)
		f.buf = buf
	}

	return f.buf.Read(p)
}

func (f *URL) Close() error {
	return nil
}
