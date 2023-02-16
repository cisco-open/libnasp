package client

import (
	"encoding/json"

	"emperror.dev/errors"
	"github.com/xtaci/smux"

	"github.com/cisco-open/nasp/pkg/tunnel/api"
	"github.com/cisco-open/nasp/pkg/tunnel/common"
)

type ctrlStream struct {
	api.ControlStream

	client *client
}

func NewControlStream(client *client, str *smux.Stream) api.ControlStream {
	cs := common.NewControlStream(str, client.logger)

	s := &ctrlStream{
		ControlStream: cs,
		client:        client,
	}

	cs.AddMessageHandler("addPortResponse", s.addPortResponse)
	cs.AddMessageHandler("requestConnection", s.requestConnection)
	cs.AddMessageHandler("ping", s.ping)

	return s
}

func (s *ctrlStream) ping(msg []byte) error {
	s.client.logger.Info("ping arrived, send pong")

	_, _, err := api.SendMessage(s, api.PongMessageType, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *ctrlStream) requestConnection(msg []byte) error {
	var req api.RequestConnectionMessage
	if err := json.Unmarshal(msg, &req); err != nil {
		return errors.WrapIf(err, "could not unmarshal requestConnection message")
	}

	var mp *managedPort
	if v, ok := s.client.managedPorts.Load(req.Port); ok {
		if p, ok := v.(*managedPort); ok {
			mp = p
		}
	}

	if mp == nil {
		return errors.WithStackIf(api.ErrInvalidPort)
	}

	conn, err := s.client.session.OpenTCPStream(req.Port, req.Identifier)
	if err != nil {
		return errors.WrapIfWithDetails(err, "could not open tcp stream", "port", req.Port, "id", req.Identifier)
	}

	s.client.logger.V(2).Info("put stream into the connection channel", "port", mp.port, "remoteAddress", mp.remoteAddress)

	mp.connChan <- conn

	return nil
}

func (s *ctrlStream) addPortResponse(msg []byte) error {
	var resp api.AddPortResponseMessage
	if err := json.Unmarshal(msg, &resp); err != nil {
		return errors.WrapIf(err, "could not unmarshal addPortResponse message")
	}

	if v, ok := s.client.managedPorts.Load(resp.Port); ok {
		if mp, ok := v.(*managedPort); ok {
			mp.remoteAddress = resp.Address
			mp.initialized = true

			s.client.logger.V(2).Info("port added", "port", resp.Port, "remoteAddress", resp.Address)

			return nil
		}
	}

	return errors.WithStackIf(api.ErrInvalidPort)
}
