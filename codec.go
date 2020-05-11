package qnet

import (
	"encoding/binary"
	"errors"
	"log"

	"github.com/panjf2000/gnet"
)

type Logic func(msg *NetMsg, c Conn) *NetMsg

type INetMsgCodec interface {
	RegisterMsgHandler(id uint16, handler Logic)
	DecodeNetMsg(data []byte) (*NetMsg, error)
	EncodeNetMsg(msg *NetMsg) []byte
	React(frame []byte, c Conn) (out []byte, action Action)
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

func (cs *CSCodec) DecodeNetMsg(data []byte) (*NetMsg, error) {
	head, err := CSMsgHeadDeserializer(data[:6])
	if err != nil {
		return nil, err
	}
	return NewNetMsg(head, data[6:]), nil
}

func (cs *CSCodec) EncodeNetMsg(msg *NetMsg) []byte {
	return append(CSMsgHeadSerializer(msg), msg.GetMsg()...)
}

func (cs *CSCodec) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	msg, err := cs.DecodeNetMsg(frame)
	if err != nil {
		//todo:error handle
		return nil, gnet.None
	}

	handler, isExist := cs.handlerMap[msg.GetMsgID()]
	if !isExist {
		//todo: error handle, return some to client
		return nil, gnet.None
	}

	// todo : async
	if retMsg := handler(msg, c); retMsg != nil {
		return cs.EncodeNetMsg(retMsg), gnet.None
	}

	return nil, gnet.None
}
