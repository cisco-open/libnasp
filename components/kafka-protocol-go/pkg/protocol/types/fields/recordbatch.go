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
	"math"
	"strconv"
	"strings"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/serialization"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/crc32c"

	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/compression/gzip"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/compression/lz4"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/compression/snappy"
	"github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/compression/zstd"

	"emperror.dev/errors"

	typesbytes "github.com/cisco-open/libnasp/components/kafka-protocol-go/pkg/protocol/types/bytes"
)

// CompressionType represents the compression applied to a record  batch.
type CompressionType int8

const (
	None   CompressionType = 0
	Gzip   CompressionType = 1
	Snappy CompressionType = 2
	Lz4    CompressionType = 3
	Zstd   CompressionType = 4
)

// RecordBatch provides setter/getter methods for kafka record batch https://kafka.apache.org/documentation/#recordbatch
type RecordBatch struct {
	records              []BatchRecord
	baseOffset           int64
	producerId           int64
	maxTimestamp         int64
	baseTimestamp        int64
	lastOffsetDelta      int32
	crc                  uint32
	partitionLeaderEpoch int32
	baseSequence         int32
	batchLength          int32
	attributes           int16
	producerEpoch        int16
	magic                int8
	isComplete           bool
}

func (b *RecordBatch) BaseOffset() int64 {
	return b.baseOffset
}

func (b *RecordBatch) SetBaseOffset(baseOffset int64) {
	b.baseOffset = baseOffset
}

func (b *RecordBatch) PartitionLeaderEpoch() int32 {
	return b.partitionLeaderEpoch
}

func (b *RecordBatch) SetPartitionLeaderEpoch(partitionLeaderEpoch int32) {
	b.partitionLeaderEpoch = partitionLeaderEpoch
}

func (b *RecordBatch) Magic() int8 {
	return b.magic
}

func (b *RecordBatch) Attributes() int16 {
	return b.attributes
}

func (b *RecordBatch) SetAttributes(attributes int16) {
	b.attributes = attributes
}

func (b *RecordBatch) LastOffsetDelta() int32 {
	return b.lastOffsetDelta
}

func (b *RecordBatch) SetLastOffsetDelta(lastOffsetDelta int32) {
	b.lastOffsetDelta = lastOffsetDelta
}

func (b *RecordBatch) BaseTimestamp() int64 {
	return b.baseTimestamp
}

func (b *RecordBatch) SetBaseTimestamp(baseTimestamp int64) {
	b.baseTimestamp = baseTimestamp
}

func (b *RecordBatch) MaxTimestamp() int64 {
	return b.maxTimestamp
}

func (b *RecordBatch) SetMaxTimestamp(maxTimestamp int64) {
	b.maxTimestamp = maxTimestamp
}

func (b *RecordBatch) ProducerId() int64 {
	return b.producerId
}

func (b *RecordBatch) SetProducerId(producerId int64) {
	b.producerId = producerId
}

func (b *RecordBatch) ProducerEpoch() int16 {
	return b.producerEpoch
}

func (b *RecordBatch) SetProducerEpoch(producerEpoch int16) {
	b.producerEpoch = producerEpoch
}

func (b *RecordBatch) BaseSequence() int32 {
	return b.baseSequence
}

func (b *RecordBatch) SetBaseSequence(baseSequence int32) {
	b.baseSequence = baseSequence
}

func (b *RecordBatch) Records() []BatchRecord {
	return b.records
}

func (b *RecordBatch) SetRecords(records []BatchRecord) {
	b.records = records
}

func (b *RecordBatch) IsComplete() bool {
	return b.isComplete
}

func (b *RecordBatch) Clear() {
	b.records = nil
	b.isComplete = false
}

func (b *RecordBatch) Complete() {
	b.isComplete = true
}

