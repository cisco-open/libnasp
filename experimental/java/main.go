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

package nasp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
	"sync/atomic"

	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cisco-open/nasp/pkg/istio"
	itcp "github.com/cisco-open/nasp/pkg/istio/tcp"
	"github.com/pborman/uuid"

	"k8s.io/klog/v2"
)

const MAX_WRITE_BUFFERS = 64

const (
	OP_READ    = 1 << 0
	OP_WRITE   = 1 << 2
	OP_CONNECT = 1 << 3
	OP_ACCEPT  = 1 << 4
	OP_WAKEUP  = -1
)

var logger = klog.NewKlogr()

// TCP related structs and functions

type TCPListener struct {
	listener net.Listener

	asyncConnectionsLock sync.Mutex
	asyncConnections     []*Connection
	terminated           atomic.Bool
}

type NaspIntegrationHandler struct {
	iih    *istio.IstioIntegrationHandler
	ctx    context.Context
	cancel context.CancelFunc
}

type Connection struct {
	conn         net.Conn
	readBuffer   *bytes.Buffer
	readLock     sync.Mutex
	writeChannel chan []byte
	ctx          context.Context
	cancel       context.CancelFunc

	terminated atomic.Bool

	EOF atomic.Bool
}

type Address struct {
	Host string
	Port int32
}

type SelectedKey struct {
	SelectedKeyId int32
	Operation     int32
}

type Selector struct {
	queue        chan SelectedKey
	selected     map[SelectedKey]struct{}
	writableKeys sync.Map
}

func NewSelector() *Selector {
	return &Selector{queue: make(chan SelectedKey, 64), selected: map[SelectedKey]struct{}{}}
}

func (s *Selector) Close() {
	close(s.queue)
	s.writableKeys = sync.Map{}
	s.selected = nil
}

func (s *Selector) Select(timeoutMs int64) int {
	ctx, cancel := context.WithCancel(context.Background())
	if timeoutMs != -1 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	}
	defer cancel()

	//nolint:forcetypeassert
	s.writableKeys.Range(func(key, value any) bool {
		check := value.(func() bool)
		if check() {
			selectedKey := SelectedKey{
				SelectedKeyId: key.(int32),
				Operation:     OP_WRITE,
			}
			select {
			case s.queue <- selectedKey:
				// TODO: do we miss the default case here?
			}
		}
		return true
	})

	if (timeoutMs == 0) && len(s.queue) == 0 {
		return 0
	}

	select {
	case c := <-s.queue:
		if c.Operation == OP_WAKEUP {
			return 0
		}

		s.selected[c] = struct{}{}

		l := len(s.queue)

		for i := 0; i < l; i++ {
			s.selected[<-s.queue] = struct{}{}
		}

		return l + 1
	case <-ctx.Done():
		return 0
	}
}

func (s *Selector) registerWriter(selectedKeyId int32, check func() bool) {
	s.writableKeys.Store(selectedKeyId, check)
}

func (s *Selector) unregisterWriter(selectedKeyId int32) {
	s.writableKeys.Delete(selectedKeyId)
}

func (s *Selector) WakeUp() {
	s.queue <- SelectedKey{Operation: OP_ACCEPT, SelectedKeyId: -1}
}

func (s *Selector) NextSelectedKey() *SelectedKey {
	for k := range s.selected {
		delete(s.selected, k)
		return &k
	}
	return nil
}

