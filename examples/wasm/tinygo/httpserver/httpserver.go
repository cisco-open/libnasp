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

package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/julienschmidt/httprouter"

	"wwwin-github.cisco.com/eti/nasp/pkg/wasm/sdk"
)

func main() {
	stream, err := sdk.OpenStream("socket-accept:///1")
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	defer stream.Close()

	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	router := httprouter.New()
	router.GET("/hello", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Add("a", "b")
		fmt.Fprint(w, "Welcome!\n")
	})

	router.GET("/error", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	rr.Result().Write(stream)
}
