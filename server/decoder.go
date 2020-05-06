package server

import (
	"fmt"
	"io"

	"github.com/overtalk/qnet/base"
)

type MsgHandler func(session Session, msg *base.NetMsg) *base.NetMsg

type decoder struct {
	length              base.HeadLength
	handlerMap          map[uint16]MsgHandler
	headDeserializeFunc base.HeadDeserializeFunc
}

func newDecoder(length base.HeadLength, decoderFunc base.HeadDeserializeFunc) *decoder {
	ret := &decoder{
		length:              length,
		handlerMap:          make(map[uint16]MsgHandler),
		headDeserializeFunc: decoderFunc,
	}

	return ret
}

func (decoder *decoder) registerMsgHandler(id uint16, handler MsgHandler) error {
	if _, isExist := decoder.handlerMap[id]; isExist {
		return fmt.Errorf("message id %d is already registered", id)
	}

	decoder.handlerMap[id] = handler
	return nil
}

func (decoder *decoder) handler(session Session) {
	for {
		// decode head
		headerBytes := make([]byte, decoder.length)
		if _, err := io.ReadFull(session, headerBytes); err != nil {
			break
		}

		head, err := decoder.headDeserializeFunc(headerBytes)
		if err != nil {
			continue
		}

		bodyByte := make([]byte, head.GetMsgLength())
		if _, err := io.ReadFull(session, bodyByte); err != nil {
			break
		}

		msg := base.NewNetMsg(head, bodyByte)

		f, flag := decoder.handlerMap[msg.GetMsgID()]
		if !flag {
			// todo: add log
			continue
		}

		// todo: handle return
		f(session, msg)
	}
}
