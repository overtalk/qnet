package server

import (
	"errors"
	"fmt"

	"github.com/overtalk/qnet/base"
	"github.com/overtalk/qnet/model"
)

// Handler defines the func to handle net message
type Handler func(session Session)

type server struct {
	server             model.IServer
	sessionManager     *sessionManager
	handler            Handler
	connectHookList    []Handler
	disconnectHookList []Handler
}

func NewServerFromString(url string, f Handler) (*server, error) {
	ep, err := base.NewFromString(url)
	if err != nil {
		return nil, err
	}

	return NewServerWithHandler(ep, f)
}

func NewServerWithHandler(ep *base.Endpoint, f Handler) (*server, error) {
	svr, err := NewServer(ep)
	if err != nil {
		return nil, err
	}

	svr.SetHandler(f)
	return svr, nil
}

func NewServer(ep *base.Endpoint) (*server, error) {
	svr := &server{
		sessionManager: newSessionManager(),
	}

	switch ep.Proto() {
	case base.ProtoTypeTcp:
		s, err := newTcp(ep, svr)
		if err != nil {
			return nil, err
		}
		svr.server = s
		// add default tcp hook
		svr.connectHookList = append(svr.connectHookList, svr.defaultTcpConnectedHook)
	case base.ProtoTypeUdp:
		s, err := newUdp(ep, svr)
		if err != nil {
			return nil, err
		}
		svr.server = s
	case base.ProtoTypeWs:
	default:
		return nil, fmt.Errorf("invalid net protocol : %s", ep.Proto())
	}

	return svr, nil
}

func (svr *server) Start() error {
	if svr.handler == nil {
		return errors.New("message handler is nil")
	}

	return svr.server.Start()
}

func (svr *server) Stop() {
	svr.server.Stop()
	svr.sessionManager.ClearSession()
}

func (svr *server) SetHandler(f Handler) {
	svr.handler = f
}

func (svr *server) SendBySessionID(sessionID uint64, data []byte) (int, error) {
	session, err := svr.sessionManager.Get(sessionID)
	if err != nil {
		return 0, err
	}

	return session.Write(data)
}

func (svr *server) SetSessionMeta(sessionID uint64, key string, value interface{}) error {
	session, err := svr.sessionManager.Get(sessionID)
	if err != nil {
		return err
	}

	session.SetMeta(key, value)
	return nil
}

func (svr *server) GetSessionMeta(sessionID uint64, key string) (interface{}, error) {
	session, err := svr.sessionManager.Get(sessionID)
	if err != nil {
		return nil, err
	}

	return session.GetMeta(key)
}

// ----------------------- hook ---------------------------
func (svr *server) AddConnectHook(hook Handler) *server {
	svr.connectHookList = append(svr.connectHookList, hook)
	return svr
}

func (svr *server) AddDisconnectHook(hook Handler) *server {
	svr.disconnectHookList = append(svr.disconnectHookList, hook)
	return svr
}

func (svr *server) defaultTcpConnectedHook(session Session) {
	svr.sessionManager.Add(session)
}

func (svr *server) defaultTcpDisconnectedHook(session Session) {
	session.SetClosed(true)
}

//close signal
//func (svr *server) signalShutdown() {
//	svr.once.Do(func() {
//		svr.cond.L.Lock()
//		svr.cond.Signal()
//		svr.cond.L.Unlock()
//	})
//}
//
//// waitForShutdown waits for a signal to shutdown
//func (svr *server) waitForShutdown() {
//	svr.cond.L.Lock()
//	svr.cond.Wait()
//	svr.cond.L.Unlock()
//}
