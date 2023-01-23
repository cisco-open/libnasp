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
	"io"
	"net"

	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cisco-open/nasp/pkg/istio"
	itcp "github.com/cisco-open/nasp/pkg/istio/tcp"

	"k8s.io/klog/v2"
)

const (
	OP_READ    = 1 << 0
	OP_WRITE   = 1 << 2
	OP_CONNECT = 1 << 3
	OP_ACCEPT  = 1 << 4
)

var logger = klog.NewKlogr()

// TCP related structs and functions

type TCPListener struct {
	iih      *istio.IstioIntegrationHandler
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc

	asyncConnectionsLock sync.Mutex
	asyncConnections     []*Connection
}

type Connection struct {
	conn         net.Conn
	readBuffer   *bytes.Buffer
	readLock     sync.Mutex
	writeChannel chan []byte
}

type SelectedKey struct {
	SelectedKeyId int32
	Operation     int32
}

type Selector struct {
	queue    chan *SelectedKey
	selected []*SelectedKey
}

func NewSelector() *Selector {
	return &Selector{queue: make(chan *SelectedKey, 64)}
}

func (s *Selector) Select(timeout int64) int {
	timeoutChan := make(<-chan time.Time)
	if timeout != -1 {
		timeoutChan = time.After(time.Duration(timeout) * time.Millisecond)
	}
	select {
	case c := <-s.queue:
		s.selected = append(s.selected, c)

		l := len(s.queue)

		for i := 0; i < l; i++ {
			s.selected = append(s.selected, <-s.queue)
		}

		return l + 1
	case <-timeoutChan:
		return 0
	}
}

func (s *Selector) NextSelectedKey() *SelectedKey {
	if len(s.selected) != 0 {
		selected := s.selected[0]
		s.selected = s.selected[1:]
		return selected
	}
	return nil
}

func NewTCPListener(heimdallURL, clientID, clientSecret string) (*TCPListener, error) {
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

	listener, err := net.Listen("tcp", "localhost:10000")
	if err != nil {
		return nil, err
	}

	listener, err = iih.GetTCPListener(listener)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go iih.Run(ctx)

	return &TCPListener{
		iih:      iih,
		listener: listener,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (l *TCPListener) Accept() (*Connection, error) {
	c, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &Connection{conn: c, readBuffer: new(bytes.Buffer), writeChannel: make(chan []byte, 64)}, nil
}

func (l *TCPListener) StartAsyncAccept(selectedKeyId int32, selector *Selector) {
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				println(err.Error())
				continue
			}

			l.asyncConnectionsLock.Lock()
			l.asyncConnections = append(l.asyncConnections, conn)
			l.asyncConnectionsLock.Unlock()

			selector.queue <- &SelectedKey{Operation: OP_ACCEPT, SelectedKeyId: selectedKeyId}
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

func (c *Connection) StartAsyncRead(selectedKeyId int32, selector *Selector) {
	go func() {
		tempBuffer := make([]byte, 1024)
		for {
			num, err := c.Read(tempBuffer)
			if err != nil {
				if err == io.EOF {
					break
				}
				println("Error received:")
				println(err.Error())
				continue
			}
			c.readLock.Lock()
			c.readBuffer.Write(tempBuffer[:num])
			c.readLock.Unlock()
			selector.queue <- &SelectedKey{Operation: OP_READ, SelectedKeyId: selectedKeyId}
		}
	}()
}
func (c *Connection) AsyncRead(b []byte) (int32, error) {
	c.readLock.Lock()
	defer c.readLock.Unlock()
	bufferLen := c.readBuffer.Len()
	//TODO b cap nehogy tobbett olvassak
	if bufferLen > 0 {
		copy(b, c.readBuffer.Bytes())
	}
	return int32(bufferLen), nil
}

func (c *Connection) StartAsyncWrite(selectedKeyId int32, selector *Selector) {
	go func() {
		for {
			select {
			case buff := <-c.writeChannel:
				for {
					num, err := c.Write(buff)
					if err != nil {
						println(err.Error())
						continue
					}
					if len(buff) == num {
						break
					}
					// } else {
					// 	buff = buff[num:]
					// }
				}
			}
			selector.queue <- &SelectedKey{Operation: OP_WRITE, SelectedKeyId: selectedKeyId}
		}
	}()
}
func (c *Connection) AsyncWrite(b []byte) (int32, error) {
	c.writeChannel <- b
	return int32(len(b)), nil
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

	dialer, err := iih.GetTCPDialer()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go iih.Run(ctx)

	return &TCPDialer{
		iih:    iih,
		dialer: dialer,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (d *TCPDialer) Dial() (*Connection, error) {
	c, err := d.dialer.DialContext(context.Background(), "tcp", "localhost:10000")
	if err != nil {
		return nil, err
	}
	return &Connection{conn: c}, nil
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
