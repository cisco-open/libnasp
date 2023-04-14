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
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
	"sync/atomic"
	"syscall"

	"github.com/go-logr/logr"

	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pborman/uuid"

	"github.com/cisco-open/nasp/pkg/ads"
	"github.com/cisco-open/nasp/pkg/istio"
	itcp "github.com/cisco-open/nasp/pkg/istio/tcp"
)

const MAX_WRITE_BUFFERS = 64

type ReadyOps uint32

const (
	OP_READ    ReadyOps = 1 << 0
	OP_WRITE   ReadyOps = 1 << 2
	OP_CONNECT ReadyOps = 1 << 3
	OP_ACCEPT  ReadyOps = 1 << 4
	OP_WAKEUP  ReadyOps = 1 << 5
)

var setupOnce = &sync.Once{}
var integrationHandler *naspIntegrationHandler

func Setup(logLevel int) {
	setupOnce.Do(func() {
		setupLogger(LogLevel(logLevel))
		integrationHandler = newNaspIntegrationHandler()
	})
}

// TCP related structs and functions

type TCPListener struct {
	listener net.Listener

	asyncConnection chan *Connection
	terminated      atomic.Bool

	acceptInProgress atomic.Bool

	logger logr.Logger
}

type naspIntegrationHandler struct {
	iih    istio.IstioIntegrationHandler
	ctx    context.Context
	cancel context.CancelFunc
}

type Connection struct {
	id           string
	conn         net.Conn
	readBuffer   *bytes.Buffer
	readLock     sync.RWMutex
	writeChannel chan []byte

	terminated atomic.Bool

	EOF atomic.Bool

	logger logr.Logger
}

type Address struct {
	Host string
	Port int32
}

// SelectedKey holds the data of a Selected Key formatted on 64 bits as follows:
// |----16bits---|-----------------------------8bits------------|---------8bits--------|----32bits-----|
// |    unused   |read/write running ops status for selected key|selected key ready ops|selected key id|
type SelectedKey uint64

func NewSelectedKey(operation ReadyOps, id uint32) SelectedKey {
	return SelectedKey((uint64(operation) & 0xFF << 32) | uint64(id))
}

func (k SelectedKey) Operation() ReadyOps {
	return ReadyOps(k >> 32 & 0xFF)
}

func (k SelectedKey) SelectedKeyId() uint32 {
	return uint32(k)
}

type Selector struct {
	queue        chan SelectedKey
	writableKeys sync.Map
	readableKeys sync.Map

	readInProgress  sync.Map
	writeInProgress sync.Map
}

func NewSelector() *Selector {
	return &Selector{queue: make(chan SelectedKey, 64)}
}

func (s *Selector) Close() {
	close(s.queue)
	s.writableKeys = sync.Map{}
	s.readableKeys = sync.Map{}
	s.readInProgress = sync.Map{}
	s.writeInProgress = sync.Map{}
}

func (s *Selector) Select(timeoutMs int64) []byte {
	ctx, cancel := context.WithCancel(context.Background())
	if timeoutMs != -1 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	}
	defer cancel()

	//nolint:forcetypeassert
	s.writableKeys.Range(func(key, value any) bool {
		check := value.(func() bool)
		if check() {
			select {
			case s.queue <- NewSelectedKey(OP_WRITE, uint32(key.(int32))):
			default:
				// best effort non-blocking write
			}
		}
		return true
	})

	//nolint:forcetypeassert
	s.readableKeys.Range(func(key, value any) bool {
		check := value.(func() bool)
		if check() {
			select {
			case s.queue <- NewSelectedKey(OP_READ, uint32(key.(int32))):
			default:
				// best effort non-blocking read
			}
		} else if v, _ := s.readInProgress.Load(key.(int32)); v != nil && !v.(bool) {
			s.unregisterReader(key.(int32))
		}
		return true
	})

	if (timeoutMs == 0) && len(s.queue) == 0 {
		return nil
	}

	select {
	case e := <-s.queue:
		selected := make(map[uint32]SelectedKey)
		if e.Operation() != OP_WAKEUP {
			selected[e.SelectedKeyId()] |= e
		}

		l := len(s.queue)

		for i := 0; i < l; i++ {
			e := <-s.queue
			if e.Operation() != OP_WAKEUP {
				selected[e.SelectedKeyId()] |= e
			}
		}

		b := make([]byte, len(selected)*8)
		i := 0
		for k, v := range selected {
			// add current running ops to selected key
			var runningOps uint64
			readInProgress, _ := s.readInProgress.Load(int32(k))
			//nolint:forcetypeassert
			if readInProgress != nil && readInProgress.(bool) {
				runningOps |= uint64(OP_READ)
			}
			writeInProgress, _ := s.writeInProgress.Load(int32(k))
			//nolint:forcetypeassert
			if writeInProgress != nil && writeInProgress.(bool) {
				runningOps |= uint64(OP_WRITE)
			}

			binary.BigEndian.PutUint64(b[i*8:], uint64(v)|(runningOps<<40))
			i++
		}

		return b
	case <-ctx.Done():
		return nil
	}
}