func NewNaspIntegrationHandler(heimdallURL, clientID, clientSecret string) (*NaspIntegrationHandler, error) {
	iih, err := istio.NewIstioIntegrationHandler(&istio.IstioIntegrationHandlerConfig{
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(heimdallURL, clientID, clientSecret, "v1"),
		PushgatewayConfig: &istio.PushgatewayConfig{
			Address:          "push-gw-prometheus-pushgateway.prometheus-pushgateway.svc.cluster.local:9091",
			UseUniqueIDLabel: true,
		},
	}, logger)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go iih.Run(ctx)

	return &NaspIntegrationHandler{
		iih:    iih,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (h *NaspIntegrationHandler) Bind(address string, port int) (*TCPListener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}

	return &TCPListener{
		listener: listener,
	}, nil

	listener, err = h.iih.GetTCPListener(listener)
	if err != nil {
		return nil, err
	}

	return &TCPListener{
		listener: listener,
	}, nil
}

// isTerminated returns whether this TCPListener is terminated or not. This method is thread-safe.
func (l *TCPListener) isTerminated() bool {
	return l.terminated.Load()
}

func (l *TCPListener) Close() {
	if !l.terminated.CompareAndSwap(false, true) {
		return // listener already terminated
	}

	err := l.listener.Close()
	if err != nil {
		logger.Error(err, "couldn't close listener", "address", l.listener.Addr())
	}
}

func (l *TCPListener) Accept() (*Connection, error) {
	c, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), "id", uuid.New()))
	return &Connection{
		conn:         c,
		readBuffer:   new(bytes.Buffer),
		writeChannel: make(chan []byte, MAX_WRITE_BUFFERS),
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

func (l *TCPListener) StartAsyncAccept(selectedKeyId int32, selector *Selector) {
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				if l.isTerminated() {
					break
				}
				println(err.Error())
				continue
			}

			l.asyncConnectionsLock.Lock()
			l.asyncConnections = append(l.asyncConnections, conn)
			l.asyncConnectionsLock.Unlock()

			selector.queue <- SelectedKey{Operation: OP_ACCEPT, SelectedKeyId: selectedKeyId}
		}
	}()
}

func (l *TCPListener) AsyncAccept() *Connection {
	l.asyncConnectionsLock.Lock()
	defer l.asyncConnectionsLock.Unlock()

	if len(l.asyncConnections) > 0 {
		conn := l.asyncConnections[0]
		l.asyncConnections = l.asyncConnections[1:]
		return conn
	}

	return nil
}

func (c *Connection) GetAddress() (*Address, error) {
	address := c.conn.LocalAddr().String()
	port, err := strconv.Atoi(address[strings.LastIndex(address, ":")+1:])
	if err != nil {
		return nil, err
	}
	return &Address{
		Host: address[0:strings.LastIndex(address, ":")],
		//nolint:gosec
		Port: int32(port)}, nil
}

func (c *Connection) GetRemoteAddress() (*Address, error) {
	address := c.conn.RemoteAddr().String()
	port, err := strconv.Atoi(address[strings.LastIndex(address, ":")+1:])
	if err != nil {
		return nil, err
	}
	return &Address{
		Host: address[0:strings.LastIndex(address, ":")],
		//nolint:gosec
		Port: int32(port)}, nil
}

func (c *Connection) StartAsyncRead(selectedKeyId int32, selector *Selector) {
	logCtx := []interface{}{
		"selected key id", selectedKeyId,
		"id", c.ctx.Value("id"),
	}
	logger := logger.WithName("StartAsyncRead")
	logger.Info("Invoked", logCtx...) // log to see if StartAsyncRead is invoked multiple times on the same connection with same selected key id which should not happen !!!
	go func() {
		tempBuffer := make([]byte, 1024)
		for {
			num, err := c.Read(tempBuffer)
			if err != nil {
				if errors.Is(err, io.EOF) || strings.Contains(err.Error(), net.ErrClosed.Error()) {
					c.setEofReceived()
					break
				}
				logger.Error(err, "reading data from connection failed", logCtx...)
				continue
			}
			if num == 0 {
				logger.Info("received 0 bytes on connection", logCtx...)
			}
			c.readLock.Lock()
			n, err := c.readBuffer.Write(tempBuffer[:num])
			c.readLock.Unlock()
			if err != nil {
				logger.Error(err, "couldn't write data to read buffer", logCtx...)
			}
			if n != num {
				logger.Info("wrote less data into read buffer than the received amount of data !!!", logCtx...)
			}
			selector.queue <- SelectedKey{Operation: OP_READ, SelectedKeyId: selectedKeyId}
		}

		logger.Info("StartAsyncRead finished", "id", c.ctx.Value("id"), "selected key id", selectedKeyId)
	}()
}
func (c *Connection) AsyncRead(b []byte) (int32, error) {
	if c.eofReceived() {
		return -1, nil
	}
	c.readLock.Lock()
	defer c.readLock.Unlock()
	max := int(math.Min(float64(c.readBuffer.Len()), float64(len(b))))
	n, err := c.readBuffer.Read(b[:max])
	return int32(n), err
}

