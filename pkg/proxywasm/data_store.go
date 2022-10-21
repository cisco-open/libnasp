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

package proxywasm

import (
	"io"
	"io/fs"
	"net/http"

	"emperror.dev/errors"

	"wwwin-github.cisco.com/eti/nasp/pkg/proxywasm/api"
)

type dataSource struct {
	fs       fs.FS
	filename string
	bytes    []byte
}

type urlDataSource struct {
	bytes []byte
	url   string
}

func NewURLDataSource(url string) api.DataSource {
	return &urlDataSource{
		url: url,
	}
}

func (s *urlDataSource) Get() ([]byte, error) {
	if s.bytes != nil {
		return s.bytes, nil
	}

	resp, err := http.DefaultClient.Get(s.url) //nolint:noctx
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "could not parse url", "url", s.url)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WrapIf(err, "could not read response body")
	}

	s.bytes = body

	return s.bytes, nil
}

func NewFileDataSource(fs fs.FS, filename string) api.DataSource {
	return &dataSource{
		fs:       fs,
		filename: filename,
	}
}

func NewBytesDataSource(bytes []byte) api.DataSource {
	return &dataSource{
		bytes: bytes,
	}
}

func (s *dataSource) Get() ([]byte, error) {
	if s.bytes != nil {
		return s.bytes, nil
	}

	f, err := s.fs.Open(s.filename)
	if err != nil {
		return nil, errors.WrapIf(err, "could not open file")
	}

	if s.filename != "" {
		bytes, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		s.bytes = bytes
	}

	return s.bytes, nil
}
