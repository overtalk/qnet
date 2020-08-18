package tunnel

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/overtalk/qnet/common"
	"github.com/overtalk/qnet/packet"
	"github.com/overtalk/qnet/pool"
	"github.com/overtalk/qnet/slab"
)

var backendPool *SessionPool

// InitBackendPool init some pools for the frontend
func InitBackendPool() {
	backendPool = NewSessionPool(
		slab.NewAtomPool(512, 32*1024, 2, 8*1024*1024), // pre-allocated: 56MBytes
		pool.NewBufReaderPool(1000, 64*1024),
	)
}

// BackendRequest a request for backend
type BackendRequest struct {
	buffer common.IPacketBuffer
}

// NewBackendRequest create a BackendRequest
func NewBackendRequest() *BackendRequest {
	return &BackendRequest{
		buffer: common.NewPacketBuffer(
			packet.MaxPacketSize,
			backendPool.GetRdrBufPool(),
		),
	}
}

// Read read a steam of data from the backend session
func (req *BackendRequest) Read(r common.IPacketReader) error {
	return r.ReadPacket(req.buffer)
}

// Free free its underlying resource
func (req *BackendRequest) Free() {
	req.buffer.Free()
}

// GetPacket get the packet
func (req *BackendRequest) GetPacket() packet.Packet {
	return packet.Packet(req.buffer.Bytes())
}

// backendConnState a state for backend connection
type backendConnState struct {
	totalTryNum uint32
	lastTryTime int64
}

func newBackendConnState() *backendConnState {
	return &backendConnState{
		totalTryNum: 0, lastTryTime: 0,
	}
}

func (bcs *backendConnState) reset() {
	bcs.totalTryNum, bcs.lastTryTime = 0, 0
}

func (bcs *backendConnState) update(now int64) {
	bcs.totalTryNum++
	bcs.lastTryTime = now
}

var (
	// binary exponential backoff
	minSecondsForTryConnect = [8]byte{1, 1, 2, 2, 2, 4, 4, 8}
)

func (bcs *backendConnState) getMinTryTime() int64 {
	return int64(minSecondsForTryConnect[bcs.totalTryNum&0x07])
}

func (bcs *backendConnState) tryAgain() bool {
	nowTS := time.Now().Unix()
	ok := (nowTS - bcs.lastTryTime) >= bcs.getMinTryTime()
	if ok {
		// update its state after each trying
		bcs.update(nowTS)
	}
	return ok
}

// BackendSession backend services
type BackendSession struct {
	id       uint32
	conn     *common.BaseConn
	closed   int32
	sigClose chan struct{} // notify the session closed
	pingtime int64         // timestamp for ping

	// TODO: use sync.Map to reduce the lock contention
	// manage all FrontendSession attached to it
	frontends map[uint32]*FrontendSession
	lock      sync.RWMutex

	// session id generator
	idCounter uint32
	timeStart time.Time

	// wait all reqeusts being done
	waitRequest *sync.WaitGroup
}

const (
	minPingTime = 20
)

// NewBackendSession create a BackendSession struct
func NewBackendSession(id uint32, nc net.Conn) *BackendSession {
	baseConn := common.NewBaseConn(nc, backendPool.GetBufReader(nc))
	baseConn.SetTimeout(10 * time.Second)
	nowTime := time.Now()
	return &BackendSession{
		id:          id,
		conn:        baseConn,
		closed:      0,
		sigClose:    make(chan struct{}),
		pingtime:    nowTime.Unix(),
		frontends:   map[uint32]*FrontendSession{},
		idCounter:   0,
		timeStart:   nowTime,
		waitRequest: new(sync.WaitGroup),
	}
}

// GetID set the session id
func (s *BackendSession) GetID() uint32 {
	return s.id
}

// ClientAddr get the remoted client address
func (s *BackendSession) ClientAddr() string {
	return s.conn.RemoteAddr()
}

// Register register it to an agent
func (s *BackendSession) Register(sid uint32) error {
	_, err := s.Write(packet.NewRegister(sid))
	return err
}