func (s *Selector) registerWriter(selectedKeyId int32, check func() bool) {
	s.writableKeys.Store(selectedKeyId, check)
}

func (s *Selector) unregisterWriter(selectedKeyId int32) {
	s.writableKeys.Delete(selectedKeyId)
	s.writeInProgress.Delete(selectedKeyId)
}

func (s *Selector) registerReader(selectedKeyId int32, check func() bool) {
	s.readableKeys.Store(selectedKeyId, check)
}

func (s *Selector) unregisterReader(selectedKeyId int32) {
	s.readableKeys.Delete(selectedKeyId)
	s.readInProgress.Delete(selectedKeyId)
}

func (s *Selector) WakeUp() {
	s.queue <- NewSelectedKey(OP_WAKEUP, 0)
}

func newNaspIntegrationHandler() *naspIntegrationHandler {
	ctx, cancel := context.WithCancel(logr.NewContext(context.Background(), logger))

	istioHandlerConfig := istio.DefaultIstioIntegrationHandlerConfig

	iih, err := istio.NewIstioIntegrationHandler(&istioHandlerConfig, logger)
	if err != nil {
		cancel()
		panic(err)
	}

	if err := iih.Run(ctx); err != nil {
		cancel()
		panic(err)
	}

	return &naspIntegrationHandler{
		iih:    iih,
		ctx:    ctx,
		cancel: cancel,
	}
}

func CheckAddress(address string) (string, error) {
	ips, err := integrationHandler.iih.GetDiscoveryClient().ResolveHost(address)
	if ads.IgnoreHostNotFound(err) != nil {
		return "", err
	}
	if len(ips) != 0 {
		return ips[0].String(), nil
	}
	return "", nil
}

func Bind(address string, port int) (*TCPListener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}

	listener, err = integrationHandler.iih.GetTCPListener(listener)
	if err != nil {
		return nil, err
	}

	return &TCPListener{
		listener:        listener,
		logger:          logger.WithName("TCPListener"),
		asyncConnection: make(chan *Connection, 1),
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
		l.logger.Error(err, "couldn't close listener", "address", l.listener.Addr())
	}
}

func (l *TCPListener) Accept() (*Connection, error) {
	c, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}

	return &Connection{
		conn:         c,
		readBuffer:   new(bytes.Buffer),
		writeChannel: make(chan []byte, MAX_WRITE_BUFFERS),
		id:           uuid.New(),
		logger:       l.logger.WithName("Connection"),
	}, nil
}

func (l *TCPListener) StartAsyncAccept(selectedKeyId int32, selector *Selector) {
	if !l.acceptInProgress.CompareAndSwap(false, true) {
		return
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				if l.isTerminated() {
					l.acceptInProgress.Store(false)
					break
				}
				println(err.Error())
				continue
			}
			l.asyncConnection <- conn

			selector.queue <- NewSelectedKey(OP_ACCEPT, uint32(selectedKeyId))
		}
	}()
}

