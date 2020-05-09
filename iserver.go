package qnet

// SessionHandler defines the func to handle a session/connection
type SessionHandler func(session Session)

type MsgHandler func(session Session, msg *NetMsg) *NetMsg

type IServer interface {
	Start() error
	Stop()
}