func (b *RecordBatch) Equal(that *RecordBatch) bool {
	if !(b.baseOffset == that.baseOffset &&
		b.batchLength == that.batchLength &&
		b.partitionLeaderEpoch == that.partitionLeaderEpoch &&
		b.magic == that.magic &&
		b.crc == that.crc &&
		b.attributes == that.attributes &&
		b.lastOffsetDelta == that.lastOffsetDelta &&
		b.baseTimestamp == that.baseTimestamp &&
		b.maxTimestamp == that.maxTimestamp &&
		b.producerId == that.producerId &&
		b.producerEpoch == that.producerEpoch &&
		b.baseSequence == that.baseSequence) {
		return false
	}

	if len(b.records) != len(that.records) {
		return false
	}
	for i := range b.records {
		if !b.records[i].Equal(&that.records[i]) {
			return false
		}
	}

	return true
}

func (b *RecordBatch) String() string {
	s, err := b.MarshalJSON()
	if err != nil {
		return err.Error()
	}

	return string(s)
}

func (b *RecordBatch) MarshalJSON() ([]byte, error) {
	if b == nil || !b.isComplete {
		return []byte("null"), nil
	}

	var s [13][]byte
	s[0] = []byte("{\"baseOffset\": " + strconv.FormatInt(b.baseOffset, 10))
	s[1] = []byte("\"batchLength\": " + strconv.FormatInt(int64(b.batchLength), 10))
	s[2] = []byte("\"partitionLeaderEpoch\": " + strconv.FormatInt(int64(b.partitionLeaderEpoch), 10))
	s[3] = []byte("\"magic\": " + strconv.FormatInt(int64(b.magic), 10))
	s[4] = []byte("\"crc\": " + strconv.FormatInt(int64(b.crc), 10))
	s[5] = []byte("\"attributes\": " + strconv.FormatInt(int64(b.attributes), 10))
	s[6] = []byte("\"lastOffsetDelta\": " + strconv.FormatInt(int64(b.lastOffsetDelta), 10))
	s[7] = []byte("\"baseTimestamp\": " + strconv.FormatInt(b.baseTimestamp, 10))
	s[8] = []byte("\"maxTimestamp\": " + strconv.FormatInt(b.maxTimestamp, 10))
	s[9] = []byte("\"producerId\": " + strconv.FormatInt(b.producerId, 10))
	s[10] = []byte("\"producerEpoch\": " + strconv.FormatInt(int64(b.producerEpoch), 10))
	s[11] = []byte("\"baseSequence\": " + strconv.FormatInt(int64(b.baseSequence), 10))
	//nolint:nestif
	if b.records == nil {
		s[12] = []byte("\"records\": null}")
	} else {
		r := make([][]byte, 0, len(b.records))
		for i := range b.records {
			j, err := b.records[i].MarshalJSON()
			if err != nil {
				return nil, err
			}
			r = append(r, j)
		}

		var arr bytes.Buffer
		if _, err := arr.WriteString("\"records\": "); err != nil {
			return nil, err
		}
		if err := arr.WriteByte('['); err != nil {
			return nil, err
		}
		if _, err := arr.Write(bytes.Join(r, []byte(", "))); err != nil {
			return nil, err
		}
		if _, err := arr.WriteString("]}"); err != nil {
			return nil, err
		}

		s[12] = arr.Bytes()
	}

	return bytes.Join(s[:], []byte(", ")), nil
}

func (b *RecordBatch) Read(r *typesbytes.ChunkReader) error {
	var magicByte = make([]byte, 1)
	_, err := r.ReadAt(magicByte, r.Size()-int64(r.Len())+16)
	if err != nil {
		return errors.WrapIf(err, "couldn't read \"magic\" field")
	}

	magic := magicByte[0]
	switch magic {
	case 0:
		return errors.New("message sets version 0 (pre Kafka 0.10) are not supported")
	case 1:
		return errors.New("message sets version 1 (Kafka 0.10) are not supported")
	case 2:
		// magic = 2 corresponds to Kafka versions 0.11 and above. Kafka 0.11 replaced message sets with record batches
		return b.read(r)
	default:
		return errors.New(strings.Join([]string{"unrecognized 'magic' number", strconv.Itoa(int(magic)), "in records field"}, " "))
	}
}

