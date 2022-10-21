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

package sdk

const (
	O_RDONLY int = 0 // open the file read-only.
	O_WRONLY int = 1 // open the file write-only.
	O_RDWR   int = 2 // open the file read-write.
	// The remaining values may be or'ed in to control behavior.
	O_APPEND int = 8    // append data to the file when writing.
	O_CREATE int = 512  // create a new file if none exists.
	O_EXCL   int = 2048 // used with O_CREATE, file must not exist.
	O_SYNC   int = 128  // open for synchronous I/O.
	O_TRUNC  int = 1024 // truncate regular writable file when opened.
)