func (c *Connection) StartAsyncWrite(selectedKeyId int32, selector *Selector) {
	logCtx := []interface{}{
		"selected key id", selectedKeyId,
	}
	logger := logger.WithName("StartAsyncWrite")

	selector.registerWriter(selectedKeyId, func() bool {
		b := len(c.writeChannel) < MAX_WRITE_BUFFERS
		if !b {
			logger.Info("write channel is full !!!")
		}
		return b
	})

	go func() {
	out:
		for {
			select {
			case buff := <-c.writeChannel:
				for {
					num, err := c.Write(buff)
					if err != nil {
						if c.isTerminated() {
							break out
						}
						if errors.Is(err, io.EOF) {
							c.setEofReceived()
							break out
						}
						logger.Error(err, "received error while writing data to connection", logCtx...)
						continue
					}
					if len(buff) == num {
						break
					} else {
						buff = buff[num:]
					}
				}
			case <-c.ctx.Done():
				logger.Info("CANCEL CONTEXT RECEIVED", "id", c.ctx.Value("id"), "selected key id", selectedKeyId)
				selector.unregisterWriter(selectedKeyId)
				return
			}
		}
		logger.Info("StartAsyncWrite finished", "id", c.ctx.Value("id"), "selected key id", selectedKeyId)
	}()
}

// isTerminated returns whether this Connection is terminated or not. This method is thread-safe.
func (c *Connection) isTerminated() bool {
	return c.terminated.Load()
}

// eofReceived returns true if EOF was received while writing/reading data to/from connection
func (c *Connection) eofReceived() bool {
	return c.EOF.Load()
}

func (c *Connection) setEofReceived() {
	c.EOF.Store(true)
}

func (c *Connection) Close() {
	if !c.terminated.CompareAndSwap(false, true) {
		return // connection already closed
	}

	logger.Info("CLOSING CONNECTION", "id", c.ctx.Value("id"))
	c.cancel()
	err := c.conn.Close()
	if err != nil {
		logger.Error(err, "couldn't close connection")
	}
}

func (c *Connection) AsyncWrite(b []byte) (int32, error) {
	if c.eofReceived() {
		return -1, nil
	}

	bCopy := make([]byte, len(b))
	copy(bCopy, b)
	select {
	case c.writeChannel <- bCopy:
		return int32(len(bCopy)), nil
	default:
		return 0, nil
	}
}

func (c *Connection) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *Connection) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

type TCPDialer struct {
	iih    *istio.IstioIntegrationHandler
	dialer itcp.Dialer
	ctx    context.Context
	cancel context.CancelFunc

	asyncConnectionsLock sync.Mutex
	asyncConnections     []*Connection
}

