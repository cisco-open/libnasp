// Copyright (c) 2023 Cisco and/or its affiliates. All rights reserved.
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
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	//_ "github.com/wasilibs/nottinygc"
	"io"
	"strconv"
	"strings"
	"wwwin-github.cisco.com/eti/kafka-protocol-go/pkg/request"
	"wwwin-github.cisco.com/eti/kafka-protocol-go/pkg/response"
)

const kafkaMsgSizeBytesLen = 4 // kafka messages size is represented on 4 bytes

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

//export sched_yield
func sched_yield() int32 {
	return 0
}

type vmContext struct {
	types.DefaultVMContext
}

func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{contextID: contextID}
}

type pluginContext struct {
	types.DefaultPluginContext

	contextID uint32
}

type state int

const (
	running state = iota
	failed
)

type trafficDirection int32

const (
	// Default option is unspecified.
	traffic_direction_unspecified trafficDirection = 0
	// The transport is used for incoming traffic.
	traffic_direction_inbound trafficDirection = 1
	// The transport is used for outgoing traffic.
	traffic_direction_outbound trafficDirection = 2
)

func (ctx *pluginContext) NewTcpContext(contextID uint32) types.TcpContext {
	listenerDirection, err := proxywasm.GetProperty([]string{"listener_direction"})
	if err != nil {
		proxywasm.LogCritical(strings.Join([]string{"couldn't get listener_direction:", err.Error()}, " "))
	}

	return &networkContext{
		contextID:        contextID,
		inFlightRequests: make([]requestReference, 0, 100),
		direction:        trafficDirection(listenerDirection[0]),
	}
}

type networkContext struct {
	types.DefaultTcpContext

	contextID uint32
	direction trafficDirection

	// incompleteDownstreamData holds downstream data fragment of a Kafka message that
	// follows the current Kafka message being processed
	incompleteDownstreamData bytes.Buffer
	// outputDownstreamData holds the resulting downstream data produced by the filter
	outputDownstreamData bytes.Buffer

	// incompleteUpstreamData holds downstream data fragment of a Kafka message that
	// follows the current Kafka message being processed
	incompleteUpstreamData bytes.Buffer
	// outputUpstreamData holds the resulting upstream data produced by the filter
	outputUpstreamData bytes.Buffer

	requestHandlerState state
	requestMessageSize  int32

	responseHandlerState state
	responseMessageSize  int32

	inFlightRequests []requestReference
	r                int
}

type requestReference struct {
	requestCorrelationId int32
	requestApiKey        int16
	requestApiVersion    int16
}

func (ctx *networkContext) OnNewConnection() types.Action {
	proxywasm.LogDebug("new connection!")

	ctx.resetRequestHandler()
	ctx.resetResponseHandler()

	ctx.resetInFlightRequests()

	ctx.resetDownstreamBuffers()
	ctx.resetUpstreamBuffers()

	return types.ActionContinue
}

