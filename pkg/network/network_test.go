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

//nolint:goerr113,noctx
package network_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/textproto"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cisco-open/libnasp/pkg/ca"
	"github.com/cisco-open/libnasp/pkg/ca/selfsigned"
	"github.com/cisco-open/libnasp/pkg/network"
	"github.com/cisco-open/libnasp/pkg/network/listener"
	"github.com/cisco-open/libnasp/pkg/network/pool"
)

type NetworkTestSuite struct {
	suite.Suite

	caClient ca.Client
	certPool *x509.CertPool

	simpleTCPServerRunning bool
	simpleTCPServerAddr    string

	wrappedTCPServerRunning bool
	wrappedTCPServerAddr    string

	wrappedTLSServerRunning bool
	wrappedTLSServerAddr    string
	wrappedTLServerCertURI  string

	wrappedHTTPServerRunning bool
	wrappedHTTPServerAddr    string

	wrappedHTTPSServerRunning bool
	wrappedHTTPSServerAddr    string
	wrappedHTTPSServerCertURI string
}

func (s *NetworkTestSuite) SetupTest() {
	caClient, err := selfsigned.NewSelfSignedCAClient(selfsigned.WithKeySize(2048))
	s.Require().Nil(err)

	s.caClient = caClient

	s.certPool = x509.NewCertPool()
	s.certPool.AppendCertsFromPEM(caClient.GetCAPem())

	s.simpleTCPServerAddr = fmt.Sprintf("127.0.0.1:%d", randPort())

	s.wrappedTCPServerAddr = fmt.Sprintf("127.0.0.1:%d", randPort())

	s.wrappedTLSServerAddr = fmt.Sprintf("127.0.0.1:%d", randPort())
	s.wrappedTLServerCertURI = "spiffe://acme.corp/test-tls-server"

	s.wrappedHTTPServerAddr = fmt.Sprintf("127.0.0.1:%d", randPort())

	s.wrappedHTTPSServerAddr = fmt.Sprintf("127.0.0.1:%d", randPort())
	s.wrappedHTTPSServerCertURI = "spiffe://acme.corp/test-https-server"
}

func (s *NetworkTestSuite) TestSelfsignedCertificate() {
	ttl := time.Hour * 24

	tlsCert, err := s.caClient.GetCertificate("test.example.com", ttl)
	s.Require().Nil(err)
	s.Require().Implements((*ca.Certificate)(nil), tlsCert)

	cert, err := network.ParseTLSCertificate(tlsCert.GetTLSCertificate())
	s.Require().Nil(err)
	s.Require().Implements((*network.Certificate)(nil), cert)

	s.Require().Equal(ttl, cert.GetCertificate().NotAfter.Sub(cert.GetCertificate().NotBefore))
}

func (s *NetworkTestSuite) simpleTCPServer(ctx context.Context) error {
	if s.simpleTCPServerRunning {
		return errors.New("simple tcp server already running")
	}

	defer func() {
		s.simpleTCPServerRunning = false
	}()
	l, err := net.Listen("tcp", s.simpleTCPServerAddr)
	if err != nil {
		return err
	}

	s.simpleTCPServerRunning = true

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			c, err := l.Accept()
			if err != nil {
				return err
			}
			go s.handleTCPConnection(c)
		}
	}
}

func (s *NetworkTestSuite) wrappedTCPServer(ctx context.Context, handler func(net.Conn)) error {
	if s.wrappedTCPServerRunning {
		return errors.New("wrapped tcp server already running")
	}

	defer func() {
		s.wrappedTCPServerRunning = false
	}()

	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.wrappedTCPServerAddr)
	if err != nil {
		return err
	}

	l = network.WrappedConnectionListener(l)

	s.wrappedTCPServerRunning = true

	go func() {
		for {
			<-ctx.Done()
			l.Close()
			return
		}
	}()

	for {
		c, err := l.Accept()
		if errors.Is(err, net.ErrClosed) {
			return nil
		}
		if err != nil {
			return err
		}
		go handler(c)
	}
}

