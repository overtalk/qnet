package server

import (
	"fmt"
	"sync"
	"time"
)

type sessionManager struct {
	sync.RWMutex
	sessions map[uint64]Session // connections，for both tcp & udp
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		sessions: make(map[uint64]Session),
	}
}

func (manger *sessionManager) Add(s Session) {
	manger.Lock()
	//TODO: add a log
	manger.sessions[s.GetSessionID()] = s
	manger.Unlock()
}

func (manger *sessionManager) Remove(sessionID uint64) {
	manger.Lock()
	//TODO: add a log
	delete(manger.sessions, sessionID)
	manger.Unlock()
}

func (manger *sessionManager) Get(sessionID uint64) (Session, error) {
	manger.RLock()
	session, ok := manger.sessions[sessionID]
	manger.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session(id = %d) is absent", sessionID)
	}
	return session, nil
}

func (manger *sessionManager) Len() int {
	return len(manger.sessions)
}

func (manger *sessionManager) ClearSession() {
	manger.Lock()
	defer manger.Unlock()

	for sessionID, session := range manger.sessions {
		fmt.Println(session)
		// TODO: close all session
		session.Close()
		delete(manger.sessions, sessionID)
	}
}

func (manger *sessionManager) ClearClosedSession() {
	manger.Lock()
	defer manger.Unlock()

	for sessionID, session := range manger.sessions {
		if session.GetClosed() {
			fmt.Println(session)
			// TODO: close all session
			session.Close()
			delete(manger.sessions, sessionID)
		}

	}
}

func (manger *sessionManager) daemon() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:

		}
	}
}
