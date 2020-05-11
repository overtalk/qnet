package qnet

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"net/http"
	"sync/atomic"
)

type websocketServer struct {
	svr      *NServer
	wsServer http.Server
	stopFlag bool
	stopChan chan interface{} // close signal channel
}

func NewWebsocketServer(svr *NServer) *websocketServer {
	ws := &websocketServer{
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
	mux.HandleFunc(svr.ep.GetPath(), ws.websocketHandler)

	ws.wsServer = http.Server{
		Addr:    fmt.Sprintf("%s:%d", svr.ep.GetIP(), svr.ep.GetPort()),
		Handler: c.Handler(mux),
	}

	return ws
}

func (ws *websocketServer) Start() error {
	return ws.wsServer.ListenAndServe()
}

func (ws *websocketServer) Stop() {
	ws.stopFlag = true
	ws.stopChan <- struct{}{}
}

func (ws *websocketServer) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		// todo : handle error
		return
	}

	connID := atomic.AddUint64(&ws.svr.baseID, 1)

	NewWsSession(connID, conn)

	// do some hook
	ws.svr.onOpened(nil)

	// do logic
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if mt != websocket.BinaryMessage {
			// todo: handle error
			break
		}

		retBytes, action := ws.svr.React(message, nil)
		if retBytes != nil {
			conn.WriteMessage(websocket.BinaryMessage, retBytes)
		}

		if action != 0 {

		}

	}

	// do some hook
	ws.svr.onClosed(nil, nil)
}
