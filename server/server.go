package server

import (
	"errors"
	"fmt"

	"github.com/overtalk/qnet/base"
	"github.com/overtalk/qnet/iface"
)

type Server struct {
	stream             bool
	msgRouter          *msgRouter
	sessionManager     *base.SessionManager
	server             iface.IServer
	handler            iface.Handler
	connectHookList    []iface.Handler
	disconnectHookList []iface.Handler
}

type Option func(svr *Server) error

func WithMsgRouter(length base.HeadLength, decoderFunc base.HeadDeserializeFunc) Option {
	return func(svr *Server) error {
		if svr.msgRouter != nil {
			return errors.New("repetitive server msgRouter")
		}

		if svr.handler != nil {
			return errors.New("repetitive server message handler")
		}

		decoder := newMsgRouter(length, decoderFunc)
		svr.msgRouter = decoder
		//svr.handler = decoder.streamMsgHandler
		return nil
	}
}

func WithHandler(h iface.Handler) Option {
	return func(svr *Server) error {
		if svr.handler != nil {
			return errors.New("repetitive server message handler")
		}

		svr.handler = h
		return nil
	}
}

func WithConnectHook(handlers ...iface.Handler) Option {
	return func(svr *Server) error {
		for _, hook := range handlers {
			svr.connectHookList = append(svr.connectHookList, hook)
		}
		return nil
	}
}

func WithDisconnectHook(handlers ...iface.Handler) Option {
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
			svr.stream = true
			// add default tcp hook
			svr.sessionManager = base.NewSessionManager()
			svr.connectHookList = append(svr.connectHookList, svr.defaultConnectedHook)
			svr.disconnectHookList = append(svr.disconnectHookList, svr.defaultDisconnectedHook)
		case base.ProtoTypeUdp:
			s, err := newUdp(ep, svr)
			if err != nil {
				return err
			}
			svr.server = s
		case base.ProtoTypeWs:
			svr.server = newWS(ep, svr)
			svr.sessionManager = base.NewSessionManager()
			svr.connectHookList = append(svr.connectHookList, svr.defaultConnectedHook)
			svr.disconnectHookList = append(svr.disconnectHookList, svr.defaultDisconnectedHook)
		default:
			return fmt.Errorf("invalid net protocol : %s", ep.Proto())
		}

		return nil
	}
}

func NewServer(options ...Option) (*Server, error) {
	svr := new(Server)

	for _, opt := range options {
		if err := opt(svr); err != nil {
			return nil, err
		}
	}

	return svr, nil
}

func (svr *Server) Start() error {
	if svr.handler == nil {
		if svr.msgRouter == nil {
			return errors.New("message handler is nil")
		}

		if svr.stream {
			svr.handler = svr.msgRouter.streamMsgHandler
		} else {
			svr.handler = svr.msgRouter.packetMsgHandler
		}
	}

	if svr.server == nil {
		return errors.New("server is nil")
	}

	return svr.server.Start()
}

func (svr *Server) Stop() {
	svr.server.Stop()
	svr.sessionManager.ClearSession()
}

func (svr *Server) RegisterMsgHandler(id uint16, handler iface.MsgHandler) error {
	if svr.msgRouter == nil {
		return errors.New("msgRouter is absent")
	}

	return svr.msgRouter.registerMsgHandler(id, handler)
}

func (svr *Server) SendBySessionID(sessionID uint64, data []byte) (int, error) {
	session, err := svr.sessionManager.Get(sessionID)
	if err != nil {
		return 0, err
	}

	return session.Write(data)
}

func (svr *Server) SetSessionMeta(sessionID uint64, key string, value interface{}) error {
	session, err := svr.sessionManager.Get(sessionID)
	if err != nil {
		return err
	}

	session.SetMeta(key, value)
	return nil
}

func (svr *Server) GetSessionMeta(sessionID uint64, key string) (interface{}, error) {
	session, err := svr.sessionManager.Get(sessionID)
	if err != nil {
		return nil, err
	}

	return session.GetMeta(key)
}

// ----------------------- hook ---------------------------
func (svr *Server) defaultConnectedHook(session base.Session) {
	svr.sessionManager.Add(session)
}

func (svr *Server) defaultDisconnectedHook(session base.Session) {
	session.Close()
	svr.sessionManager.Remove(session.GetSessionID())
}