func (ctx *networkContext) OnDownstreamData(dataSize int, endOfStream bool) types.Action {
	if ctx.requestHandlerState == failed || ctx.responseHandlerState == failed || ctx.direction == traffic_direction_unspecified {
		// in failed status, can not do any further processing until new connection initiated
		return types.ActionContinue
	}

	availableDataSize := dataSize + ctx.incompleteDownstreamData.Len()
	if endOfStream && availableDataSize == 0 {
		// we don't expect to receive more data so pass control to next filters in the chain
		ctx.resetDownstreamBuffers()
		ctx.resetDownstreamMessageHandler()
		return types.ActionContinue
	}

	// wait until downstream data is accumulated by host that is enough to process a kafka message
	ready, err := ctx.downstreamKafkaMessageDataAvailable(dataSize)
	if err != nil {
		proxywasm.LogCritical(strings.Join([]string{"couldn't verify if the entire downstream Kafka message data has been received over the network due to:", err.Error()}, " "))
		ctx.failDownstreamMessageHandler()
		return types.ActionContinue
	}
	if !ready {
		if endOfStream && dataSize == 0 {
			// we don't expect to receive more data so pass control to next filters in the chain
			ctx.resetDownstreamBuffers()
			ctx.resetDownstreamMessageHandler()
			return types.ActionContinue
		}
		return types.ActionPause // wait for more data
	}

	data, err := proxywasm.GetDownstreamData(0, dataSize)
	if err != nil && !errors.Is(err, io.EOF) {
		proxywasm.LogCritical(strings.Join([]string{"couldn't get downstream data bytes that holds kafka message due to:", err.Error()}, " "))
		ctx.failDownstreamMessageHandler()
		return types.ActionContinue
	}
	if len(data) < dataSize {
		proxywasm.LogCritical(strings.Join([]string{"expected to read", strconv.Itoa(dataSize), "downstream data bytes but got only", strconv.Itoa(len(data))}, " "))
		ctx.failDownstreamMessageHandler()
		return types.ActionContinue
	}

	ctx.incompleteDownstreamData.Write(data)

	// at this stage the host might have buffered bytes that cover more complete messages and an incomplete one thus
	// we iterate through the complete ones and leave the incomplete one in the internal buffer
	for {
		if ctx.incompleteDownstreamData.Len() == 0 {
			ctx.onCompleteDownstreamDataProcessingDone()
			return types.ActionContinue
		}

		msgSize, err := kafkaMessageSize(ctx.incompleteDownstreamData.Bytes())
		if err != nil {
			proxywasm.LogCritical(strings.Join([]string{"couldn't retrieve kafka message size from downstream data due to:", err.Error()}, " "))
			ctx.failDownstreamMessageHandler()
			return types.ActionContinue
		}
		completeDataSize := int(kafkaMsgSizeBytesLen + msgSize)
		if ctx.incompleteDownstreamData.Len() < completeDataSize {
			ctx.onCompleteDownstreamDataProcessingDone()
			return types.ActionContinue
		}

		sizeAndMsgData := ctx.incompleteDownstreamData.Next(completeDataSize)
		err = ctx.handleDownstreamKafkaMessage(sizeAndMsgData)
		if err != nil {
			proxywasm.LogCritical(strings.Join([]string{"couldn't handle kafka request due to:", err.Error()}, " "))
			ctx.failDownstreamMessageHandler()

			err = proxywasm.ReplaceDownstreamData(ctx.incompleteDownstreamData.Bytes())
			if err != nil {
				proxywasm.LogCritical(strings.Join([]string{"couldn't replace downstream data due to:", err.Error()}, " "))
			}
			return types.ActionContinue
		}
	}
}