func (l *TCPListener) AsyncAccept() *Connection {
	select {
	case conn := <-l.asyncConnection:
		return conn
	default:
		return nil
	}
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
	}

	logger := c.logger.WithValues(logCtx...)
	logger.V(1).Info("StartAsyncRead invoked")

	//nolint:forcetypeassert
	if v, _ := selector.readInProgress.Swap(selectedKeyId, true); v != nil && v.(bool) {
		return
	}

	go func() {
		tempBuffer := make([]byte, 1024)
		for {
			num, err := c.Read(tempBuffer)
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) ||
					errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) {
					c.setEofReceived()
					selector.readInProgress.Store(selectedKeyId, false)
					break
				}

				logger.Error(err, "reading data from connection failed")
				var tlsErr tls.RecordHeaderError
				if errors.As(err, &tlsErr) { // e.g. tls handshake error
					c.setEofReceived()
					selector.readInProgress.Store(selectedKeyId, false)
					break
				}

				continue
			}
			if num > 0 {
				c.readLock.Lock()
				n, err := c.readBuffer.Write(tempBuffer[:num])
				c.readLock.Unlock()
				if err != nil {
					logger.Error(err, "couldn't write data to read buffer")
				}
				if n != num {
					logger.Info("wrote less data into read buffer than the received amount of data!")
				}
			} else {
				logger.V(1).Info("received 0 bytes on connection")
			}

			selector.registerReader(selectedKeyId, func() bool {
				c.readLock.RLock()
				defer c.readLock.RUnlock()
				return c.readBuffer.Len() > 0
			})
		}

		logger.V(1).Info("StartAsyncRead finished")
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

	logger := c.logger.WithValues(logCtx...)
	logger.V(1).Info("StartAsyncWrite invoked")

	//nolint:forcetypeassert
	if v, _ := selector.writeInProgress.Swap(selectedKeyId, true); v != nil && v.(bool) {
		return
	}

	selector.registerWriter(selectedKeyId, func() bool {
		b := len(c.writeChannel) < MAX_WRITE_BUFFERS
		if !b {
			logger.V(1).Info("write channel is full!")
		}
		return b
	})

	go func() {
	out:
		for buff := range c.writeChannel {
			for {
				num, err := c.Write(buff)
				if err != nil {
					if c.isTerminated() {
						selector.writeInProgress.Store(selectedKeyId, false)
						break out
					}
					if errors.Is(err, io.EOF) {
						c.setEofReceived()
						selector.writeInProgress.Store(selectedKeyId, false)
						break out
					}
					logger.Error(err, "received error while writing data to connection")
					continue
				}
				if len(buff) == num {
					break
				} else {
					buff = buff[num:]
				}
			}
		}
		selector.unregisterWriter(selectedKeyId)
		logger.V(1).Info("StartAsyncWrite finished")
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
	if c.EOF.CompareAndSwap(false, true) {
		close(c.writeChannel)
	}
}

func (c *Connection) Close() {
	if !c.terminated.CompareAndSwap(false, true) {
		return // connection already closed
	}

	c.setEofReceived()

	err := c.conn.Close()
	if err != nil {
		c.logger.Error(err, "couldn't close connection")
	}
}

func (c *Connection) AsyncWrite(b []byte) (int32, error) {
	if c.eofReceived() {
		return 0, nil
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
	dialer itcp.Dialer

	asyncConnectionsLock sync.Mutex
	asyncConnections     []*Connection

	logger logr.Logger
}

func NewTCPDialer() (*TCPDialer, error) {
	dialer, err := integrationHandler.iih.GetTCPDialer()
	if err != nil {
		integrationHandler.cancel()
		return nil, err
	}

	return &TCPDialer{
		dialer: dialer,
		logger: logger.WithName("TCPDialer"),
	}, nil
}

func (d *TCPDialer) StartAsyncDial(selectedKeyId int32, selector *Selector, address string, port int) {
	go func() {
		conn, err := d.Dial(address, port)
		if err != nil {
			d.logger.Error(err, "couldn't connect to address", "address", address, "port", port)
			return
		}

		d.asyncConnectionsLock.Lock()
		d.asyncConnections = append(d.asyncConnections, conn)
		d.asyncConnectionsLock.Unlock()

		selector.queue <- NewSelectedKey(OP_CONNECT, uint32(selectedKeyId))
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := d.dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}

	return &Connection{
		conn:         c,
		readBuffer:   new(bytes.Buffer),
		writeChannel: make(chan []byte, MAX_WRITE_BUFFERS),
		id:           uuid.New(),
		logger:       d.logger.WithName("Connection"),
	}, nil
}

// HTTP related structs and functions

type HTTPTransport struct {
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

func NewHTTPTransport() (*HTTPTransport, error) {
	transport, err := integrationHandler.iih.GetHTTPTransport(http.DefaultTransport)
	if err != nil {
		integrationHandler.cancel()
		return nil, err
	}

	client := &http.Client{
		Transport: transport,
	}

	return &HTTPTransport{
		client: client,
		ctx:    integrationHandler.ctx,
		cancel: integrationHandler.cancel,
	}, nil
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

func ListenAndServe(address string, handler HttpHandler) error {
	var err error
	go func() {
		err = integrationHandler.iih.ListenAndServe(integrationHandler.ctx, address, &NaspHttpHandler{handler: handler})
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
