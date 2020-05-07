package server

import (
	"errors"
	"net"

	"github.com/overtalk/qnet/base"
)

type udp struct {
	svr     *Server
	ep      *base.Endpoint       // endpoint
	udpConn *net.UDPConn         // for udp
	natMap  map[int]*net.UDPAddr // nat map
}

func newUdp(ep *base.Endpoint, svr *Server) (*udp, error) {
	addr, err := ep.UDPAddr()
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP(string(base.ProtoTypeUdp), addr)
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

	session := NewTcpSession(0, u.udpConn)

	go u.svr.handler(session)
	return nil
}

func (u *udp) Stop() {
	u.udpConn.Close()
}
