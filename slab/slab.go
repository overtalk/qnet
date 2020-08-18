package slab

// Pool memory pool interface protocol
type Pool interface {
	Alloc(int) []byte
	Free([]byte)
}

// NoPool a non-pool memory allocator
type NoPool struct{}

// Alloc make a size of bytes
func (p *NoPool) Alloc(size int) []byte {
	return make([]byte, size)
}

// Free free allocated bytes
func (p *NoPool) Free(_ []byte) {}

var _ Pool = (*NoPool)(nil)
var _ Pool = (*ChanPool)(nil)
var _ Pool = (*SyncPool)(nil)
var _ Pool = (*AtomPool)(nil)