func (s *NetworkTestSuite) wrappedTLSServer(ctx context.Context, handler func(net.Conn)) error {
	if s.wrappedTLSServerRunning {
		return errors.New("wrapped tls server already running")
	}

	defer func() {
		s.wrappedTLSServerRunning = false
	}()

	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.wrappedTLSServerAddr)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequestClientCert,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := s.caClient.GetCertificate(s.wrappedTLServerCertURI, time.Duration(168)*time.Hour)
			if err != nil {
				return nil, err
			}

			return cert.GetTLSCertificate(), nil
		},
		RootCAs:            s.certPool,
		InsecureSkipVerify: true,
	}

	l = listener.NewUnifiedListener(l, network.WrapTLSConfig(tlsConfig), listener.TLSModePermissive,
		listener.UnifiedListenerWithTLSConnectionCreator(network.CreateTLSServerConn),
		listener.UnifiedListenerWithConnectionWrapper(func(c net.Conn) net.Conn {
			return network.WrapConnection(c)
		}),
	)

	s.wrappedTLSServerRunning = true

	go func() {
		for {
			<-ctx.Done()
			l.Close()
			return
		}
	}()

	for {
		c, err := l.Accept()
		if errors.Is(err, net.ErrClosed) {
			return nil
		}
		if err != nil {
			return err
		}
		go handler(c)
	}
}

func (s *NetworkTestSuite) handleTCPConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if !errors.Is(err, io.EOF) {
				s.Require().Nil(err)
			}
			return
		}

		line := fmt.Sprintf("%s %s", "<< ", bytes)
		_, err = conn.Write([]byte(line))
		s.Require().Nil(err)
	}
}

func (s *NetworkTestSuite) wrappedHTTPServer(ctx context.Context, handler http.Handler) error {
	if s.wrappedHTTPServerRunning {
		return errors.New("wrapped http server already running")
	}

	defer func() {
		s.wrappedHTTPServerRunning = false
	}()

	server := network.WrapHTTPServer(&http.Server{
		Handler: handler,
	})

	l, err := net.Listen("tcp", s.wrappedHTTPServerAddr)
	if err != nil {
		return err
	}

	s.wrappedHTTPServerRunning = true

	go func() {
		err := server.Serve(l)
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		s.Require().Nil(err)
	}()

	go func() {
		<-ctx.Done()
		err := server.GetHTTPServer().Shutdown(context.Background())
		s.Require().Nil(err)
	}()

	<-ctx.Done()

	return nil
}

func (s *NetworkTestSuite) wrappedHTTPSServer(ctx context.Context, handler http.Handler) error {
	if s.wrappedHTTPSServerRunning {
		return errors.New("wrapped https server already running")
	}

	defer func() {
		s.wrappedHTTPSServerRunning = false
	}()

	server := network.WrapHTTPServer(&http.Server{
		Handler: handler,
	})

	l, err := net.Listen("tcp", s.wrappedHTTPSServerAddr)
	if err != nil {
		return err
	}

	s.wrappedHTTPSServerRunning = true

	go func() {
		err := server.ServeWithTLSConfig(l, &tls.Config{
			ClientAuth: tls.RequestClientCert,
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				cert, err := s.caClient.GetCertificate(s.wrappedHTTPSServerCertURI, time.Duration(168)*time.Hour)
				if err != nil {
					return nil, err
				}

				return cert.GetTLSCertificate(), nil
			},
			RootCAs:            s.certPool,
			InsecureSkipVerify: true,
			NextProtos:         []string{"h2"},
		})
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		s.Require().Nil(err)
	}()
	go func() {
		<-ctx.Done()
		err := server.GetHTTPServer().Shutdown(context.Background())
		s.Require().Nil(err)
	}()

	<-ctx.Done()

	return nil
}

