package tunnel

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/overtalk/qnet/common"
	"github.com/overtalk/qnet/packet"
	"github.com/overtalk/qnet/pool"
	"github.com/overtalk/qnet/slab"
)

var frontendPool *SessionPool

// InitFrontendPool init some pools for the frontend
func InitFrontendPool() {
	frontendPool = NewSessionPool(
		slab.NewAtomPool(512, 4*1024, 2, 4*1024*1024), // pre-allocated: 16MBytes
		pool.NewBufReaderPool(10000, 1024),
	)
}

// FrontendSession frontend clients
type FrontendSession struct {
	id     uint32
	conn   *common.BaseConn
	buffer common.IPacketBuffer
	closed int32
	done   chan struct{}

	// connected backend
	backend *BackendSession
}

// NewFrontendSession create a FrontendSession struct
func NewFrontendSession(nc net.Conn) *FrontendSession {
	baseConn := common.NewBaseConn(nc, frontendPool.GetBufReader(nc))
	baseConn.SetTimeout(10 * time.Second)
	return &FrontendSession{
		id:   0,
		conn: baseConn,
		buffer: common.NewPacketBuffer(
			packet.MaxPacketSize,
			frontendPool.GetRdrBufPool(),
		),
		done: make(chan struct{}),
	}
}

// ClientAddr get the remoted client address
func (s *FrontendSession) ClientAddr() string {
	return s.conn.RemoteAddr()
}

// GetID set the session id
func (s *FrontendSession) GetID() uint32 {
	return s.id
}

// ReadPacket read a packet
func (s *FrontendSession) ReadPacket() (packet.Packet, error) {
	err := s.conn.ReadPacket(s.buffer)
	if err == nil {
		return packet.Packet(s.buffer.Bytes()), nil
	}
	return nil, err
}

func (s *FrontendSession) Write(b []byte) (int, error) {
	return s.conn.Write(b)
}

// Close close the underlying tcp session and release the resource
func (s *FrontendSession) Close() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		s.conn.Close()
		s.buffer.Free()
	}
}

// IsClosed check whether the session is closed
func (s *FrontendSession) IsClosed() bool {
	return atomic.LoadInt32(&s.closed) == 1
}

// BindBackendSession bind it to a backend session
func (s *FrontendSession) BindBackendSession(backend *BackendSession) {
	if s.backend == nil {
		s.id = backend.NewFrontendSessionID()
		s.backend = backend
		backend.AddFrontendSession(s)
	}
}

// UnBindBackendSession unbind it from a backend session
func (s *FrontendSession) UnBindBackendSession() {
	if s.backend != nil {
		s.backend.DelFrontendSession(s.id)
		s.id, s.backend = 0, nil
	}
}

// WaitResponse wait its response arriving
func (s *FrontendSession) WaitResponse() {
	select {
	case <-s.done:
	case <-time.After(10 * time.Second):
		// zaplog.S.Errorf("client-%d@%s: response timeout", s.id, s.ClientAddr())
	}
}

// DoneResponse set its completion
func (s *FrontendSession) DoneResponse() {
	close(s.done)
}