func NewTCPDialer(heimdallURL, clientID, clientSecret string) (*TCPDialer, error) {
	iih, err := istio.NewIstioIntegrationHandler(&istio.IstioIntegrationHandlerConfig{
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(heimdallURL, clientID, clientSecret, "v1"),
		PushgatewayConfig: &istio.PushgatewayConfig{
			Address:          "push-gw-prometheus-pushgateway.prometheus-pushgateway.svc.cluster.local:9091",
			UseUniqueIDLabel: true,
		},
	}, logger)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{}

	// dialer, err := iih.GetTCPDialer()
	// if err != nil {
	// 	return nil, err
	// }

	ctx, cancel := context.WithCancel(context.Background())
	go iih.Run(ctx)

	return &TCPDialer{
		iih:    iih,
		dialer: dialer,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (d *TCPDialer) StartAsyncDial(selectedKeyId int32, selector *Selector, address string, port int) {
	go func() {
		conn, err := d.Dial(address, port)
		if err != nil {
			println(err.Error())
			return
		}
		if err != nil {
			println(err.Error())
			return
		}

		d.asyncConnectionsLock.Lock()
		d.asyncConnections = append(d.asyncConnections, conn)
		d.asyncConnectionsLock.Unlock()

		selector.queue <- SelectedKey{Operation: OP_CONNECT, SelectedKeyId: selectedKeyId}
	}()
}

func (d *TCPDialer) AsyncDial() *Connection {
	d.asyncConnectionsLock.Lock()
	defer d.asyncConnectionsLock.Unlock()

	if len(d.asyncConnections) > 0 {
		conn := d.asyncConnections[0]
		d.asyncConnections = d.asyncConnections[1:]
		return conn
	}

	return nil
}

func (d *TCPDialer) Dial(address string, port int) (*Connection, error) {
	backgroundCtx := context.WithValue(context.Background(), "id", uuid.New())
	c, err := d.dialer.DialContext(backgroundCtx, "tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(backgroundCtx)
	return &Connection{
		conn:         c,
		readBuffer:   new(bytes.Buffer),
		writeChannel: make(chan []byte, MAX_WRITE_BUFFERS),
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// HTTP related structs and functions

type HTTPTransport struct {
	iih    *istio.IstioIntegrationHandler
	client *http.Client
	ctx    context.Context
	cancel context.CancelFunc
}

type HTTPHeader struct {
	header http.Header
}

func (h *HTTPHeader) Get(key string) string {
	return h.header.Get(key)
}

func (h *HTTPHeader) Set(key, value string) {
	h.header.Set(key, value)
}

type HTTPResponse struct {
	StatusCode int
	Version    string
	Headers    *HTTPHeader
	Body       []byte
}

func NewHTTPTransport(heimdallURL, clientID, clientSecret string) (*HTTPTransport, error) {
	iih, err := istio.NewIstioIntegrationHandler(&istio.IstioIntegrationHandlerConfig{
		UseTLS:              true,
		IstioCAConfigGetter: istio.IstioCAConfigGetterHeimdall(heimdallURL, clientID, clientSecret, "v1"),
		PushgatewayConfig: &istio.PushgatewayConfig{
			Address:          "push-gw-prometheus-pushgateway.prometheus-pushgateway.svc.cluster.local:9091",
			UseUniqueIDLabel: true,
		},
	}, logger)
	if err != nil {
		return nil, err
	}

	transport, err := iih.GetHTTPTransport(http.DefaultTransport)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go iih.Run(ctx)

	client := &http.Client{
		Transport: transport,
	}

	return &HTTPTransport{iih: iih, client: client, ctx: ctx, cancel: cancel}, nil
}

func (t *HTTPTransport) Request(method, url, body string) (*HTTPResponse, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	request, err := http.NewRequestWithContext(context.Background(), method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	response, err := t.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: response.StatusCode,
		Version:    response.Proto,
		Headers:    &HTTPHeader{response.Header},
		Body:       bodyBytes,
	}, nil
}

func (t *HTTPTransport) ListenAndServe(address string, handler HttpHandler) error {
	var err error
	go func() {
		err = t.iih.ListenAndServe(t.ctx, address, &NaspHttpHandler{handler: handler})
	}()
	time.Sleep(1 * time.Second)
	return err
}

func (t *HTTPTransport) Await() {
	<-t.ctx.Done()
}

func (t *HTTPTransport) Close() {
	t.cancel()
}

type HttpHandler interface {
	ServeHTTP(NaspResponseWriter, *NaspHttpRequest)
}

type NaspHttpRequest struct {
	req *http.Request
}

func (r *NaspHttpRequest) Method() string {
	return r.req.Method
}

func (r *NaspHttpRequest) URI() string {
	return r.req.URL.RequestURI()
}

func (r *NaspHttpRequest) Header() []byte {
	hjson, err := json.Marshal(r.req.Header)
	if err != nil {
		panic(err)
	}
	return hjson
}

type Body io.ReadCloser

func (r *NaspHttpRequest) Body() Body {
	return r.req.Body
}

type Header struct {
	h http.Header
}

func (h *Header) Get(key string) string {
	return h.h.Get(key)
}

func (h *Header) Set(key, value string) {
	h.h.Set(key, value)
}

func (h *Header) Add(key, value string) {
	h.h.Add(key, value)
}

type NaspResponseWriter interface {
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
	Header() *Header
}

type naspResponseWriter struct {
	resp http.ResponseWriter
}

func (rw *naspResponseWriter) Write(b []byte) (int, error) {
	return rw.resp.Write(b)
}

func (rw *naspResponseWriter) WriteHeader(statusCode int) {
	rw.resp.WriteHeader(statusCode)
}

func (rw *naspResponseWriter) Header() *Header {
	return &Header{rw.resp.Header()}
}

type NaspHttpHandler struct {
	handler HttpHandler
}

func (h *NaspHttpHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	h.handler.ServeHTTP(&naspResponseWriter{resp}, &NaspHttpRequest{req})
}
