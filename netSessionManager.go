package qnet

import (
	"fmt"
	"sync"
)

type SessionManager struct {
	sync.RWMutex
	sessions map[uint64]Session // connections，for both tcp & udp
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[uint64]Session),
	}
}

func (manger *SessionManager) Add(s Session) {
	manger.Lock()
	//TODO: add a log
	manger.sessions[s.GetSessionID()] = s
	manger.Unlock()
}

func (manger *SessionManager) Remove(sessionID uint64) {
	manger.Lock()
	//TODO: add a log
	delete(manger.sessions, sessionID)
	manger.Unlock()
}

func (manger *SessionManager) Get(sessionID uint64) (Session, error) {
	manger.RLock()
	session, ok := manger.sessions[sessionID]
	manger.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session(id = %d) is absent", sessionID)
	}
	return session, nil
}

func (manger *SessionManager) Len() int {
	return len(manger.sessions)
}

func (manger *SessionManager) ClearSession() {
	manger.Lock()
	defer manger.Unlock()

	for sessionID, session := range manger.sessions {
		fmt.Println(session)
		// TODO: close all session
		session.Close()
		delete(manger.sessions, sessionID)
	}
}

func (manger *SessionManager) ClearClosedSession() {
	manger.Lock()
	defer manger.Unlock()

	for sessionID, session := range manger.sessions {
		fmt.Println(session)
		// TODO: close all session
		session.Close()
		delete(manger.sessions, sessionID)
	}
}