func (ctx *networkContext) OnUpstreamData(dataSize int, endOfStream bool) types.Action {
	if ctx.requestHandlerState == failed || ctx.responseHandlerState == failed || ctx.direction == traffic_direction_unspecified {
		// in failed status, can not do any further processing until new connection initiated
		return types.ActionContinue
	}

	availableDataSize := dataSize + ctx.incompleteUpstreamData.Len()
	if endOfStream && availableDataSize == 0 {
		// we don't expect to receive more data so pass control to next filters in the chain
		ctx.resetUpstreamBuffers()
		ctx.resetUpstreamMessageHandler()
		return types.ActionContinue
	}

	// wait until upstream data is accumulated by host that is enough to process a kafka message
	ready, err := ctx.upstreamKafkaMessageDataAvailable(dataSize)
	if err != nil {
		proxywasm.LogCritical(strings.Join([]string{"couldn't verify if the entire upstream Kafka message data has been received over the network due to:", err.Error()}, " "))
		ctx.failUpstreamMessageHandler()
		return types.ActionContinue
	}
	if !ready {
		if endOfStream && dataSize == 0 {
			// we don't expect to receive more data so pass control to next filters in the chain
			ctx.resetUpstreamBuffers()
			ctx.resetUpstreamMessageHandler()
			return types.ActionContinue
		}
		return types.ActionPause // wait for more data
	}

	data, err := proxywasm.GetUpstreamData(0, dataSize)
	if err != nil && !errors.Is(err, io.EOF) {
		proxywasm.LogCritical(strings.Join([]string{"couldn't get upstream data bytes that holds kafka message due to:", err.Error()}, " "))
		ctx.failUpstreamMessageHandler()
		return types.ActionContinue
	}
	if len(data) < dataSize {
		proxywasm.LogCritical(strings.Join([]string{"expected to read", strconv.Itoa(dataSize), "upstream data bytes but got only", strconv.Itoa(len(data))}, " "))
		ctx.failUpstreamMessageHandler()
		return types.ActionContinue
	}

	ctx.incompleteUpstreamData.Write(data)

	// at this stage the host might have buffered bytes that cover more complete messages and an incomplete one thus
	// we iterate through the complete ones and leave the incomplete one in the internal buffer
	for {
		if ctx.incompleteUpstreamData.Len() == 0 {
			ctx.onCompleteUpstreamDataProcessingDone()
			return types.ActionContinue
		}

		msgSize, err := kafkaMessageSize(ctx.incompleteUpstreamData.Bytes())
		if err != nil {
			proxywasm.LogCritical(strings.Join([]string{"couldn't retrieve kafka message size from upstream data due to:", err.Error()}, " "))
			ctx.failUpstreamMessageHandler()
			return types.ActionContinue
		}
		completeDataSize := int(kafkaMsgSizeBytesLen + msgSize)
		if ctx.incompleteUpstreamData.Len() < completeDataSize {
			ctx.onCompleteUpstreamDataProcessingDone()
			return types.ActionContinue
		}

		sizeAndMsgData := ctx.incompleteUpstreamData.Next(completeDataSize)
		err = ctx.handleUpstreamKafkaMessage(sizeAndMsgData)
		if err != nil {
			proxywasm.LogCritical(strings.Join([]string{"couldn't handle kafka response due to:", err.Error()}, " "))
			ctx.failUpstreamMessageHandler()

			err = proxywasm.ReplaceUpstreamData(ctx.incompleteUpstreamData.Bytes())
			if err != nil {
				proxywasm.LogCritical(strings.Join([]string{"couldn't replace upstream data due to:", err.Error()}, " "))
			}
			return types.ActionContinue
		}
	}
}

func (ctx *networkContext) OnDownstreamClose(_ types.PeerType) {
	ctx.resetRequestHandler()
	ctx.resetResponseHandler()
	ctx.resetDownstreamBuffers()
}

func (ctx *networkContext) OnUpstreamClose(_ types.PeerType) {
	ctx.resetRequestHandler()
	ctx.resetResponseHandler()
	ctx.resetUpstreamBuffers()
}

func (ctx *networkContext) resetDownstreamMessageHandler() {
	switch ctx.direction {
	case traffic_direction_inbound:
		ctx.resetRequestHandler()
	case traffic_direction_outbound:
		ctx.resetResponseHandler()
	}
}

func (ctx *networkContext) resetUpstreamMessageHandler() {
	switch ctx.direction {
	case traffic_direction_inbound:
		ctx.resetResponseHandler()
	case traffic_direction_outbound:
		ctx.resetRequestHandler()
	}
}

func (ctx *networkContext) resetRequestHandler() {
	ctx.requestHandlerState = running
	ctx.requestMessageSize = 0
}

func (ctx *networkContext) resetResponseHandler() {
	ctx.responseHandlerState = running
	ctx.responseMessageSize = 0
}

func (ctx *networkContext) failDownstreamMessageHandler() {
	switch ctx.direction {
	case traffic_direction_inbound:
		ctx.failRequestHandler()
	case traffic_direction_outbound:
		ctx.failResponseHandler()
	}
}

func (ctx *networkContext) failUpstreamMessageHandler() {
	switch ctx.direction {
	case traffic_direction_inbound:
		ctx.failResponseHandler()
	case traffic_direction_outbound:
		ctx.failRequestHandler()
	}
}

