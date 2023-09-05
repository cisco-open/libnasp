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

package fields

import (
	"bytes"
	"encoding/binary"
	"io"
	"strconv"
	"strings"

	typesbytes "github.com/cisco-open/nasp/components/kafka-protocol-go/pkg/protocol/types/bytes"

	"emperror.dev/errors"
)

// RecordBatches manipulates RecordBatch collections
type RecordBatches struct {
	items []RecordBatch
	isNil bool
}

func (b *RecordBatches) Items() []RecordBatch {
	return b.items
}

func (b *RecordBatches) SetItems(items []RecordBatch) {
	b.items = items
	b.isNil = false
}

func (b *RecordBatches) IsNil() bool {
	return b.isNil
}

func (b *RecordBatches) Release() {
	for i := range b.items {
		b.items[i].Release()
	}

	b.items = nil
}

func (b *RecordBatches) Clear() {
	b.Release()
	b.isNil = true
}

func (b *RecordBatches) Complete() {
	b.isNil = false
}

func (b *RecordBatches) ClearAndComplete() {
	b.Release()
	b.isNil = false
}

func (b *RecordBatches) String() string {
	s, err := b.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (b *RecordBatches) Equal(that *RecordBatches) bool {
	if len(b.items) != len(that.items) {
		return false
	}

	for i := range b.items {
		if !b.items[i].Equal(&that.items[i]) {
			return false
		}
	}
	return true
}

func (b *RecordBatches) MarshalJSON() ([]byte, error) {
	if b == nil || b.isNil || b.items == nil {
		return []byte("null"), nil
	}

	s := make([][]byte, 0, len(b.items))

	for i := range b.items {
		j, err := b.items[i].MarshalJSON()
		if err != nil {
			return nil, err
		}
		s = append(s, j)
	}

	var arr bytes.Buffer
	if err := arr.WriteByte('['); err != nil {
		return nil, err
	}
	if _, err := arr.Write(bytes.Join(s, []byte(", "))); err != nil {
		return nil, err
	}
	if err := arr.WriteByte(']'); err != nil {
		return nil, err
	}

	return arr.Bytes(), nil
}

func (b *RecordBatches) Read(r *typesbytes.ChunkReader) error {
	b.Clear()

	for i := 0; ; i++ {
		batchSize, err := b.nextBatchSize(r)
		if err != nil {
			return errors.WrapIf(err, strings.Join([]string{"couldn't determine the record batch size in bytes for record batch item:", strconv.Itoa(i)}, " "))
		}
		if batchSize == 0 || r.Len() < batchSize {
			// if next batch's size is zero or r has fewer bytes remaining the batchSize
			// than we already processed the last complete record batch
			if r.Len() < batchSize {
				_, err = r.Seek(0, io.SeekEnd)
				if err != nil {
					return err
				}
			}
			b.Complete()
			return nil
		}

		var rb RecordBatch
		scr := typesbytes.NewSubChunkReader(r, int64(batchSize))
		err = rb.Read(&scr)
		if err != nil {
			return errors.WrapIf(err, strings.Join([]string{"couldn't parse record batch item", strconv.Itoa(i)}, " "))
		}

		b.items = append(b.items, rb)
	}
}

func (b *RecordBatches) Write(buf *typesbytes.SliceWriter) error {
	if b.IsNil() {
		return nil
	}

	for i := range b.items {
		err := b.items[i].Write(buf)
		if err != nil {
			return errors.WrapIf(err, strings.Join([]string{"couldn't serialize record batch item:", strconv.Itoa(i)}, " "))
		}
	}

	return nil
}

// SizeInBytes returns the size of this record batch in bytes when it's serialized
func (b *RecordBatches) SizeInBytes() int {
	if b.IsNil() {
		return 0
	}

	length := 0
	for i := range b.items {
		length += b.items[i].SizeInBytes()
	}

	return length
}

func (b *RecordBatches) nextBatchSize(r *typesbytes.ChunkReader) (int, error) {
	if r.Len() < 12 {
		// remaining bytes in buffer is less than the batch overhead size
		// which consists of the baseOffset and batchLength fields: https://kafka.apache.org/documentation/#recordbatch
		return 0, nil
	}

	var batchLengthBytes [4]byte
	n, err := r.ReadAt(batchLengthBytes[:], r.Size()-int64(r.Len())+8)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't read record \"batchLength\"")
	}
	if n < len(batchLengthBytes) {
		return 0, errors.New(strings.Join([]string{"expected to read 4 bytes to get record \"batchLength\" but got only", strconv.Itoa(n), "bytes"}, " "))
	}

	batchLength := int32(binary.BigEndian.Uint32(batchLengthBytes[:]))

	if r.Len() < 17 {
		// remaining bytes in buffer is less than the batch header size up to the magic byte (inclusive)
		return 0, nil
	}

	var magicByte [1]byte
	n, err = r.ReadAt(magicByte[:], r.Size()-int64(r.Len())+16)
	if err != nil {
		return 0, errors.WrapIf(err, "couldn't read record batch \"magic\" byte")
	}
	if n < len(magicByte) {
		return 0, errors.New(strings.Join([]string{"expected to read 1 byte to get record batch \"magic\" byte but got only", strconv.Itoa(n), "bytes"}, " "))
	}
	magic := int8(magicByte[0])
	if magic < 0 || magic > 2 {
		return 0, errors.New(strings.Join([]string{"invalid magic byte found in record:", strconv.Itoa(int(magic))}, " "))
	}

	logOverheadBytesSize := 12 // baseOffset and batchLength fields size in bytes
	return int(batchLength) + logOverheadBytesSize, nil
}
