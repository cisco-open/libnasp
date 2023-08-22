package main

import (
	"encoding/binary"
	"errors"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	"strconv"
	"strings"
	"wwwin-github.cisco.com/eti/kafka-protocol-go/pkg/request"
	"wwwin-github.cisco.com/eti/kafka-protocol-go/pkg/response"
	//_ "github.com/wasilibs/nottinygc"
)

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
	waitingMessageSizeBytes state = iota
	waitingMessageBytes
	failed
)

func (ctx *pluginContext) NewTcpContext(contextID uint32) types.TcpContext {
	return &networkContext{
		contextID:        contextID,
		inFlightRequests: make([]requestReference, 0, 100),
	}
}

type networkContext struct {
	types.DefaultTcpContext

	contextID uint32

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

	ctx.r = 0
	ctx.inFlightRequests = ctx.inFlightRequests[:0]

	return types.ActionContinue
}

func (ctx *networkContext) OnDownstreamData(dataSize int, endOfStream bool) types.Action {
	if ctx.requestHandlerState == failed || ctx.responseHandlerState == failed {
		// in failed status, can not do any further processing until new connection initiated
		return types.ActionContinue
	}

	if endOfStream && dataSize == 0 {
		// we don't expect more data so pass control to next filters in the chain
		return ctx.succeedRequestHandler()
	}

	switch ctx.requestHandlerState {
	case waitingMessageSizeBytes:
		if dataSize < 4 {
			if endOfStream {
				return types.ActionContinue
			}

			proxywasm.LogDebug(strings.Join([]string{strconv.Itoa(dataSize), "downstream data bytes available, waiting for more to process Kafka request message size"}, " "))
			return types.ActionPause // wait for more bytes
		}

		msgSizeBytes, err := proxywasm.GetDownstreamData(0, 4)
		if err != nil {
			proxywasm.LogCritical(strings.Join([]string{"couldn't get downstream data bytes that holds kafka message size due to:", err.Error()}, " "))
			return ctx.failRequestHandler()
		}

		msgSize, err := kafkaMessageSize(msgSizeBytes)
		if err != nil {
			proxywasm.LogCritical(strings.Join([]string{"couldn't deserialize kafka message size due to:", err.Error()}, " "))
			return ctx.failRequestHandler()
		}

		ctx.requestMessageSize = msgSize
		return ctx.handleKafkaRequestMessage(dataSize, endOfStream)

	case waitingMessageBytes:
		return ctx.handleKafkaRequestMessage(dataSize, endOfStream)
	}

	return types.ActionContinue
}

func (ctx *networkContext) OnUpstreamData(dataSize int, endOfStream bool) types.Action {
	if ctx.requestHandlerState == failed || ctx.responseHandlerState == failed {
		// in failed status, can not do any further processing until new connection initiated
		return types.ActionContinue
	}

	if endOfStream && dataSize == 0 {
		// we don't expect more data so pass control to next filters in the chain
		return ctx.succeedResponseHandler()
	}

	switch ctx.responseHandlerState {
	case waitingMessageSizeBytes:
		if dataSize < 4 {
			if endOfStream {
				return types.ActionContinue
			}

			proxywasm.LogDebug(strings.Join([]string{strconv.Itoa(dataSize), "upstream data bytes available, waiting for more to process Kafka message size"}, " "))
			return types.ActionPause // wait for more bytes
		}

		msgSizeBytes, err := proxywasm.GetUpstreamData(0, 4)
		if err != nil {
			proxywasm.LogCritical(strings.Join([]string{"couldn't get upstream data bytes that holds kafka response message size due to:", err.Error()}, " "))
			return ctx.failResponseHandler()
		}

		msgSize, err := kafkaMessageSize(msgSizeBytes)
		if err != nil {
			proxywasm.LogCritical(strings.Join([]string{"couldn't deserialize kafka message size due to:", err.Error()}, " "))
			return ctx.failResponseHandler()
		}

		ctx.responseMessageSize = msgSize
		return ctx.handleKafkaResponseMessage(dataSize, endOfStream)
	case waitingMessageBytes:
		return ctx.handleKafkaResponseMessage(dataSize, endOfStream)
	}
	return types.ActionContinue
}

func (ctx *networkContext) OnDownstreamClose(_ types.PeerType) {
	ctx.resetRequestHandler()
	ctx.resetResponseHandler()
}

func (ctx *networkContext) OnUpstreamClose(_ types.PeerType) {
	ctx.resetRequestHandler()
	ctx.resetResponseHandler()
}

func (ctx *networkContext) resetRequestHandler() {
	ctx.requestHandlerState = waitingMessageSizeBytes
	ctx.requestMessageSize = 0
}

func (ctx *networkContext) resetResponseHandler() {
	ctx.responseHandlerState = waitingMessageSizeBytes
	ctx.responseMessageSize = 0
}

func (ctx *networkContext) failRequestHandler() types.Action {
	ctx.requestHandlerState = failed

	ctx.r = 0
	ctx.inFlightRequests = ctx.inFlightRequests[:0]

	return types.ActionContinue
}

