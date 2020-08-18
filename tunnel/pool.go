package tunnel

import (
	"net"

	"github.com/overtalk/qnet/pool"
	"github.com/overtalk/qnet/slab"
)

// SessionPool some pools for the session read/write
type SessionPool struct {
	rdrBufPool slab.Pool
	bufRdrPool *pool.BufReaderPool
}

// NewSessionPool create a SessionPool struct
func NewSessionPool(rdrBufPool slab.Pool, bufRdrPool *pool.BufReaderPool) *SessionPool {
	return &SessionPool{rdrBufPool, bufRdrPool}
}

// GetBufReader get a bytes buffer from the pool
func (sp *SessionPool) GetBufReader(nc net.Conn) *pool.BufReader {
	return sp.bufRdrPool.Get(nc)
}

// GetRdrBufPool get the rdrBufPool
func (sp *SessionPool) GetRdrBufPool() slab.Pool {
	return sp.rdrBufPool
}
