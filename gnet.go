package qnet

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet"
)

type gNetServer struct {
	svr *NServer
}

func newGNetServer(svr *NServer) *gNetServer {
	return &gNetServer{svr: svr}
}

func (gns *gNetServer) Start() {
	if gns.svr.gNetOption == nil {
		log.Fatal(gnet.Serve(gns, gns.svr.ep.ToString()))
	} else {
		log.Fatal(gnet.Serve(gns, gns.svr.ep.ToString(), gnet.WithOptions(*gns.svr.gNetOption)))
	}
}

// for gnet
func (gns *gNetServer) OnInitComplete(server gnet.Server) Action {
	return gns.svr.onInitComplete(server)
}
func (gns *gNetServer) OnShutdown(server interface{}) { gns.svr.onShutdown(server) }
func (gns *gNetServer) OnOpened(c Conn) ([]byte, Action) {
	// add connection
	id := atomic.AddUint64(&gns.svr.baseID, 1)
	c.SetContext(id)
	gns.svr.rwMutex.Lock()
	gns.svr.connections[id] = c
	gns.svr.rwMutex.Unlock()

	return gns.svr.onOpened(c)
}
func (gns *gNetServer) OnClosed(c Conn, err error) Action {
	// delete connections
	id := c.Context().(uint64)
	gns.svr.rwMutex.Lock()
	delete(gns.svr.connections, id)
	gns.svr.rwMutex.Unlock()

	return gns.svr.onClosed(c, err)
}
func (gns *gNetServer) PreWrite()                                  { gns.svr.preWrite() }
func (gns *gNetServer) Tick() (delay time.Duration, action Action) { return gns.svr.tick() }
func (gns *gNetServer) React(frame []byte, c Conn) ([]byte, Action) {
	if gns.svr.msgRouterSwitch {
		// todo: use msg router
		return gns.svr.msgRouter.React(frame, c)
	}
	return gns.svr.react(frame, c)
}