func (s *NetworkTestSuite) TestWrappedHTTPServerWithSimpleClient() {
	ctx, cancelContext := context.WithCancel(context.Background())
	defer cancelContext()

	body := []byte("hello world")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		connectionState, ok := network.ConnectionStateFromContext(r.Context())
		s.Require().True(ok)
		s.Require().NotNil(connectionState)
		s.Equal(s.wrappedHTTPServerAddr, connectionState.LocalAddr().String())
		connectionState.ResetTimeToFirstByte()

		_, err := w.Write(body)
		s.Require().Nil(err)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		s.WithinRange(connectionState.GetTimeToFirstByte(), startTime, time.Now().Add(time.Second))
	})

	go func() {
		err := s.wrappedHTTPServer(ctx, mux)
		s.Require().Nil(err)
	}()

	s.Eventually(func() bool {
		return s.wrappedHTTPServerRunning
	}, time.Second*5, time.Millisecond*10, "wrapped http server didn't come up")

	c := http.Client{}

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := c.Get("http://" + s.wrappedHTTPServerAddr)
			s.Require().Nil(err)
			defer resp.Body.Close()
			s.Require().NotNil(resp)
			s.Equal(resp.StatusCode, 200)

			b, err := io.ReadAll(resp.Body)
			s.Require().Nil(err)
			s.Equal(b, body)
		}()
	}
	wg.Wait()
}

func (s *NetworkTestSuite) TestWrappedHTTPServerWithWrappedClient() {
	ctx, cancelContext := context.WithCancel(context.Background())
	defer cancelContext()

	body := []byte("hello world")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		connectionState, ok := network.ConnectionStateFromContext(r.Context())
		s.Require().True(ok)
		s.Require().NotNil(connectionState)
		s.Equal(s.wrappedHTTPServerAddr, connectionState.LocalAddr().String())
		connectionState.ResetTimeToFirstByte()

		_, err := w.Write(body)
		s.Require().Nil(err)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		s.WithinRange(connectionState.GetTimeToFirstByte(), startTime, time.Now().Add(time.Second))
	})

	go func() {
		err := s.wrappedHTTPServer(ctx, mux)
		s.Require().Nil(err)
	}()

	s.Eventually(func() bool {
		return s.wrappedHTTPServerRunning
	}, time.Second*5, time.Millisecond*10, "wrapped http server didn't come up")

	c := http.Client{ //nolint:forcetypeassert
		Transport: network.WrapHTTPTransport(http.DefaultTransport.(*http.Transport).Clone(), network.NewDialer()),
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			startTime := time.Now()

			defer wg.Done()
			resp, err := c.Get("http://" + s.wrappedHTTPServerAddr)
			s.Require().Nil(err)
			defer resp.Body.Close()
			s.Require().NotNil(resp)
			s.Equal(resp.StatusCode, 200)

			b, err := io.ReadAll(resp.Body)
			s.Require().Nil(err)
			s.Equal(b, body)

			connectionState, ok := network.ConnectionStateFromContext(resp.Request.Context())
			s.Require().True(ok)
			s.Require().NotNil(connectionState)
			s.Equal(s.wrappedHTTPServerAddr, connectionState.RemoteAddr().String())
			s.WithinRange(connectionState.GetTimeToFirstByte(), startTime, time.Now().Add(time.Second))
		}()
	}
	wg.Wait()
}