func (ctx *networkContext) failRequestHandler() {
	ctx.requestHandlerState = failed
	ctx.resetInFlightRequests()

}

func (ctx *networkContext) failResponseHandler() {
	ctx.responseHandlerState = failed
	ctx.resetInFlightRequests()
}

func (ctx *networkContext) handleDownstreamKafkaMessage(sizeAndMsgData []byte) error {
	var err error
	switch ctx.direction {
	case traffic_direction_inbound:
		err = ctx.handleKafkaRequestMessage(sizeAndMsgData)
	case traffic_direction_outbound:
		err = ctx.handleKafkaResponseMessage(sizeAndMsgData)
	}

	if err != nil {
		return err
	}

	ctx.outputDownstreamData.Write(sizeAndMsgData)
	return nil
}

func (ctx *networkContext) handleUpstreamKafkaMessage(sizeAndMsgData []byte) error {
	var err error
	switch ctx.direction {
	case traffic_direction_inbound:
		err = ctx.handleKafkaResponseMessage(sizeAndMsgData)
	case traffic_direction_outbound:
		err = ctx.handleKafkaRequestMessage(sizeAndMsgData)
	}

	if err != nil {
		return err
	}
	ctx.outputUpstreamData.Write(sizeAndMsgData)

	return nil
}

func (ctx *networkContext) handleKafkaRequestMessage(sizeAndMsgData []byte) error {
	req, err := request.Parse(sizeAndMsgData[kafkaMsgSizeBytesLen:])
	if err != nil {
		return errors.New(strings.Join([]string{"couldn't parse Kafka request message data bytes due to:", err.Error(), ", raw size and message:", base64.StdEncoding.EncodeToString(sizeAndMsgData)}, " "))
	}
	defer req.Release()

	proxywasm.LogDebug(strings.Join([]string{"processed Kafka request message:", req.HeaderData().String()}, " "))

	ctx.enqueueInFlightRequest(req.HeaderData().RequestApiKey(), req.HeaderData().RequestApiVersion(), req.HeaderData().CorrelationId())

	return nil
}

func (ctx *networkContext) handleKafkaResponseMessage(sizeAndMsgData []byte) error {
	if len(ctx.inFlightRequests) == 0 {
		return nil // skip as there is no req to match this response to
	}
	req := ctx.nextInFlightRequest()

	resp, err := response.Parse(sizeAndMsgData[kafkaMsgSizeBytesLen:], req.requestApiKey, req.requestApiVersion, req.requestCorrelationId)
	if err != nil {
		return errors.New(strings.Join([]string{"couldn't parse Kafka response message data bytes due to:", err.Error(), ", raw size and message:", base64.StdEncoding.EncodeToString(sizeAndMsgData)}, " "))
	}
	defer resp.Release()

	proxywasm.LogDebug(strings.Join([]string{"processed Kafka response message:", resp.String()}, " "))

	return nil
}

func (ctx *networkContext) onCompleteDownstreamDataProcessingDone() {
	// replace stream with complete part
	err := proxywasm.ReplaceDownstreamData(ctx.outputDownstreamData.Bytes())
	if err != nil {
		proxywasm.LogCritical(strings.Join([]string{"couldn't replace downstream data due to:", err.Error()}, " "))
		ctx.failDownstreamMessageHandler()

		return
	}
	ctx.outputDownstreamData.Reset()
	ctx.resetDownstreamMessageHandler()

}

func (ctx *networkContext) onCompleteUpstreamDataProcessingDone() {
	// replace stream with complete part
	err := proxywasm.ReplaceUpstreamData(ctx.outputUpstreamData.Bytes())
	if err != nil {
		proxywasm.LogCritical(strings.Join([]string{"couldn't replace upstream data due to:", err.Error()}, " "))
		ctx.failUpstreamMessageHandler()
		return
	}
	ctx.outputUpstreamData.Reset()
	ctx.resetUpstreamMessageHandler()
}

func (ctx *networkContext) resetDownstreamBuffers() {
	ctx.incompleteDownstreamData.Reset()
	ctx.outputDownstreamData.Reset()
}

