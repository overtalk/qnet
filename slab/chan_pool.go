package slab

import "unsafe"

// ChanPool is a chan based slab allocation memory pool.
type ChanPool struct {
	pages   []chanPage
	minSize int
	maxSize int
}

// NewChanPool create a chan based slab allocation memory pool.
// minSize is the smallest chunk size.
// maxSize is the lagest chunk size.
// factor is used to control growth of chunk size.
// pageSize is the memory size of each slab class.
func NewChanPool(minSize, maxSize, factor, pageSize int) *ChanPool {
	pool := &ChanPool{make([]chanPage, 0, 10), minSize, maxSize}
	for chunkSize := minSize; chunkSize <= maxSize && chunkSize <= pageSize; chunkSize *= factor {
		c := chanPage{
			size:   chunkSize,
			page:   make([]byte, pageSize),
			chunks: make(chan []byte, pageSize/chunkSize),
		}
		c.pageBegin = uintptr(unsafe.Pointer(&c.page[0]))
		for i := 0; i < pageSize/chunkSize; i++ {
			// lock down the capacity to protect append operation
			mem := c.page[i*chunkSize : (i+1)*chunkSize : (i+1)*chunkSize]
			c.chunks <- mem
			if i == len(c.chunks)-1 {
				c.pageEnd = uintptr(unsafe.Pointer(&mem[0]))
			}
		}
		pool.pages = append(pool.pages, c)
		pool.maxSize = chunkSize
	}
	return pool
}

// Alloc try alloc a []byte from internal slab class if no free chunk in slab class Alloc will make one.
func (pool *ChanPool) Alloc(size int) []byte {
	if size <= pool.maxSize {
		for i := 0; i < len(pool.pages); i++ {
			if pool.pages[i].size >= size {
				mem := pool.pages[i].Pop()
				if mem != nil {
					return mem[:size]
				}
				break
			}
		}
	}
	return make([]byte, size)
}

// Free release a []byte that alloc from Pool.Alloc.
func (pool *ChanPool) Free(mem []byte) {
	if size := cap(mem); size <= pool.maxSize {
		for i := 0; i < len(pool.pages); i++ {
			if pool.pages[i].size == size {
				pool.pages[i].Push(mem)
				break
			}
		}
	}
}

type chanPage struct {
	size      int
	page      []byte
	pageBegin uintptr
	pageEnd   uintptr
	chunks    chan []byte
}

func (c *chanPage) Push(mem []byte) {
	select {
	case c.chunks <- mem:
	default:
		mem = nil
	}
}

func (c *chanPage) Pop() []byte {
	select {
	case mem := <-c.chunks:
		return mem
	default:
		return nil
	}
}