func (ctx *networkContext) failResponseHandler() types.Action {
	ctx.responseHandlerState = failed

	ctx.r = 0
	ctx.inFlightRequests = ctx.inFlightRequests[:0]

	return types.ActionContinue
}

func (ctx *networkContext) succeedRequestHandler() types.Action {
	ctx.resetRequestHandler()
	return types.ActionContinue
}

func (ctx *networkContext) succeedResponseHandler() types.Action {
	ctx.resetResponseHandler()
	return types.ActionContinue
}

func (ctx *networkContext) handleKafkaRequestMessage(dataSize int, endOfStream bool) types.Action {
	expectedDataSize := int(ctx.requestMessageSize) + 4
	if dataSize < expectedDataSize {
		if endOfStream {
			return types.ActionContinue
		}

		ctx.requestHandlerState = waitingMessageBytes
		proxywasm.LogDebug(strings.Join([]string{"expected", strconv.Itoa(expectedDataSize), "downstream data bytes to be available to process Kafka request message(available:", strconv.Itoa(dataSize), "bytes)...waiting for more"}, " "))
		return types.ActionPause // wait for more bytes
	}

	if dataSize > expectedDataSize {
		// more bytes available as expected, most likely added by other filters thus we can ignore them; we just log it to be aware of it
		proxywasm.LogWarn(strings.Join([]string{"more than expected downstream data bytes available, expected:", strconv.Itoa(expectedDataSize), ", got:", strconv.Itoa(dataSize)}, " "))
	}

	msg, err := proxywasm.GetDownstreamData(4, int(ctx.requestMessageSize))
	if err != nil {
		proxywasm.LogCritical(strings.Join([]string{"couldn't get downstream data bytes that holds kafka message due to:", err.Error()}, " "))
		return ctx.failRequestHandler()
	}

	req, err := request.Parse(msg)
	if err != nil {
		proxywasm.LogCritical(strings.Join([]string{"couldn't parse Kafka request message data bytes due to:", err.Error()}, " "))
		return ctx.failRequestHandler()
	}
	defer req.Release()

	proxywasm.LogDebug(strings.Join([]string{"processed Kafka request message:", req.HeaderData().String()}, " "))

	ctx.enqueueInFlightRequest(req.HeaderData().RequestApiKey(), req.HeaderData().RequestApiVersion(), req.HeaderData().CorrelationId())
	return ctx.succeedRequestHandler()
}

func (ctx *networkContext) handleKafkaResponseMessage(dataSize int, endOfStream bool) types.Action {
	expectedDataSize := int(ctx.responseMessageSize) + 4
	if dataSize < expectedDataSize {
		if endOfStream {
			return types.ActionContinue
		}

		ctx.responseHandlerState = waitingMessageBytes
		proxywasm.LogDebug(strings.Join([]string{"expected", strconv.Itoa(expectedDataSize), "upstream data bytes to be available to process Kafka response message(available:", strconv.Itoa(dataSize), "bytes)...waiting for more"}, " "))
		return types.ActionPause // wait for more bytes
	}

	if dataSize > expectedDataSize {
		// more bytes available as expected, most likely added by other filters thus we can ignore them; we just log it to be aware of it
		proxywasm.LogWarn(strings.Join([]string{"more than expected upstream data bytes available, expected:", strconv.Itoa(expectedDataSize), ", got:", strconv.Itoa(dataSize)}, " "))
	}

	if len(ctx.inFlightRequests) == 0 {
		return types.ActionContinue // skip as there is no req to match this response to
	}
	req := ctx.nextInFlightRequest()

	msg, err := proxywasm.GetUpstreamData(4, int(ctx.responseMessageSize))
	if err != nil {
		proxywasm.LogCritical(strings.Join([]string{"couldn't get upstream data bytes that holds kafka message due to: %v", err.Error()}, " '"))
		return ctx.failResponseHandler()
	}

	resp, err := response.Parse(msg, req.requestApiKey, req.requestApiVersion, req.requestCorrelationId)
	if err != nil {
		proxywasm.LogCritical(strings.Join([]string{"couldn't parse Kafka response message data bytes due to:", err.Error()}, " "))
		return ctx.failResponseHandler()
	}
	defer resp.Release()

	proxywasm.LogDebug(strings.Join([]string{"processed Kafka response message:", resp.String()}, " "))
	return ctx.succeedResponseHandler()
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
		ctx.r = 0
		ctx.inFlightRequests = ctx.inFlightRequests[:0]
	}

	return reqRef
}

func kafkaMessageSize(msgSizeBytes []byte) (int32, error) {
	n := len(msgSizeBytes)
	if n < 4 {
		return 0, errors.New(strings.Join([]string{"expected 4 bytes, but got", strconv.Itoa(n)}, " "))
	}

	return int32(binary.BigEndian.Uint32(msgSizeBytes)), nil
}