func (s *NetworkTestSuite) TestWrappedHTTPSServerWithWrappedClient() {
	ctx, cancelContext := context.WithCancel(context.Background())
	defer cancelContext()

	body := []byte("hello world")
	uri := "spiffe://acme.corp/https-wrapped-client"

	type Stats struct {
		Count                  int
		HandshakeCompleteCount int
		Addresses              map[string]int
	}

	serverStats := &Stats{
		Addresses: make(map[string]int),
	}
	clientStats := &Stats{
		Addresses: make(map[string]int),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		connectionState, ok := network.ConnectionStateFromContext(r.Context())
		s.Require().True(ok)
		s.Require().NotNil(connectionState)
		s.Equal(s.wrappedHTTPSServerAddr, connectionState.LocalAddr().String())
		connectionState.ResetTimeToFirstByte()

		_, err := w.Write(body)
		s.Require().Nil(err)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		s.WithinRange(connectionState.GetTimeToFirstByte(), startTime, time.Now().Add(time.Second))

		s.Require().NotNil(connectionState.GetPeerCertificate())
		s.Equal(uri, connectionState.GetPeerCertificate().GetFirstURI())

		s.Require().NotNil(connectionState.GetLocalCertificate())
		s.Equal(s.wrappedHTTPSServerCertURI, connectionState.GetLocalCertificate().GetFirstURI())

		serverStats.Count++
		if connectionState.GetTLSConnectionState().HandshakeComplete {
			serverStats.HandshakeCompleteCount++
		}
		serverStats.Addresses[connectionState.RemoteAddr().String()]++
	})

	go func() {
		err := s.wrappedHTTPSServer(ctx, mux)
		s.Require().Nil(err)
	}()

	s.Eventually(func() bool {
		return s.wrappedHTTPSServerRunning
	}, time.Second*5, time.Millisecond*10, "wrapped http server didn't come up")

	d := network.NewDialerWithTLSConfig(&tls.Config{
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := s.caClient.GetCertificate(uri, time.Duration(168)*time.Hour)
			if err != nil {
				return nil, err
			}

			return cert.GetTLSCertificate(), nil
		},
		RootCAs:            s.certPool,
		InsecureSkipVerify: true,
	})

	var t *http.Transport
	if tp, ok := http.DefaultTransport.(*http.Transport); ok {
		t = tp.Clone()
	}
	s.Require().NotNil(t)
	t.MaxIdleConns = 10
	t.MaxConnsPerHost = 10
	t.MaxIdleConnsPerHost = 10

	c := http.Client{
		Transport: network.WrapHTTPTransport(t, d),
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			startTime := time.Now()

			defer wg.Done()
			resp, err := c.Get("https://" + s.wrappedHTTPSServerAddr + "/")
			s.Require().Nil(err)
			defer resp.Body.Close()
			s.Require().NotNil(resp)
			s.Equal(200, resp.StatusCode)

			b, err := io.ReadAll(resp.Body)
			s.Require().Nil(err)
			s.Equal(body, b)

			connectionState, ok := network.ConnectionStateFromContext(resp.Request.Context())
			s.Require().True(ok)
			s.Require().NotNil(connectionState)
			s.Equal(s.wrappedHTTPSServerAddr, connectionState.RemoteAddr().String())
			s.WithinRange(connectionState.GetTimeToFirstByte(), startTime, time.Now().Add(time.Second))

			s.Require().NotNil(connectionState.GetLocalCertificate())
			s.Equal(uri, connectionState.GetLocalCertificate().GetFirstURI())

			s.Require().NotNil(connectionState.GetPeerCertificate())
			s.Equal(s.wrappedHTTPSServerCertURI, connectionState.GetPeerCertificate().GetFirstURI())

			clientStats.Count++
			if connectionState.GetTLSConnectionState().HandshakeComplete {
				clientStats.HandshakeCompleteCount++
			}
			clientStats.Addresses[connectionState.LocalAddr().String()]++
		}()
	}
	wg.Wait()

	s.Require().Equal(serverStats.Count, clientStats.Count)
	s.Require().Equal(serverStats.HandshakeCompleteCount, clientStats.HandshakeCompleteCount)
	s.Require().Equal(serverStats.Addresses, clientStats.Addresses)
}

