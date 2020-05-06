package server

import (
	"fmt"
	"io"

	"github.com/overtalk/qnet/base"
)

type MsgHandler func(session Session, msg *base.NetMsg) *base.NetMsg

type msgRouter struct {
	length              base.HeadLength
	handlerMap          map[uint16]MsgHandler
	headDeserializeFunc base.HeadDeserializeFunc
}

func newMsgRouter(length base.HeadLength, decoderFunc base.HeadDeserializeFunc) *msgRouter {
	ret := &msgRouter{
		length:              length,
		handlerMap:          make(map[uint16]MsgHandler),
		headDeserializeFunc: decoderFunc,
	}

	return ret
}

func (router *msgRouter) registerMsgHandler(id uint16, handler MsgHandler) error {
	if _, isExist := router.handlerMap[id]; isExist {
		return fmt.Errorf("message id %d is already registered", id)
	}

	router.handlerMap[id] = handler
	return nil
}

func (router *msgRouter) streamMsgHandler(session Session) {
	for {
		// decode head
		headerBytes := make([]byte, router.length)
		if _, err := io.ReadFull(session, headerBytes); err != nil {
			break
		}

		head, err := router.headDeserializeFunc(headerBytes)
		if err != nil {
			continue
		}

		bodyByte := make([]byte, head.GetMsgLength())
		if _, err := io.ReadFull(session, bodyByte); err != nil {
			//todo: add log
			fmt.Println(err)
			break
		}

		msg := base.NewNetMsg(head, bodyByte)

		f, flag := router.handlerMap[msg.GetMsgID()]
		if !flag {
			// todo: add log
			continue
		}

		// todo: handle return
		f(session, msg)
	}
}

func (router *msgRouter) packetMsgHandler(session Session) {
	for {
		packet, err := session.ReadPacket()
		if err != nil {
			// todo: add log
			break
		}

		// decode head
		head, err := router.headDeserializeFunc(packet[:router.length])
		if err != nil {
			continue
		}

		msg := base.NewNetMsg(head, packet[router.length:])

		f, flag := router.handlerMap[msg.GetMsgID()]
		if !flag {
			// todo: add log
			continue
		}

		// todo: handle return
		f(session, msg)
	}
}
