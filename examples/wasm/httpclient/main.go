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
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"

	"wwwin-github.cisco.com/eti/nasp/pkg/wasm/stream"
)

var contextIDGenerator int32
var rootContextID int32
var once sync.Once

func main() {
	ctx := context.Background()

	wasm, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Panicln(err)
	}

	r := wazero.NewRuntimeWithConfig(context.Background(), wazero.NewRuntimeConfig().WithWasmCore2())
	defer r.Close(ctx)

	if _, err := stream.New(r).Instantiate(ctx, r); err != nil {
		log.Panicln(err)
	}

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		log.Panicln(err)
	}

	// Compile the WebAssembly module using the default configuration.
	code, err := r.CompileModule(ctx, wasm, wazero.NewCompileConfig())
	if err != nil {
		log.Panicln(err)
	}

	config := wazero.NewModuleConfig().
		WithStdout(os.Stdout).WithStderr(os.Stderr).WithFS(os.DirFS("/")).
		WithArgs(append([]string{"wasi"}, os.Args[2:]...)...).
		WithName(time.Now().String())

	if _, err := r.InstantiateModule(ctx, code, config); err != nil {
		if exitErr, ok := err.(*sys.ExitError); ok && exitErr.ExitCode() != 0 {
			fmt.Fprintf(os.Stderr, "exit_code: %d\n", exitErr.ExitCode())
		} else if !ok {
			log.Panicln(err)
		}
	}
}
