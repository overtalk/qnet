package qnet

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"net"
	"net/http"
	"sync/atomic"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type webSocketServer struct {
	svr      *NServer
	wsServer http.Server
	stopFlag bool
	stopChan chan interface{} // close signal channel
}

func newWebSocketServer(svr *NServer) *webSocketServer {
	ws := &webSocketServer{
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
	mux.HandleFunc(svr.ep.GetPath(), ws.webSocketHandler)

	ws.wsServer = http.Server{
		Addr:    fmt.Sprintf("%s:%d", svr.ep.GetIP(), svr.ep.GetPort()),
		Handler: c.Handler(mux),
	}

	return ws
}

func (ws *webSocketServer) Start() error {
	return ws.wsServer.ListenAndServe()
}

func (ws *webSocketServer) Stop() {
	ws.stopFlag = true
	ws.stopChan <- struct{}{}
}

func (ws *webSocketServer) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		// todo : handle error
		return
	}

	connID := atomic.AddUint64(&ws.svr.baseID, 1)

	wsConn := newWebSocketConn(conn)
	wsConn.SetContext(connID)

	// do some hook
	ws.svr.onOpened(wsConn)

	// do logic
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if mt != websocket.BinaryMessage {
			// todo: handle error
			continue
		}

		retBytes, action := ws.svr.React(message, wsConn)
		if retBytes != nil {
			conn.WriteMessage(websocket.BinaryMessage, retBytes)
		}

		if action != 0 {

		}
	}

	// do some hook
	ws.svr.onClosed(nil, nil)
}

// -------------------------------------------------
type webSocketConn struct {
	svr  *NServer
	conn *websocket.Conn
	ctx  interface{} // user-defined context
}

func newWebSocketConn(conn *websocket.Conn) *webSocketConn {
	return &webSocketConn{conn: conn}
}

func (wc *webSocketConn) Context() (ctx interface{})  { return wc.ctx }
func (wc *webSocketConn) SetContext(ctx interface{})  { wc.ctx = ctx }
func (wc *webSocketConn) LocalAddr() (addr net.Addr)  { return wc.conn.LocalAddr() }
func (wc *webSocketConn) RemoteAddr() (addr net.Addr) { return wc.conn.RemoteAddr() }
func (wc *webSocketConn) Close() error                { return wc.conn.Close() }
func (wc *webSocketConn) SendTo(buf []byte) error {
	return wc.conn.WriteMessage(websocket.BinaryMessage, buf)
}

func (wc *webSocketConn) Wake() error {
	wc.svr.React(nil, wc)
	return nil
}

// useless func
func (wc *webSocketConn) Read() (buf []byte)                 { return nil }
func (wc *webSocketConn) ResetBuffer()                       {}
func (wc *webSocketConn) ReadN(n int) (size int, buf []byte) { return 0, nil }
func (wc *webSocketConn) ShiftN(n int) (size int)            { return 0 }
func (wc *webSocketConn) BufferLength() (size int)           { return 0 }
func (wc *webSocketConn) AsyncWrite(buf []byte) error        { return nil }
