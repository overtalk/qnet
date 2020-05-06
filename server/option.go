package server

import (
	"errors"
	"fmt"

	"github.com/overtalk/qnet/base"
)

type Option func(svr *Server) error

func WithDecoder(length base.HeadLength, decoderFunc base.HeadDeserializeFunc) Option {
	return func(svr *Server) error {
		if svr.decoder != nil {
			return errors.New("repetitive server decoder")
		}

		if svr.handler != nil {
			return errors.New("repetitive server message handler")
		}

		decoder := newDecoder(length, decoderFunc)
		svr.decoder = decoder
		svr.handler = decoder.handler
		return nil
	}
}

func WithHandler(h Handler) Option {
	return func(svr *Server) error {
		if svr.handler != nil {
			return errors.New("repetitive server message handler")
		}

		svr.handler = h
		return nil
	}
}

func WithConnectHook(handlers ...Handler) Option {
	return func(svr *Server) error {
		for _, hook := range handlers {
			svr.connectHookList = append(svr.connectHookList, hook)
		}
		return nil
	}
}

func WithDisconnectHook(handlers ...Handler) Option {
	return func(svr *Server) error {
		for _, hook := range handlers {
			svr.disconnectHookList = append(svr.disconnectHookList, hook)
		}
		return nil
	}
}

func WithURL(url string) Option {
	return func(svr *Server) error {
		if svr.server != nil {
			return errors.New("repetitive server endpoint")
		}

		ep, err := base.NewFromString(url)
		if err != nil {
			return err
		}

		return WithEndPoint(ep)(svr)
	}
}

func WithEndPoint(ep *base.Endpoint) Option {
	return func(svr *Server) error {
		if svr.server != nil {
			return errors.New("repetitive server endpoint")
		}

		switch ep.Proto() {
		case base.ProtoTypeTcp:
			s, err := newTcp(ep, svr)
			if err != nil {
				return err
			}
			svr.server = s
			// add default tcp hook
			svr.sessionManager = newSessionManager()
			svr.connectHookList = append(svr.connectHookList, svr.defaultTcpConnectedHook)
			svr.disconnectHookList = append(svr.disconnectHookList, svr.defaultTcpDisconnectedHook)
		case base.ProtoTypeUdp:
			s, err := newUdp(ep, svr)
			if err != nil {
				return err
			}
			svr.server = s
		case base.ProtoTypeWs:
		default:
			return fmt.Errorf("invalid net protocol : %s", ep.Proto())
		}

		return nil
	}
}
