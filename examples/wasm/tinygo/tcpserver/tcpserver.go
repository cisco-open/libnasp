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
	"io"
	"os"

	"wwwin-github.cisco.com/eti/nasp/pkg/wasm/sdk"
)

func main() {
	stream, err := sdk.OpenStream("socket-accept:///1")
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	defer stream.Close()

	for {
		buf := make([]byte, 1024)
		_, err := stream.Read(buf)
		if err == io.EOF {
			break
		}
		// echo
		stream.Write(buf)
	}
}
