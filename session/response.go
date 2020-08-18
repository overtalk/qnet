package session

import (
	"io"

	"github.com/overtalk/qnet/common"
	"github.com/overtalk/qnet/packet"
)

// Response a response for game request
type Response struct {
	MID    uint8
	AID    uint8
	PVer   uint8
	PFlag  uint8
	Result common.IOutProtocol
}

// WriteTo write some data to a writer
func (rsp *Response) WriteTo(w io.Writer) (int, error) {
	out, err := rsp.Result.Marshal()
	if err != nil {
		return 0, err
	}
	// zaplog.S.Debugf("mid: %d, aid: %d, data: %v", rsp.MID, rsp.AID, out)
	outPacket := packet.NewFromData(out, nil, packet.NoneCompresser)
	outPacket.SetConnID(0)
	outPacket.SetProtoMID(rsp.MID)
	outPacket.SetProtoAID(rsp.AID)
	outPacket.SetProtoVer(rsp.PVer)
	outPacket.SetDataFlag(rsp.PFlag)
	outPacket.Encrypt(packet.XORCrypto)
	return w.Write(outPacket)
}
