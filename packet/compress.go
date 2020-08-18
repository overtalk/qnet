package packet

import "github.com/overtalk/qnet/pool"

// provide some methods to compress a packet of data
var (
	zlibPool *pool.ZlibWriterPool
)

// InitZlibPool initialize a zlibpool with a pool size
func InitZlibPool(size int) {
	zlibPool = pool.NewZlibWriterPool(size)
}

// ICompresser a data compresser
type ICompresser interface {
	Compress([]byte) ([]byte, bool)
	Close()
}

// noneCompresser a compresser doing nothing
type noneCompresser struct{}

// NoneCompresser a global noneCompresser
var NoneCompresser = &noneCompresser{}

func (*noneCompresser) Compress(b []byte) ([]byte, bool) { return b, false }
func (*noneCompresser) Close()                           {}

// zlibCompresser
type zlibCompresser struct {
	minSize int
	writer  *pool.ZlibWriter
}

// NewZlibCompresser create a zlibCompresser struct
func NewZlibCompresser(minSize int) ICompresser {
	return &zlibCompresser{minSize: minSize, writer: nil}
}

// Compress compress some data
func (zc *zlibCompresser) Compress(b []byte) ([]byte, bool) {
	if len(b) > zc.minSize {
		zc.writer = zlibPool.Get()
		return zc.writer.Compress(b), true
	}
	return b, false
}

// Close close the compresser
func (zc *zlibCompresser) Close() {
	if zc.writer != nil {
		zc.writer.Free()
		zc.writer = nil
	}
}