func (b *RecordBatch) Write(buf *typesbytes.SliceWriter) error {
	switch b.magic {
	case 0:
		return errors.New("message sets version 0 (pre Kafka 0.10) are not supported")
	case 1:
		return errors.New("message sets version 1 (Kafka 0.10) are not supported")
	case 2:
		// magic = 2 corresponds to Kafka versions 0.11 and above. Kafka 0.11 replaced message sets with record batches
	default:
		return errors.New(strings.Join([]string{"unrecognized 'magic' number", strconv.Itoa(int(b.magic)), "in records field"}, " "))
	}

	return b.write(buf)
}

// SizeInBytes returns the size of this record batch when it's records are not compressed
func (b *RecordBatch) SizeInBytes() int {
	recordBatchOverhead := 61 // RecordBatch's primitive fields length + 4 bytes holding the records count

	return recordBatchOverhead + b.recordsSizeInBytes()
}

func (b *RecordBatch) IsControlBatch() bool {
	return IsControlBatch(b.attributes)
}

func (b *RecordBatch) IsTransactional() bool {
	return IsTransactional(b.attributes)
}

func (b *RecordBatch) Release() {
	for i := range b.records {
		b.records[i].Release()
	}
	b.records = nil
}

// recordsSizeInBytes returns the length in bytes of records in uncompressed form
func (b *RecordBatch) recordsSizeInBytes() int {
	recordsLength := 0
	for i := range b.records {
		recordsLength += b.records[i].SizeInBytes()
	}

	return recordsLength
}

//nolint:funlen
func (b *RecordBatch) read(rd *typesbytes.ChunkReader) error {
	b.Clear()

	var err error

	var baseOffset int64
	err = types.ReadInt64_(rd, &baseOffset)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"baseOffset\" field")
	}

	var batchLength int32
	err = types.ReadInt32_(rd, &batchLength)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"batchLength\" field")
	}

	if rd.Len() < int(batchLength) {
		return errors.New(strings.Join([]string{"expected", strconv.Itoa(int(batchLength)), "remaining bytes for batch record, but got", strconv.Itoa(rd.Len())}, " "))
	}

	sr := typesbytes.NewSubChunkReader(rd, int64(batchLength))
	r := &sr

	var partitionLeaderEpoch int32
	err = types.ReadInt32_(r, &partitionLeaderEpoch)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"partitionLeaderEpoch\" field")
	}

	var magic int8
	err = types.ReadInt8__(r, &magic)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"magic\" field")
	}

	var crc int32
	err = types.ReadInt32_(r, &crc)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"crc\" field")
	}

	var attributes int16
	err = types.ReadInt16_(r, &attributes)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"attributes\" field")
	}

	var lastOffsetDelta int32
	err = types.ReadInt32_(r, &lastOffsetDelta)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"lastOffsetDelta\" field")
	}

	var baseTimestamp int64
	err = types.ReadInt64_(r, &baseTimestamp)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"baseTimestamp\" field")
	}

	var maxTimestamp int64
	err = types.ReadInt64_(r, &maxTimestamp)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"maxTimestamp\" field")
	}

	var producerId int64
	err = types.ReadInt64_(r, &producerId)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"producerId\" field")
	}

	var producerEpoch int16
	err = types.ReadInt16_(r, &producerEpoch)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"producerEpoch\" field")
	}

	var baseSequence int32
	err = types.ReadInt32_(r, &baseSequence)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's \"baseSequence\" field")
	}

	var numRecords int32
	err = types.ReadInt32_(r, &numRecords)
	if err != nil {
		return errors.WrapIf(err, "couldn't read record.RecordBatch's records count ")
	}
	records := make([]BatchRecord, numRecords)

	compressType := GetCompressionType(attributes)

	var recordsReader *serialization.Reader
	if compressType == None {
		s := serialization.NewReaderWithChunkReader(r)
		recordsReader = &s
	} else {
		s, err := newCompressedRecordsReader(r, compressType)
		if err != nil {
			return err
		}
		recordsReader = &s
		defer recordsReader.Close()
	}

	for i := range records {
		err := records[i].Read(recordsReader)
		if err != nil {
			return errors.WrapIf(err, strings.Join([]string{"couldn't read record.RecordBatch's record item", strconv.Itoa(i)}, " "))
		}
	}

	if IsControlBatch(attributes) && len(records) > 1 {
		return errors.New("control batches should contain only a single record")
	}

	b.baseOffset = baseOffset
	b.batchLength = batchLength
	b.partitionLeaderEpoch = partitionLeaderEpoch
	b.magic = magic
	b.crc = uint32(crc)
	b.attributes = attributes
	b.lastOffsetDelta = lastOffsetDelta
	b.baseTimestamp = baseTimestamp
	b.maxTimestamp = maxTimestamp
	b.producerId = producerId
	b.producerEpoch = producerEpoch
	b.baseSequence = baseSequence
	b.records = records

	b.Complete()

	return nil
}

