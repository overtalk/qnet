package base

//
//import (
//	"errors"
//	"net"
//
//	ringbuffer "github.com/ArkNX/ark-go/utils/ringBuffer"
//	"github.com/ArkNX/ark-go/utils/ringQueue"
//)
//
//type Conn interface {
//	// Context returns a user-defined context.
//	Context() (ctx interface{})
//
//	// SetContext sets a user-defined context.
//	SetContext(ctx interface{})
//
//	// LocalAddr is the connection's local socket address.
//	LocalAddr() (addr net.Addr)
//
//	// RemoteAddr is the connection's remote peer address.
//	RemoteAddr() (addr net.Addr)
//
//	// Read reads all data from inbound ring-buffer and event-loop-buffer without moving "read" pointer, which means
//	// it does not evict the data from buffers actually and those data will present in buffers until the
//	// ResetBuffer method is invoked.
//	Read() (buf []byte)
//
//	// ResetBuffer resets the buffers, which means all data in inbound ring-buffer and event-loop-buffer
//	// will be evicted.
//	ResetBuffer()
//
//	// ReadN reads bytes with the given length from inbound ring-buffer and event-loop-buffer without moving
//	// "read" pointer, which means it will not evict the data from buffers until the ShiftN method is invoked,
//	// it reads data from the inbound ring-buffer and event-loop-buffer and returns the size of bytes.
//	// If the length of the available data is less than the given "n", ReadN will returns all available data, so you
//	// should make use of the variable "size" returned by it to be aware of the exact length of the returned data.
//	ReadN(n int) (size int, buf []byte)
//
//	// ShiftN shifts "read" pointer in buffers with the given length.
//	ShiftN(n int) (size int)
//
//	// BufferLength returns the length of available data in the inbound ring-buffer.
//	BufferLength() (size int)
//
//	// InboundBuffer returns the inbound ring-buffer.
//	//InboundBuffer() *ringbuffer.RingBuffer
//
//	// SendTo writes data for UDP sockets, it allows you to send data back to UDP socket in individual goroutines.
//	SendTo(buf []byte) error
//
//	// AsyncWrite writes data to client/connection asynchronously, usually you would invoke it in individual goroutines
//	// instead of the event-loop goroutines.
//	AsyncWrite(buf []byte) error
//
//	// Wake triggers a React event for this connection.
//	Wake() error
//
//	// Close closes the current connection.
//	Close() error
//}
//
//////////////////////////////////////////////////////
//// net session
//////////////////////////////////////////////////////
//type NetSession struct {
//	sessionID  int64
//	headLength HeadLength
//	connected  bool
//	needRemove bool
//	conn       Conn
//
//	buffer   *ringbuffer.RingBuffer // for stream protocol, like tcp
//	msgQueue *ringQueue.RingQueue
//}
//
//type Option func(session *NetSession)
//
//func WithBuffer(size int) Option {
//	return func(session *NetSession) {
//		session.buffer = ringbuffer.New(size)
//	}
//}
//
//func WithQueue(size int) Option {
//	return func(session *NetSession) {
//		session.msgQueue = ringQueue.New(size)
//	}
//}
//
//func NewSession(headLength HeadLength, c Conn, opts ...Option) *NetSession {
//	ret := &NetSession{
//		headLength: headLength,
//		conn:       c,
//	}
//
//	for _, option := range opts {
//		option(ret)
//	}
//
//	return ret
//}
//
//func (netSession *NetSession) AddBuffer(data []byte) int {
//	cnt, _ := netSession.buffer.Write(data)
//	return cnt
//}
//
//func (netSession *NetSession) GetBuffer(n int) ([]byte, error) {
//	header, tail := netSession.buffer.LazyRead(n)
//
//	if len(header)+len(tail) != n {
//		return nil, errors.New("unmatched bytes length")
//	}
//
//	if len(tail) != 0 {
//		header = append(header, tail...)
//	}
//
//	return header, nil
//}
//
//func (netSession *NetSession) GetBufferLen() int { return netSession.buffer.Len() }
//
//func (netSession *NetSession) GetConn() Conn { return netSession.conn }
//
//func (netSession *NetSession) GetSessionID() int64 { return netSession.sessionID }
//
//func (netSession *NetSession) SetSessionID(value int64) { netSession.sessionID = value }
//
//func (netSession *NetSession) NeedRemove() bool { return netSession.needRemove }
//
//func (netSession *NetSession) SetNeedRemove(value bool) { netSession.needRemove = value }
//
//func (netSession *NetSession) AddNetMsg(msg *NetMsg) { netSession.msgQueue.PushOne(msg) }
//
//func (netSession *NetSession) PopNetMsg() (*NetMsg, bool) {
//	e, err := netSession.msgQueue.PopOne()
//	if err != nil {
//		return nil, false
//	}
//
//	return e.(*NetMsg), true
//}
//
//// ParseBufferToMsg
//func (netSession *NetSession) ParseBufferToMsg() {
//	for {
//		msg, err := netSession.getNetMsg()
//		if err != nil {
//			break
//		}
//
//		netSession.AddNetMsg(msg)
//	}
//}
//
//// getNetMsg defines the func to read msg from queue
//func (netSession *NetSession) getNetMsg() (*NetMsg, error) {
//	headerBytes, err := netSession.GetBuffer(int(netSession.headLength))
//	if err != nil {
//		return nil, err
//	}
//
//	header, err := DeserializeMsgHead(netSession.headLength, headerBytes)
//	if err != nil {
//		return nil, err
//	}
//
//	msg, err := netSession.GetBuffer(int(header.GetMsgLength()))
//	if err != nil {
//		return nil, err
//	}
//
//	netSession.buffer.Shift(int(uint32(netSession.headLength) + header.GetMsgLength()))
//
//	return &NetMsg{
//		head:    header,
//		msgData: msg,
//	}, nil
//}
