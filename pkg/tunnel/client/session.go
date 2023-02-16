package client

import (
	"encoding/json"
	"io"
	"net"

	"emperror.dev/errors"
	"github.com/xtaci/smux"

	"github.com/cisco-open/nasp/pkg/tunnel/api"
)

type session struct {
	session *smux.Session
	client  *client
	stream  api.ControlStream
}

func NewSession(client *client, sess *smux.Session) api.Session {
	return &session{
		session: sess,
		client:  client,
	}
}

func (s *session) Close() error {
	if s.stream != nil {
		if err := s.stream.Close(); err != nil {
			s.client.logger.Error(err, "could not gracefully close control stream")
		}
	}

	return errors.WrapIf(s.session.Close(), "could not gracefully close session")
}

func (s *session) Handle() error {
	if err := s.createControlStream(); err != nil {
		return errors.WrapIf(err, "could not create control stream")
	}

	s.client.onControlStreamConnected()

	return s.stream.Handle()
}

func (s *session) GetControlStream() io.ReadWriter {
	return s.stream
}

func (s *session) OpenTCPStream(port int, id string) (net.Conn, error) {
	stream, err := s.session.OpenStream()
	if err != nil {
		return nil, errors.WrapIf(err, "could not open stream")
	}

	if _, _, err := api.SendMessage(stream, api.OpenTCPStreamMessageType, api.OpenTCPStreamMessage{
		ID: id,
	}); err != nil {
		return nil, errors.WrapIfWithDetails(err, "could not send message", "type", api.OpenTCPStreamMessageType)
	}

	var msg api.Message
	if err := json.NewDecoder(stream).Decode(&msg); err != nil {
		return nil, errors.WrapIf(err, "could not decode message")
	}

	return stream, nil
}

func (s *session) createControlStream() error {
	rawStream, err := s.session.OpenStream()
	if err != nil {
		return errors.WrapIf(err, "could not open stream")
	}

	_, _, err = api.SendMessage(rawStream, api.OpenControlStreamMessageType, nil)
	if err != nil {
		return errors.WrapIfWithDetails(err, "could not send message", "type", api.OpenControlStreamMessageType)
	}

	var msg api.Message
	if err := json.NewDecoder(rawStream).Decode(&msg); err != nil {
		return errors.WrapIf(err, "could not decode message")
	}

	switch msg.Type { //nolint:exhaustive
	case api.StreamOpenedResponseMessageType:
	default:
		return errors.WithStackIf(api.ErrCtrlInvalidResponse)
	}

	s.stream = NewControlStream(s.client, rawStream)

	return nil
}