func (b *RecordBatch) write(buf *typesbytes.SliceWriter) error {
	var err error
	basePos := buf.Len()

	var recordBatchOverhead [61]byte // RecordBatch's primitive fields length + 4 bytes holding the records count

	binary.BigEndian.PutUint64(recordBatchOverhead[0:8], uint64(b.baseOffset))
	binary.BigEndian.PutUint32(recordBatchOverhead[8:12], 0) // batchLength will be set once the length of the batch is known
	binary.BigEndian.PutUint32(recordBatchOverhead[12:16], uint32(b.partitionLeaderEpoch))
	recordBatchOverhead[16] = byte(b.magic)
	binary.BigEndian.PutUint32(recordBatchOverhead[17:21], 0) // crc will be updated when all fields of the batch needed to compute the crc has been serialized
	binary.BigEndian.PutUint16(recordBatchOverhead[21:23], uint16(b.attributes))
	binary.BigEndian.PutUint32(recordBatchOverhead[23:27], uint32(b.lastOffsetDelta))
	binary.BigEndian.PutUint64(recordBatchOverhead[27:35], uint64(b.baseTimestamp))
	binary.BigEndian.PutUint64(recordBatchOverhead[35:43], uint64(b.maxTimestamp))
	binary.BigEndian.PutUint64(recordBatchOverhead[43:51], uint64(b.producerId))
	binary.BigEndian.PutUint16(recordBatchOverhead[51:53], uint16(b.producerEpoch))
	binary.BigEndian.PutUint32(recordBatchOverhead[53:57], uint32(b.baseSequence))
	binary.BigEndian.PutUint32(recordBatchOverhead[57:61], uint32(len(b.records))) // records count

	n, err := buf.Write(recordBatchOverhead[:])
	if err != nil {
		return errors.WrapIf(err, "couldn't serialize record batch overhead fields")
	}
	if n != 61 {
		return errors.New(strings.Join([]string{"expected to write 61 bytes of record batch overhead fields but wrote", strconv.Itoa(n)}, " "))
	}

	batchLength := 49 // covers all fields starting from partitionLeaderEpoch to records included
	//nolint: nestif
	if len(b.records) > 0 {
		var recordsWriter *serialization.Writer

		recordsStartBufPos := buf.Len()
		compressType := GetCompressionType(b.attributes)

		if compressType == None {
			s := serialization.NewWriterWithSliceWriter(buf)
			recordsWriter = &s
		} else {
			s, err := newCompressedRecordsWriter(buf, compressType)
			if err != nil {
				return err
			}
			recordsWriter = &s
			defer recordsWriter.Close()
		}

		for i := range b.records {
			if err := b.records[i].Write(recordsWriter); err != nil {
				return errors.WrapIf(err, strings.Join([]string{"couldn't serialize batch record item", strconv.Itoa(i)}, " "))
			}
		}
		// flush serialized records
		err = recordsWriter.Close()
		if err != nil {
			return errors.WrapIf(err, "couldn't close compressed record writer")
		}

		recordsEndBufPos := buf.Len()
		batchLength += recordsEndBufPos - recordsStartBufPos // add records length to batchLength
	}
	if int64(batchLength) > math.MaxUint32 {
		return errors.New(strings.Join([]string{"batchLength has invalid value:", strconv.Itoa(batchLength)}, " "))
	}

	batchLengthBytes, err := buf.Slice(basePos+8, basePos+12)
	if err != nil {
		return errors.WrapIf(err, "couldn't get the byte slice holding batch length")
	}
	binary.BigEndian.PutUint32(batchLengthBytes, uint32(batchLength))

	// the CRC covers the data from the attributes to the end of the batch (i.e. all the bytes that follow the CRC)
	crcInputBytes, err := buf.Slice(basePos+21, buf.Len())
	if err != nil {
		return errors.WrapIf(err, "couldn't get the byte slice to compute CRC from")
	}
	crc := crc32c.Compute(crcInputBytes)
	crcBytes, err := buf.Slice(basePos+17, basePos+21)
	if err != nil {
		return errors.WrapIf(err, "couldn't get the byte slice holding crc")
	}
	binary.BigEndian.PutUint32(crcBytes, crc)

	return nil
}

