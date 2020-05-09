package qnet

import (
	"fmt"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type ws struct {
	id       uint64
	svr      *Server
	ep       *Endpoint // endpoint
	wsServer http.Server
	stopFlag bool
	stopChan chan interface{} // close signal channel
}

func newWS(ep *Endpoint, svr *Server) *ws {
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

	session := NewWsSession(sessionID, conn)

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

// ------------------------------------------
type WsSession struct {
	BasicSession
	conn *websocket.Conn
}

func NewWsSession(sessionID uint64, conn *websocket.Conn) *WsSession {
	return &WsSession{
		BasicSession: *NewBasicSession(sessionID),
		conn:         conn,
	}
}

func (ws *WsSession) WsRead() (messageType int, data []byte, err error) { return ws.conn.ReadMessage() }
func (ws *WsSession) WsWrite(t int, data []byte) error                  { return ws.conn.WriteMessage(t, data) }
func (ws *WsSession) Close() error                                      { return ws.conn.Close() }

func (ws *WsSession) WriteBinaryMessage(data []byte) (int, error) {
	return len(data), ws.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (ws *WsSession) ReadBinaryMessage() ([]byte, error) {
	mt, message, err := ws.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	if mt != websocket.BinaryMessage {
		// todo: handle error
		return nil, fmt.Errorf("invalid websocket messageType : %d", mt)
	}

	return message, nil
}

// for msg router
func (ws *WsSession) GetNetMsg(length HeadLength, decoderFunc HeadDeserializeFunc) (*NetMsg, *net.UDPAddr, error) {
	packet, err := ws.ReadBinaryMessage()
	if err != nil {
		return nil, nil, err
	}

	// decode head
	head, err := decoderFunc(packet[:length])
	if err != nil {
		return nil, nil, err
	}

	return NewNetMsg(head, packet[length:]), nil, nil
}
func (ws *WsSession) SendNetMsg(headSerializeFunc HeadSerializeFunc, msg *NetMsg, _ *net.UDPAddr) error {
	bytes := headSerializeFunc(msg)
	bytes = append(bytes, msg.GetMsg()...)
	_, err := ws.WriteBinaryMessage(bytes)
	return err

}
