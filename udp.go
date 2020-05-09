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

func (us *UdpSession) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	return us.conn.WriteToUDP(b, addr)
}

func (us *UdpSession) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	return us.conn.ReadFromUDP(b)
}

func (us *UdpSession) Close() error {
	return us.conn.Close()
}