func (ctx *networkContext) resetUpstreamBuffers() {
	ctx.incompleteUpstreamData.Reset()
	ctx.outputUpstreamData.Reset()
}

// downstreamKafkaMessageDataAvailable returns true if there is enough downstream data to process the current Kafka message
func (ctx *networkContext) downstreamKafkaMessageDataAvailable(dataSize int) (bool, error) {
	availableDataSize := dataSize + ctx.incompleteDownstreamData.Len()

	var cachedMessageSize int32
	switch ctx.direction {
	case traffic_direction_inbound:
		cachedMessageSize = ctx.requestMessageSize
	case traffic_direction_outbound:
		cachedMessageSize = ctx.responseMessageSize
	}

	var err error
	if cachedMessageSize == 0 {
		// message size hasn't been received yet
		if availableDataSize < kafkaMsgSizeBytesLen {
			proxywasm.LogDebug(strings.Join([]string{strconv.Itoa(availableDataSize), "downstream data bytes available, waiting for more to process Kafka message size"}, " "))
			return false, nil // wait for more bytes as incomplete data together with newly received data doesn't have enough bytes to parse message size
		}

		cachedMessageSize, err = ctx.kafkaMessageSizeFromDownstreamData()
		if err != nil {
			proxywasm.LogCritical(strings.Join([]string{"couldn't retrieve kafka message size from downstream data due to:", err.Error()}, " "))
			return false, err
		}

		switch ctx.direction {
		case traffic_direction_inbound:
			ctx.requestMessageSize = cachedMessageSize
		case traffic_direction_outbound:
			ctx.responseMessageSize = cachedMessageSize
		}
	}

	completeDataSize := int(kafkaMsgSizeBytesLen + cachedMessageSize) // the size is message size + message size
	if availableDataSize < completeDataSize {
		proxywasm.LogDebug(strings.Join([]string{"expected", strconv.Itoa(completeDataSize), "downstream data bytes to be available to process Kafka message(available:", strconv.Itoa(availableDataSize), "bytes)...waiting for more"}, " "))
		return false, nil // wait for more bytes as we haven't received all bytes needed to be able to parse the message
	}

	return true, nil
}

// upstreamKafkaMessageDataAvailable returns true if there is enough downstream data to process the current Kafka message
func (ctx *networkContext) upstreamKafkaMessageDataAvailable(dataSize int) (bool, error) {
	availableDataSize := dataSize + ctx.incompleteUpstreamData.Len()

	var cachedMessageSize int32
	switch ctx.direction {
	case traffic_direction_inbound:
		cachedMessageSize = ctx.responseMessageSize
	case traffic_direction_outbound:
		cachedMessageSize = ctx.requestMessageSize
	}

	var err error
	if cachedMessageSize == 0 {
		// message size hasn't been received yet
		if availableDataSize < kafkaMsgSizeBytesLen {
			proxywasm.LogDebug(strings.Join([]string{strconv.Itoa(availableDataSize), "downstream data bytes available, waiting for more to process Kafka message size"}, " "))
			return false, nil // wait for more bytes as incomplete data together with newly received data doesn't have enough bytes to parse message size
		}

		cachedMessageSize, err = ctx.kafkaMessageSizeFromUpstreamData()
		if err != nil {
			proxywasm.LogCritical(strings.Join([]string{"couldn't retrieve kafka message size from upstream data due to:", err.Error()}, " "))
			return false, err
		}

		switch ctx.direction {
		case traffic_direction_inbound:
			ctx.responseMessageSize = cachedMessageSize
		case traffic_direction_outbound:
			ctx.requestMessageSize = cachedMessageSize
		}
	}

	completeDataSize := int(kafkaMsgSizeBytesLen + cachedMessageSize) // the size is message size + message size
	if availableDataSize < completeDataSize {
		proxywasm.LogDebug(strings.Join([]string{"expected", strconv.Itoa(completeDataSize), "upstream data bytes to be available to process Kafka message(available:", strconv.Itoa(availableDataSize), "bytes)...waiting for more"}, " "))
		return false, nil // wait for more bytes as we haven't received all bytes needed to be able to parse the message
	}

	return true, nil
}