func (s *NetworkTestSuite) TestWrappedTLSServer() { //nolint:funlen
	ctx, cancelContext := context.WithCancel(context.Background())

	uri := "spiffe://acme.corp/tls-client"

	type Stats struct {
		Count                  int
		HandshakeCompleteCount int
		Addresses              map[string]int
	}

	serverMux := sync.Mutex{}
	clientMux := sync.Mutex{}

	serverStats := &Stats{
		Addresses: make(map[string]int),
	}
	clientStats := &Stats{
		Addresses: make(map[string]int),
	}

	go func() {
		err := s.wrappedTLSServer(ctx, func(conn net.Conn) {
			for {
				startTime := time.Now()

				connectionState, ok := network.ConnectionStateFromNetConn(conn)
				s.Require().True(ok)
				s.Require().NotNil(connectionState)
				connectionState.ResetTimeToFirstByte()

				r := textproto.NewReader(bufio.NewReader(conn))
				str, err := r.ReadLine()
				s.Require().Nil(err)
				s.Equal("hello", str)

				s.Equal(s.wrappedTLSServerAddr, connectionState.LocalAddr().String())
				s.WithinRange(connectionState.GetTimeToFirstByte(), startTime, time.Now().Add(time.Second))

				s.Require().NotNil(connectionState.GetPeerCertificate())
				s.Equal(uri, connectionState.GetPeerCertificate().GetFirstURI())

				s.Require().NotNil(connectionState.GetLocalCertificate())
				s.Equal(s.wrappedTLServerCertURI, connectionState.GetLocalCertificate().GetFirstURI())

				serverStats.Count++
				if connectionState.GetTLSConnectionState().HandshakeComplete {
					serverStats.HandshakeCompleteCount++
				}
				serverMux.Lock()
				serverStats.Addresses[connectionState.RemoteAddr().String()]++
				serverMux.Unlock()
			}
		})

		s.Require().Nil(err)
	}()

	s.Eventually(func() bool {
		return s.wrappedTLSServerRunning
	}, time.Second*5, time.Millisecond*10, "wrapped tls server didn't come up")

	d := network.NewDialerWithTLSConfig(&tls.Config{
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := s.caClient.GetCertificate(uri, time.Duration(168)*time.Hour)
			if err != nil {
				return nil, err
			}

			return cert.GetTLSCertificate(), nil
		},
		RootCAs:            s.certPool,
		InsecureSkipVerify: true,
	})

	p, err := pool.NewChannelPool(func() (net.Conn, error) {
		conn, err := d.DialContext(ctx, "tcp", s.wrappedTLSServerAddr)
		if connectionState, ok := network.ConnectionStateFromNetConn(conn); ok {
			connectionState.ResetTimeToFirstByte()
		}

		return conn, err
	}, pool.ChannelPoolWithInitialCap(5), pool.ChannelPoolWithMaxCap(5))
	s.Require().Nil(err)

	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			startTime := time.Now()
			defer wg.Done()

			conn, err := p.Get()
			defer func() {
				_ = conn.Close()
			}()
			s.Require().Nil(err)

			connectionState, ok := network.ConnectionStateFromNetConn(conn)
			s.Require().True(ok)
			s.Require().NotNil(connectionState)
			connectionState.ResetTimeToFirstByte()

			_, err = conn.Write([]byte("hello\n"))
			s.Require().Nil(err)

			s.Equal(s.wrappedTLSServerAddr, connectionState.RemoteAddr().String())
			s.WithinRange(connectionState.GetTimeToFirstByte(), startTime, time.Now().Add(time.Second))

			s.Require().NotNil(connectionState.GetPeerCertificate())
			s.Equal(s.wrappedTLServerCertURI, connectionState.GetPeerCertificate().GetFirstURI())

			s.Require().NotNil(connectionState.GetLocalCertificate())
			s.Equal(uri, connectionState.GetLocalCertificate().GetFirstURI())

			clientStats.Count++
			if connectionState.GetTLSConnectionState().HandshakeComplete {
				clientStats.HandshakeCompleteCount++
			}
			clientMux.Lock()
			clientStats.Addresses[connectionState.LocalAddr().String()]++
			clientMux.Unlock()
		}()

		if i%5 == 0 {
			time.Sleep(time.Millisecond * 100)
		}
	}
	wg.Wait()

	cancelContext()

	// wait for tls server to shut down to give time to server connectionState handler to act
	s.Eventually(func() bool {
		return !s.wrappedTLSServerRunning
	}, time.Second*5, time.Millisecond*10, "wrapped tls server didn't shut down")

	s.Require().Equal(serverStats.Count, clientStats.Count)
	s.Require().Equal(serverStats.HandshakeCompleteCount, clientStats.HandshakeCompleteCount)
	s.Require().Equal(serverStats.Addresses, clientStats.Addresses)
}

