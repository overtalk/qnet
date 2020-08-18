package session

import (
	"github.com/overtalk/qnet/common"
	"github.com/overtalk/qnet/packet"
	"github.com/overtalk/qnet/tunnel"
	"net"
	"time"
)

// AgentService an agent service
type AgentService struct {
	router *common.Router
}

// NewAgentService create a AgentSession struct
func NewAgentService(router *common.Router) *AgentService {
	return &AgentService{router: router}
}

// Serve serve a tcp session from the agent server
func (as *AgentService) Serve(nc net.Conn) {
	backendSess := tunnel.NewBackendSession(0, nc)
	defer func() {
		if err := recover(); err != nil {
			//zaplog.S.Error(err)
			//zaplog.S.Error(zap.Stack("").String)
		}
		backendSess.Close()
	}()

	// it's a long session
	backendSess.CheckPing()

	// all requests must be handled after breaking the for loop
	for {
		inRequest, err := backendSess.ReadRequest()
		if err == nil {
			go as.handleAgentRequest(backendSess, inRequest)
		} else {
			inRequest.Free()
			if IsNetTimeout(err) {
				//zaplog.S.Errorf("read agent@%s request: %v", backendSess.ClientAddr(), err)
				break
			}
		}
	}

	// wait 5 seconds before closing the connection and exit
	WaitAction(func() { backendSess.WaitRequestDone() }, 5*time.Second)
}

func (as *AgentService) handleAgentCmd(sess *tunnel.BackendSession, pack packet.Packet) {
	cmd := pack.GetCmd()
	switch cmd {
	case packet.CmdPing:
		sess.UpdatePing()
	default:
		//zaplog.S.Errorf("agent@%s: packet: %v, invalid cmd(%d)", sess.ClientAddr(), cmd)
	}
}

func (as *AgentService) handleAgentRequest(
	sess *tunnel.BackendSession, req *tunnel.BackendRequest) {
	sess.AddRequest()
	defer func() {
		if err := recover(); err != nil {
			//zaplog.S.Error(err)
			//zaplog.S.Error(zap.Stack("").String)
		}
		req.Free()
		sess.DoneRequest()
	}()

	// no need to decrypt the data from an agent server
	inPacket := req.GetPacket()
	if inPacket.IsCmdSize() || inPacket.IsCmdProto() {
		as.handleAgentCmd(sess, inPacket)
		return
	}

	connID := inPacket.GetConnID()
	// show packet content
	//zaplog.S.Debugf("agent@%s: cid: %d, packet: %v, size: %d", sess.ClientAddr(), connID, inPacket, len(inPacket))
	clientRequest := NewRequestFromAgent(inPacket)

	result, isTimeout := as.router.Dispatch(clientRequest)
	if isTimeout {
		//zaplog.S.Errorf(
		//	"agent@%s response timeout: cid: %d, mid: %d, aid: %d",
		//	sess.ClientAddr(), connID, clientRequest.MID, clientRequest.AID)
	}
	//zaplog.S.Debugf(
	//	"agent@%s response: cid: %d, mid: %d, aid: %d, out: [%v]",
	//	sess.ClientAddr(), connID, clientRequest.MID, clientRequest.AID, result,
	//)

	dataload, err := result.Marshal()
	if err != nil {
		//zaplog.S.Errorf(
		//	"agent@%s marshal error: cid: %d, mid: %d, aid: %d, err: %v",
		//	sess.ClientAddr(), connID, clientRequest.MID, clientRequest.AID, err)
		return
	}

	// don't encrypt the data, an agent server will do this
	outPacket := packet.NewFromData(dataload, nil, packet.NoneCompresser)
	outPacket.SetConnID(connID)
	outPacket.SetProtoMID(inPacket.GetProtoMID())
	outPacket.SetProtoAID(inPacket.GetProtoAID())
	outPacket.SetProtoVer(inPacket.GetProtoVer())

	// zaplog.S.Debugf(
	//	"agent@%s response: cid: %d, mid: %d, aid: %d, out: %v",
	//	as.sess.ClientAddr(), connID, clientRequest.MID, clientRequest.AID, outPacket,
	// )

	_, err = sess.Write(outPacket)
	if err != nil {
		//zaplog.S.Errorf(
		//	"write agent@%s response: cid: %d, mid: %d, aid: %d, err: %v",
		//	sess.ClientAddr(), connID, clientRequest.MID,
		//	clientRequest.AID, zeroutil.ParseNetError(err))
	}
}
