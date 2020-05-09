package qnet

import (
	"errors"
	"fmt"
)

type Server struct {
	protocolType       ProtoType
	msgRouter          *msgRouter
	sessionManager     *SessionManager
	server             IServer
	handler            SessionHandler
	connectHookList    []SessionHandler
	disconnectHookList []SessionHandler
}

type Option func(svr *Server) error

func WithMsgRouter(length HeadLength, decoderFunc HeadDeserializeFunc, headSerializeFunc HeadSerializeFunc) Option {
	return func(svr *Server) error {
		if svr.msgRouter != nil {
			return errors.New("repetitive server msgRouter")
		}

		if svr.handler != nil {
			return errors.New("repetitive server message handler")
		}

		decoder := newMsgRouter(length, decoderFunc, headSerializeFunc)
		svr.msgRouter = decoder
		//svr.handler = decoder.streamMsgHandler
		return nil
	}
}

func WithHandler(h SessionHandler) Option {
	return func(svr *Server) error {
		if svr.handler != nil {
			return errors.New("repetitive server message handler")
		}

		svr.handler = h
		return nil
	}
}

func WithConnectHook(handlers ...SessionHandler) Option {
	return func(svr *Server) error {
		for _, hook := range handlers {
			svr.connectHookList = append(svr.connectHookList, hook)
		}
		return nil
	}
}

func WithDisconnectHook(handlers ...SessionHandler) Option {
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

		ep, err := NewFromString(url)
		if err != nil {
			return err
		}

		return WithEndPoint(ep)(svr)
	}
}

func WithEndPoint(ep *Endpoint) Option {
	return func(svr *Server) error {
		if svr.server != nil {
			return errors.New("repetitive server endpoint")
		}

		svr.protocolType = ep.Proto()
		switch ep.Proto() {
		case ProtoTypeTcp:
			s, err := newTcp(ep, svr)
			if err != nil {
				return err
			}
			svr.server = s
			// add default tcp hook
			svr.sessionManager = NewSessionManager()
			svr.connectHookList = append(svr.connectHookList, svr.defaultConnectedHook)
			svr.disconnectHookList = append(svr.disconnectHookList, svr.defaultDisconnectedHook)
		case ProtoTypeUdp:
			s, err := newUdp(ep, svr)
			if err != nil {
				return err
			}
			svr.server = s
		case ProtoTypeWs:
			svr.server = newWS(ep, svr)
			svr.sessionManager = NewSessionManager()
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

		h, err := svr.msgRouter.getHandler(svr.protocolType)
		if err != nil {
			return err
		}
		svr.handler = h
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

func (svr *Server) RegisterMsgHandler(id uint16, handler MsgHandler) error {
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
func (svr *Server) defaultConnectedHook(session Session) {
	svr.sessionManager.Add(session)
}

func (svr *Server) defaultDisconnectedHook(session Session) {
	session.Close()
	svr.sessionManager.Remove(session.GetSessionID())
}
