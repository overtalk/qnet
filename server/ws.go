package server

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"

	"github.com/overtalk/qnet/base"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type ws struct {
	id       uint64
	svr      *Server
	ep       *base.Endpoint // endpoint
	wsServer http.Server
	stopFlag bool
	stopChan chan interface{} // close signal channel
}

func newWS(ep *base.Endpoint, svr *Server) *ws {
	ws := &ws{
		ep:       ep,
		svr:      svr,
		stopFlag: true,
		stopChan: make(chan interface{}),
	}

	mux := http.NewServeMux()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	mux.HandleFunc(ep.GetPath(), ws.websocketHandler)

	ws.wsServer = http.Server{
		Addr:    fmt.Sprintf("%s:%d", ep.GetIP(), ep.GetPort()),
		Handler: c.Handler(mux),
	}

	return ws
}

func (ws *ws) Start() error {
	return ws.wsServer.ListenAndServe()
}

func (ws *ws) Stop() {
	ws.stopFlag = true
	ws.stopChan <- struct{}{}
}

func (ws *ws) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		// todo : handle error
		return
	}

	sessionID := atomic.AddUint64(&ws.id, 1)

	session := base.NewWsSession(sessionID, conn)

	// do some hook
	for _, connectHook := range ws.svr.connectHookList {
		connectHook(session)
	}

	// do logic
	ws.svr.handler(session)

	// do some hook
	for _, connectHook := range ws.svr.disconnectHookList {
		connectHook(session)
	}

}
