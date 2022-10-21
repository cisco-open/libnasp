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
	"fmt"
	"io/ioutil"
	"os"
)

type HTTPClient struct {
	Headers map[string]string
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		Headers: make(map[string]string),
	}
}

func (c *HTTPClient) AddHeader(k, v string) *HTTPClient {
	c.Headers[k] = v

	return c
}

func (c *HTTPClient) AddHeaders(headers map[string]string) {
	for k, v := range headers {
		c.AddHeader(k, v)
	}
}

func (c *HTTPClient) Get(url string) ([]byte, error) {
	stream, err := OpenStream(os.Args[1])
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	for k, v := range c.Headers {
		_, err = stream.WriteString(fmt.Sprintf("set-header:%s:%s", k, v))
		if err != nil {
			return nil, err
		}
	}

	resp, err := ioutil.ReadAll(stream)
	if err != nil {

		return nil, err
	}

	return resp, err
}
