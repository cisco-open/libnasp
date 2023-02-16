package server

import (
	"context"
	"net"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/xtaci/smux"

	"github.com/cisco-open/nasp/pkg/tunnel/api"
)

var (
	defaultLogger         = logr.Discard()
	defaultSessionTimeout = time.Second * 60
)

type server struct {
	listenAddress string

	sessions     sync.Map
	portProvider PortProvider

	logger         logr.Logger
	sessionTimeout time.Duration
}

type ServerOption func(*server)

func ServerWithLogger(logger logr.Logger) ServerOption {
	return func(s *server) {
		s.logger = logger
	}
}

func ServerWithPortProvider(provider PortProvider) ServerOption {
	return func(s *server) {
		s.portProvider = provider
	}
}

func ServerWithSessionTimeout(timeout time.Duration) ServerOption {
	return func(s *server) {
		s.sessionTimeout = timeout
	}
}

func NewServer(listenAddress string, options ...ServerOption) (api.Server, error) {
	s := &server{
		listenAddress:  listenAddress,
		sessions:       sync.Map{},
		sessionTimeout: defaultSessionTimeout,
	}

	for _, option := range options {
		option(s)
	}

	if s.portProvider == nil {
		if pp, err := NewPortProvider(50000, 55000); err != nil {
			return nil, err
		} else {
			s.portProvider = pp
		}
	}

	if s.logger == (logr.Logger{}) {
		s.logger = defaultLogger
	}

	return s, nil
}

func (s *server) Start(ctx context.Context) error {
	s.logger.Info("start server")
	defer s.logger.Info("server stopped")

	l, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return errors.WrapIf(err, "could not listen")
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return errors.WrapIf(err, "could not accept")
		}

		go func() {
			if err := s.handleConn(conn); err != nil {
				s.logger.Error(err, "error during connection handling")
			}
		}()
	}
}

func (s *server) handleConn(conn net.Conn) error {
	sess, err := smux.Server(conn, nil)
	if err != nil {
		return errors.WrapIf(err, "could not create mux server")
	}

	session := NewSession(s, sess)

	s.sessions.Store(conn.RemoteAddr().String(), session)

	defer func() {
		s.logger.V(2).Info("session closed")
		session.Close()
		s.sessions.Delete(conn.RemoteAddr().String())
	}()

	return session.Handle()
}
