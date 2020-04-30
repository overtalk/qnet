package model

type IServer interface {
	Start() error
	Stop()

	//SendBySessionID(sessionID uint64, data []byte) (int, error)
	//SetSessionMeta(sessionID uint64, key string, value interface{}) error
	//GetSessionMeta(sessionID uint64, key string) (interface{}, error)
}