// AddRequest add a request to be done
func (s *BackendSession) AddRequest() {
	s.waitRequest.Add(1)
}

// DoneRequest make a request done
func (s *BackendSession) DoneRequest() {
	s.waitRequest.Done()
}

// WaitRequestDone wait all requests done
func (s *BackendSession) WaitRequestDone() {
	s.waitRequest.Wait()
}

// UpdatePing update the ping time
func (s *BackendSession) UpdatePing() {
	atomic.StoreInt64(&s.pingtime, time.Now().Unix())
}

// Ping send a PingPacket to another endpoint
func (s *BackendSession) Ping() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				//zaplog.S.Error(err)
				//zaplog.S.Error(zap.Stack("").String)
			}
		}()

		ticker := time.NewTicker((minPingTime - 2) * time.Second)
		for {
			select {
			case <-ticker.C:
				//zaplog.S.Infof("ping: agent@%s ---> backend-%d@%s", s.conn.LocalAddr(), s.id, s.ClientAddr())
				_, err := s.conn.Write(packet.PingPacket)
				if err != nil {
					//zaplog.S.Errorf("ping: agent@%s ---> backend-%d@%s, %v", s.conn.LocalAddr(), s.id, s.ClientAddr(), err)
					s.conn.Close()
					ticker.Stop()
					return
				}
			case <-s.sigClose:
				ticker.Stop()
				return
			}
		}
	}()
}

// CheckPing check whether the underlying is ok
func (s *BackendSession) CheckPing() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				//zaplog.S.Error(err)
				//zaplog.S.Error(zap.Stack("").String)
			}
		}()

		ticker := time.NewTicker(minPingTime * time.Second)
		for {
			select {
			case now := <-ticker.C:
				lastPingTime := atomic.LoadInt64(&s.pingtime)
				if now.Unix()-lastPingTime > minPingTime {
					// ping timeout
					//zaplog.S.Errorf("ping timeout: agent@%s -> backend-%d@%s", s.ClientAddr(), s.id, s.conn.LocalAddr())
					s.conn.Close()
					ticker.Stop()
					return
				}
			case <-s.sigClose:
				ticker.Stop()
				return
			}
		}
	}()
}

// ReadRequest read a request
func (s *BackendSession) ReadRequest() (*BackendRequest, error) {
	req := NewBackendRequest()
	err := req.Read(s.conn)
	return req, err
}

func (s *BackendSession) Write(b []byte) (int, error) {
	return s.conn.Write(b)
}

// NewFrontendSessionID create a tunnel session id
func (s *BackendSession) NewFrontendSessionID() uint32 {
	// id starts from 101, but the returned id may be 0.
	return 100 + atomic.AddUint32(&s.idCounter, 1)
}

// GetFrontendSession add a FrontendSession
func (s *BackendSession) GetFrontendSession(id uint32) *FrontendSession {
	s.lock.RLock()
	sess := s.frontends[id]
	s.lock.RUnlock()
	return sess
}

// AddFrontendSession add a FrontendSession
func (s *BackendSession) AddFrontendSession(sess *FrontendSession) {
	s.lock.Lock()
	s.frontends[sess.GetID()] = sess
	s.lock.Unlock()
}

// DelFrontendSession add a FrontendSession
func (s *BackendSession) DelFrontendSession(id uint32) {
	s.lock.Lock()
	delete(s.frontends, id)
	s.lock.Unlock()
}

// closeAllFrontendSessions close all frontend sessions(not concurrent safely)
func (s *BackendSession) closeAllFrontendSessions() {
	for _, v := range s.frontends {
		if v != nil {
			v.conn.Close()
		}
	}
	// clear all frontend sessions
	s.frontends = map[uint32]*FrontendSession{}
}

// Close close the underlying tcp session and release the resource
func (s *BackendSession) Close() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		close(s.sigClose)
		s.closeAllFrontendSessions()
		s.conn.Close()
	}
}
