package common

import (
	"encoding/json"
	"io"
	"sync"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/xtaci/smux"

	"github.com/cisco-open/nasp/pkg/tunnel/api"
)

type ctrlStream struct {
	io.ReadWriteCloser

	logger          logr.Logger
	messageHandlers sync.Map
}

func NewControlStream(str *smux.Stream, logger logr.Logger) api.ControlStream {
	s := &ctrlStream{
		ReadWriteCloser: str,

		logger:          logger,
		messageHandlers: sync.Map{},
	}

	return s
}

func (s *ctrlStream) AddMessageHandler(messageType api.MessageType, handler api.MessageHandler) {
	s.messageHandlers.Store(messageType, handler)
}

func (s *ctrlStream) Handle() error {
	s.logger.V(3).Info("handle control stream")

	var msg api.Message

	d := json.NewDecoder(s)
	for {
		if err := d.Decode(&msg); err != nil {
			return errors.WrapIf(err, "could not decode message")
		}

		if err := s.handleMessage(msg); err != nil {
			s.logger.Error(err, "could not handle message")
		}
	}
}

func (s *ctrlStream) handleMessage(msg api.Message) error {
	if v, ok := s.messageHandlers.Load(msg.Type); ok {
		if handler, ok := v.(api.MessageHandler); ok {
			return handler(msg.Data)
		}
	}

	return errors.WithDetails(api.ErrInvalidMessageType, "type", msg.Type)
}
