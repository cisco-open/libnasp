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
	"io/fs"
	"net/url"
	"os"
)

type LocalFSHandler struct{}

func (h *LocalFSHandler) Close() error {
	return nil
}

type File struct {
	*os.File
}

func (f *File) Fd() int {
	return int(f.File.Fd())
}

func (h *LocalFSHandler) Open(path string, flag int, perm uint32) (Stream, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	if u.Path == "" {
		return nil, os.ErrNotExist
	}

	path = u.Path[1:]

	f, err := os.OpenFile(path, flag|os.O_RDWR, fs.FileMode(perm))
	if err != nil {
		return nil, err
	}

	return &File{
		File: f,
	}, nil
}
