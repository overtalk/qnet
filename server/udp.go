package server

import (
	"errors"
	"net"

	"github.com/overtalk/qnet/base"
)

type udp struct {
	svr     *server
	ep      *base.Endpoint // endpoint
	udpConn *net.UDPConn   // for udp
}

func newUdp(ep *base.Endpoint, svr *server) (*udp, error) {
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

	//r := bufio.NewReader(svr.udpConn)
	//for {
	//	line, err := r.ReadBytes(byte('\n'))
	//	switch err {
	//	case nil:
	//		break
	//	case io.EOF:
	//	default:
	//		fmt.Println("ERROR", err)
	//	}
	//	svr.udpConn.Write(line)
	//}
}

func (u *udp) Stop() {}
