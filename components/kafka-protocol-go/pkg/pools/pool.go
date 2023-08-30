//  Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        https://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package pools

import (
	"bytes"
	"sync"
)

var (
	// byteSlicePool is a list of pooled byte slices
	// [0] -> byte slices of 4k max size
	// [1] -> byte slices of 8k max size
	// [2] -> byte slices of 16k max size
	// [3] -> byte slices of 32k max size
	// [4] -> byte slices of 64k max size
	// [5] -> byte slices of 128k max size
	// [6] -> byte slices which are above 128k in size
	byteSlicePool = make([]sync.Pool, 7)
)

func GetByteSlice(length, capacity int) []byte {
	return getSlice[byte](byteSlicePool, length, capacity)
}

func ReleaseByteSlice(b []byte) {
	releaseSlice(byteSlicePool, b)
}

func getBucketIndex(c int) int {
	switch {
	case c <= 4*1024:
		return 0
	case c <= 8*1024:
		return 1
	case c <= 16*1024:
		return 2
	case c <= 32*1024:
		return 3
	case c <= 64*1024:
		return 4
	case c <= 128*1024:
		return 5
	default:
		return 6
	}
}

func getSlice[T byte | bool | float64 | int8 | int16 | uint16 | int32 | int64](pools []sync.Pool, length, capacity int) []T {
	bucketIdx := getBucketIndex(capacity)
	for {
		b := pools[bucketIdx].Get()
		if b == nil {
			c := capacity
			if bucketIdx <= 2 {
				c = (1 << bucketIdx) * 4 * 1024 // in case the requested capacity is below 16k round the allocated capacity up to nearest bucket size
			}
			return make([]T, length, c)
		}

		slice, ok := b.([]T)
		if !ok {
			panic("pooled data has wrong type")
		}
		if cap(slice) < capacity {
			continue
		}

		return slice[:length]
	}
}

func releaseSlice[T byte | bool | float64 | int8 | int16 | uint16 | int32 | int64](pools []sync.Pool, b []T) {
	bucketIdx := getBucketIndex(cap(b))
	//nolint:staticcheck
	pools[bucketIdx].Put(b[:0])
}

var bytesReaderPool = sync.Pool{
	New: func() any {
		return new(bytes.Reader)
	}}

func GetBytesReader(p []byte) *bytes.Reader {
	r, _ := bytesReaderPool.Get().(*bytes.Reader)
	r.Reset(p)

	return r
}

func ReleaseBytesReader(r *bytes.Reader) {
	r.Reset(nil)
	bytesReaderPool.Put(r)
}
