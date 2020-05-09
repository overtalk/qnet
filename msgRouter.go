package qnet

import (
	"fmt"
)

type msgRouter struct {
	length              HeadLength
	headDeserializeFunc HeadDeserializeFunc
	headSerializeFunc   HeadSerializeFunc
	handlerMap          map[uint16]MsgHandler
}

func newMsgRouter(length HeadLength, decoderFunc HeadDeserializeFunc, headSerializeFunc HeadSerializeFunc) *msgRouter {
	ret := &msgRouter{
		length:              length,
		headDeserializeFunc: decoderFunc,
		headSerializeFunc:   headSerializeFunc,
		handlerMap:          make(map[uint16]MsgHandler),
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

func (router *msgRouter) handle(session Session) {
	for {
		msg, addr, err := session.GetNetMsg(router.length, router.headDeserializeFunc)
		if err != nil {
			//todo : error handler
			break
		}

		f, flag := router.handlerMap[msg.GetMsgID()]
		if !flag {
			// todo: add log
			continue
		}

		if retMsg := f(session, msg); retMsg != nil {
			session.SendNetMsg(router.headSerializeFunc, retMsg, addr)
		}
	}
}

//// getHandler return handler for each connection/session
//func (router *msgRouter) getHandler(t ProtoType) (func(session Session), error) {
//	switch t {
//	case ProtoTypeTcp:
//		return router.tcpMsgHandler, nil
//	case ProtoTypeUdp:
//		return router.udpMsgHandler, nil
//	case ProtoTypeWs:
//		return router.wsMsgHandler, nil
//	default:
//		return nil, fmt.Errorf("invalid protocolType : %s", t)
//	}
//}
//
//func (router *msgRouter) tcpMsgHandler(session Session) {
//	for {
//		// decode head
//		headerBytes := make([]byte, router.length)
//		if _, err := io.ReadFull(session, headerBytes); err != nil {
//			break
//		}
//
//		head, err := router.headDeserializeFunc(headerBytes)
//		if err != nil {
//			continue
//		}
//
//		bodyByte := make([]byte, head.GetMsgLength())
//		if _, err := io.ReadFull(session, bodyByte); err != nil {
//			//todo: add log
//			fmt.Println(err)
//			break
//		}
//
//		msg := NewNetMsg(head, bodyByte)
//
//		f, flag := router.handlerMap[msg.GetMsgID()]
//		if !flag {
//			// todo: add log
//			continue
//		}
//
//		if retMsg := f(session, msg); retMsg != nil {
//			bytes := router.headSerializeFunc(msg)
//			bytes = append(bytes, msg.GetMsg()...)
//			session.Write(bytes)
//		}
//	}
//}
//
//func (router *msgRouter) wsMsgHandler(session Session) {
//	for {
//		msg, err := session.GetNetMsg()
//		if err != nil {
//			//todo : error handler
//			break
//		}
//
//		f, flag := router.handlerMap[msg.GetMsgID()]
//		if !flag {
//			// todo: add log
//			continue
//		}
//
//		if retMsg := f(session, msg); retMsg != nil {
//			bytes := router.headSerializeFunc(msg)
//			bytes = append(bytes, msg.GetMsg()...)
//			session.Write(bytes)
//		}
//	}
//}
//
//func (router *msgRouter) udpMsgHandler(session Session) {
//	for {
//		packet := make([]byte, 1024)
//		n, remoteAddr, err := session.ReadFromUDP(packet)
//		if err != nil {
//			fmt.Println(err)
//			// todo: add log
//			break
//		}
//
//		// decode head
//		head, err := router.headDeserializeFunc(packet[:router.length])
//		if err != nil {
//			continue
//		}
//
//		msg := NewNetMsg(head, packet[router.length:n])
//
//		f, flag := router.handlerMap[msg.GetMsgID()]
//
//		if !flag {
//			// todo: add log
//			continue
//		}
//
//		// for udp, goroutine per packet
//		go func(session Session, handler MsgHandler, msg *NetMsg, remoteAddr *net.UDPAddr) {
//			if retMsg := handler(session, msg); retMsg != nil {
//				bytes := router.headSerializeFunc(msg)
//				bytes = append(bytes, msg.GetMsg()...)
//				session.WriteToUDP(bytes, remoteAddr)
//			}
//		}(session, f, msg, remoteAddr)
//	}
//}