func (s *NetworkTestSuite) TestWrappedTCPServer() {
	ctx, cancelContext := context.WithCancel(context.Background())

	go func() {
		err := s.wrappedTCPServer(ctx, func(conn net.Conn) {
			startTime := time.Now()

			b := make([]byte, 1)
			n, err := conn.Read(b)
			s.Require().Nil(err)

			s.Equal(1, n)

			connectionState, ok := network.ConnectionStateFromNetConn(conn)
			s.Require().True(ok)
			s.Require().NotNil(connectionState)

			s.Equal(s.wrappedTCPServerAddr, connectionState.LocalAddr().String())
			s.WithinRange(connectionState.GetTimeToFirstByte(), startTime, time.Now().Add(time.Second))
		})
		s.Require().Nil(err)
	}()

	s.Eventually(func() bool {
		return s.wrappedTCPServerRunning
	}, time.Second*5, time.Millisecond*10, "wrapped tcp server didn't come up")

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			startTime := time.Now()

			defer wg.Done()

			conn, err := network.NewDialer().DialContext(ctx, "tcp", s.wrappedTCPServerAddr)
			s.Require().Nil(err)

			_, err = conn.Write([]byte("hello\n"))
			s.Require().Nil(err)

			connectionState, ok := network.ConnectionStateFromNetConn(conn)
			s.Require().True(ok)
			s.Require().NotNil(connectionState)

			s.Equal(s.wrappedTCPServerAddr, connectionState.RemoteAddr().String())
			s.WithinRange(connectionState.GetTimeToFirstByte(), startTime, time.Now().Add(time.Second))
		}()
	}
	wg.Wait()

	cancelContext()

	// wait for tcp server to shut down to give time to server connectionState handler to act
	s.Eventually(func() bool {
		return !s.wrappedTCPServerRunning
	}, time.Second*5, time.Millisecond*10, "wrapped tcp server didn't shut down")
}

func (s *NetworkTestSuite) TestSimpleTCPClient() {
	ctx, cf := context.WithCancel(context.Background())
	defer cf()

	go func() {
		err := s.simpleTCPServer(ctx)
		s.Require().Nil(err)
	}()

	s.Eventually(func() bool {
		return s.simpleTCPServerRunning
	}, time.Second*5, time.Millisecond*10, "simple tcp server didn't come up")

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			startTime := time.Now()

			defer wg.Done()

			conn, err := network.NewDialer().DialContext(ctx, "tcp", s.simpleTCPServerAddr)
			s.Require().Nil(err)

			_, err = conn.Write([]byte("hello\n"))
			s.Require().Nil(err)

			reply := make([]byte, 1024)
			n, err := conn.Read(reply)
			s.Require().Nil(err)

			s.Greater(n, 0)

			connectionState, ok := network.ConnectionStateFromNetConn(conn)
			s.Require().True(ok)
			s.Require().NotNil(connectionState)

			s.Equal(s.simpleTCPServerAddr, connectionState.RemoteAddr().String())
			s.WithinRange(connectionState.GetTimeToFirstByte(), startTime, time.Now().Add(time.Second))
		}()
	}
	wg.Wait()
}

func TestNetworkTestSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(NetworkTestSuite))
}

//nolint:gosec
func randPort() int {
	min := 50000
	max := 65535

	rand.Seed(time.Now().UnixNano())

	return rand.Intn(max-min+1) + min
}
