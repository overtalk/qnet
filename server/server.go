package server

import (
	"errors"
	"github.com/overtalk/qnet/model"
)

// Handler defines the func to handle net message
type Handler func(session Session)

type Server struct {
	server             model.IServer
	decoder            *decoder
	sessionManager     *sessionManager
	handler            Handler
	connectHookList    []Handler
	disconnectHookList []Handler
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
		return errors.New("message handler is nil")
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
	if svr.decoder == nil {
		return errors.New("decoder is absent")
	}

	return svr.decoder.registerMsgHandler(id, handler)
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
func (svr *Server) defaultTcpConnectedHook(session Session) {
	svr.sessionManager.Add(session)
}

func (svr *Server) defaultTcpDisconnectedHook(session Session) {
	session.Close()
	svr.sessionManager.Remove(session.GetSessionID())
}
