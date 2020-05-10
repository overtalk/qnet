package gnet

import (
	"encoding/binary"
	"errors"
	"log"

	"github.com/overtalk/qnet"
	"github.com/panjf2000/gnet"
)

type Logic func(msg *qnet.NetMsg, c GNetConn) *qnet.NetMsg

type INetMsgCodec interface {
	//gnet.ICodec
	RegisterMsgHandler(id uint16, handler Logic)
	DecodeNetMsg(data []byte) (*qnet.NetMsg, error)
	EncodeNetMsg(msg *qnet.NetMsg) []byte
	React(frame []byte, c GNetConn) (out []byte, action GNetAction)
}

type BasicNetMsgCodec struct {
	handlerMap map[uint16]Logic
}

func (codec *BasicNetMsgCodec) RegisterMsgHandler(id uint16, handler Logic) {
	if _, isExist := codec.handlerMap[id]; isExist {
		log.Fatalf("message id %d is already registered", id)
	}

	codec.handlerMap[id] = handler
}

/*
| msg id | msg len |
|    2   |    4    | = 6
*/
type CSCodec struct {
	BasicNetMsgCodec
	byteOrder binary.ByteOrder
}

func NewCSCodec(big bool) *CSCodec {
	if big {
		return &CSCodec{byteOrder: binary.BigEndian}
	}
	return &CSCodec{byteOrder: binary.LittleEndian}
}

func (cs *CSCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) { return buf, nil }

func (cs *CSCodec) Decode(c gnet.Conn) ([]byte, error) {
	headSize, head := c.ReadN(6)
	if headSize != 6 {
		return nil, errors.New("no net message")
	}

	l := int(cs.byteOrder.Uint32(head[2:]))
	bodySize, body := c.ReadN(l)
	if bodySize != l {
		return nil, errors.New("no net message")
	}
	return append(head, body...), nil
}

func (cs *CSCodec) DecodeNetMsg(data []byte) (*qnet.NetMsg, error) {
	head, err := qnet.CSMsgHeadDeserializer(data[:6])
	if err != nil {
		return nil, err
	}
	return qnet.NewNetMsg(head, data[6:]), nil
}

func (cs *CSCodec) EncodeNetMsg(msg *qnet.NetMsg) []byte {
	return append(qnet.CSMsgHeadSerializer(msg), msg.GetMsg()...)
}

func (cs *CSCodec) React(frame []byte, c GNetConn) (out []byte, action GNetAction) {
	msg, err := cs.DecodeNetMsg(frame)
	if err != nil {
		//todo:error handle
		return nil, NoneAction
	}

	handler, isExist := cs.handlerMap[msg.GetMsgID()]
	if !isExist {
		//todo: error handle, return some to client
		return nil, NoneAction
	}

	// todo : async
	if retMsg := handler(msg, c); retMsg != nil {
		return cs.EncodeNetMsg(retMsg), NoneAction
	}

	return nil, NoneAction
}

// ------------------------------------------------------
type SSCodec struct {
	BasicNetMsgCodec
	ByteOrder binary.ByteOrder
}

func (ss *SSCodec) Decode(c gnet.Conn) ([]byte, error)             { return nil, nil }
func (ss *SSCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) { return nil, nil }