func (ctx *networkContext) enqueueInFlightRequest(requestApiKey, requestApiVersion int16, requestCorrelationId int32) {
	reqRef := requestReference{
		requestApiKey:        requestApiKey,
		requestApiVersion:    requestApiVersion,
		requestCorrelationId: requestCorrelationId,
	}

	l := len(ctx.inFlightRequests)
	c := cap(ctx.inFlightRequests)

	if l >= c {
		if ctx.r >= c/2 {
			l = copy(ctx.inFlightRequests, ctx.inFlightRequests[ctx.r:])
			ctx.r = 0
		} else {
			ctx.inFlightRequests = append(ctx.inFlightRequests, reqRef)
			return
		}
	}

	ctx.inFlightRequests = ctx.inFlightRequests[:l+1]
	ctx.inFlightRequests[l] = reqRef
}

func (ctx *networkContext) nextInFlightRequest() requestReference {
	reqRef := ctx.inFlightRequests[ctx.r]

	ctx.r++
	if ctx.r >= len(ctx.inFlightRequests) {
		ctx.resetInFlightRequests()
	}

	return reqRef
}

func (ctx *networkContext) resetInFlightRequests() {
	ctx.r = 0
	ctx.inFlightRequests = ctx.inFlightRequests[:0]
}

func (ctx *networkContext) kafkaMessageSizeFromDownstreamData() (int32, error) {
	var msgSizeBytes [kafkaMsgSizeBytesLen]byte
	n := 0

	if ctx.incompleteDownstreamData.Len() > 0 {
		n = copy(msgSizeBytes[:], ctx.incompleteDownstreamData.Bytes())
	}
	if n < kafkaMsgSizeBytesLen {
		b, err := proxywasm.GetDownstreamData(0, kafkaMsgSizeBytesLen-n)
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, errors.New(strings.Join([]string{"couldn't get", strconv.Itoa(kafkaMsgSizeBytesLen - n), "downstream data bytes from host"}, " "))
		}
		m := copy(msgSizeBytes[n:], b)
		n += m
	}

	size, err := kafkaMessageSize(msgSizeBytes[0:kafkaMsgSizeBytesLen])
	if err != nil {
		return 0, errors.New(strings.Join([]string{"couldn't deserialize kafka message size due to:", err.Error()}, ""))
	}

	return size, nil
}

func (ctx *networkContext) kafkaMessageSizeFromUpstreamData() (int32, error) {
	var msgSizeBytes [kafkaMsgSizeBytesLen]byte
	n := 0

	if ctx.incompleteUpstreamData.Len() > 0 {
		n = copy(msgSizeBytes[:], ctx.incompleteUpstreamData.Bytes())
	}
	if n < kafkaMsgSizeBytesLen {
		b, err := proxywasm.GetUpstreamData(0, kafkaMsgSizeBytesLen-n)
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, errors.New(strings.Join([]string{"couldn't get", strconv.Itoa(kafkaMsgSizeBytesLen - n), "upstream data bytes from host"}, " "))
		}
		m := copy(msgSizeBytes[n:], b)
		n += m
	}

	size, err := kafkaMessageSize(msgSizeBytes[0:kafkaMsgSizeBytesLen])
	if err != nil {
		return 0, errors.New(strings.Join([]string{"couldn't deserialize kafka message size due to:", err.Error()}, ""))
	}

	return size, nil
}

func kafkaMessageSize(msgSizeBytes []byte) (int32, error) {
	n := len(msgSizeBytes)
	if n < kafkaMsgSizeBytesLen {
		return 0, errors.New(strings.Join([]string{"expected", strconv.Itoa(kafkaMsgSizeBytesLen), "bytes, but got", strconv.Itoa(n)}, " "))
	}

	return int32(binary.BigEndian.Uint32(msgSizeBytes)), nil
}