// CompressionType returns true if the passed in attributes has either 'gzip', 'snappy', 'lz4' or 'zstd' flag set
// according to https://kafka.apache.org/documentation/#recordbatch
func (b *RecordBatch) CompressionType() CompressionType {
	return GetCompressionType(b.attributes)
}

// IsControlBatch returns true if the passed in attributes has the 'isControlBatch' flag set
// according to https://kafka.apache.org/documentation/#recordbatch
func IsControlBatch(attributes int16) bool {
	return attributes&0x20 != 0
}

// IsTransactional returns true if the passed in attributes has the 'isTransactional' flag set
// according to https://kafka.apache.org/documentation/#recordbatch
func IsTransactional(attributes int16) bool {
	return attributes&0x10 != 0
}

// GetCompressionType returns true if the passed in attributes has either 'gzip', 'snappy', 'lz4' or 'zstd' flag set
// according to https://kafka.apache.org/documentation/#recordbatch
func GetCompressionType(attributes int16) CompressionType {
	return CompressionType(attributes & 0x07)
}

func newCompressedRecordsReader(r *typesbytes.ChunkReader, compressionType CompressionType) (serialization.Reader, error) {
	switch compressionType {
	case None:
		return serialization.Reader{}, nil
	case Gzip:
		gz, err := gzip.NewReader(r)
		if err != nil {
			return serialization.Reader{}, errors.New("couldn't create gzip decoder")
		}
		return serialization.NewReaderWithGzipDecoder(gz), nil
	case Snappy:
		snappyDecoder, err := snappy.NewReader(r)
		if err != nil {
			return serialization.Reader{}, errors.New("couldn't create snappy decoder")
		}
		return serialization.NewReaderWithSnappyDecoder(snappyDecoder), nil
	case Lz4:
		lz4Decoder, err := lz4.NewReader(r)
		if err != nil {
			return serialization.Reader{}, errors.New("couldn't create lz4 decoder")
		}
		return serialization.NewReaderWithWithLz4Decoder(lz4Decoder), nil
	case Zstd:
		zstdDecoder, err := zstd.NewReader(r)
		if err != nil {
			return serialization.Reader{}, errors.New("couldn't create zstd decoder")
		}
		return serialization.NewReaderWithZstdDecoder(zstdDecoder), nil
	default:
		return serialization.Reader{}, errors.New(strings.Join([]string{"unknown compression type id:", strconv.Itoa(int(compressionType))}, " "))
	}
}

func newCompressedRecordsWriter(w *typesbytes.SliceWriter, compressionType CompressionType) (serialization.Writer, error) {
	switch compressionType {
	case None:
		return serialization.Writer{}, nil
	case Gzip:
		return serialization.NewWriterWithGzipEncoder(gzip.NewWriter(w)), nil
	case Snappy:
		swr, err := snappy.NewWriter(w)
		if err != nil {
			return serialization.Writer{}, err
		}
		return serialization.NewWriterWithSnappyEncoder(swr), nil
	case Lz4:
		lzwr, err := lz4.NewWriter(w)
		if err != nil {
			return serialization.Writer{}, err
		}
		return serialization.NewWriterWithLz4Encoder(lzwr), nil
	case Zstd:
		zwr, err := zstd.NewWriter(w)
		if err != nil {
			return serialization.Writer{}, err
		}
		return serialization.NewWriterWithZstdEncoder(zwr), nil
	default:
		return serialization.Writer{}, errors.New(strings.Join([]string{"unknown compression type id:", strconv.Itoa(int(compressionType))}, " "))
	}
}
