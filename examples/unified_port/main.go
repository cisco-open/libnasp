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

// Create a certificate first with:
// openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 365 -nodes

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	unifiedtls "wwwin-github.cisco.com/eti/nasp/pkg/tls"
)

var (
	addr = "127.0.0.1:8080"
)

func hello(w http.ResponseWriter, r *http.Request) {
	//	r.TLS.Version
	fmt.Fprintln(w, "Hello world!")
}

func main() {

	var err error

	config := &tls.Config{
		Certificates: make([]tls.Certificate, 1),
		NextProtos:   []string{"h2"},
	}

	config.Certificates[0], err = tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", hello)

	err = http.Serve(
		unifiedtls.NewUnifiedListener(ln, config, unifiedtls.TLSModePermissive),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil {
				u := url.URL{
					Scheme:   "https",
					Opaque:   r.URL.Opaque,
					User:     r.URL.User,
					Host:     addr,
					Path:     r.URL.Path,
					RawQuery: r.URL.RawQuery,
					Fragment: r.URL.Fragment,
				}
				http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
			} else {
				http.DefaultServeMux.ServeHTTP(w, r)
			}
		}))
	if err != nil {
		log.Fatal(err)
	}
}
