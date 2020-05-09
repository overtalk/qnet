package qnet

import (
	"errors"
	"net"
)

type udp struct {
	svr     *Server
	ep      *Endpoint            // endpoint
	udpConn *net.UDPConn         // for udp
	natMap  map[int]*net.UDPAddr // nat map
}

func newUdp(ep *Endpoint, svr *Server) (*udp, error) {
	addr, err := ep.UDPAddr()
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP(string(ProtoTypeUdp), addr)
	if err != nil {
		return nil, err
	}

	return &udp{
		ep:      ep,
		svr:     svr,
		udpConn: udpConn,
	}, nil
}

func (u *udp) Start() error {
	if u.udpConn == nil {
		return errors.New("udp conn is nil")
	}

	session := NewUdpSession(0, u.udpConn)

	go u.svr.handler(session)
	return nil
}

func (u *udp) Stop() {
	u.udpConn.Close()
}

// --------------------------------------------------
//type UdpSession
type UdpSession struct {
	BasicSession
	conn *net.UDPConn
}

func NewUdpSession(sessionID uint64, conn *net.UDPConn) *UdpSession {
	return &UdpSession{
		BasicSession: *NewBasicSession(sessionID),
		conn:         conn,
	}
}

func (us *UdpSession) Close() error                                   { return us.conn.Close() }
func (us *UdpSession) UdpWrite(b []byte, a *net.UDPAddr) (int, error) { return us.conn.WriteToUDP(b, a) }
func (us *UdpSession) UdpRead(b []byte) (int, *net.UDPAddr, error)    { return us.conn.ReadFromUDP(b) }

func (us *UdpSession) GetNetMsg(length HeadLength, decoderFunc HeadDeserializeFunc) (*NetMsg, *net.UDPAddr, error) {
	packet := make([]byte, 1024)
	n, remoteAddr, err := us.UdpRead(packet)
	if err != nil {
		return nil, nil, err
	}

	// decode head
	head, err := decoderFunc(packet[:length])
	if err != nil {
		return nil, nil, err
	}

	return NewNetMsg(head, packet[length:n]), remoteAddr, nil
}

func (us *UdpSession) SendNetMsg(headSerializeFunc HeadSerializeFunc, msg *NetMsg, addr *net.UDPAddr) error {
	bytes := headSerializeFunc(msg)
	bytes = append(bytes, msg.GetMsg()...)
	_, err := us.UdpWrite(bytes, addr)
	return err
}
