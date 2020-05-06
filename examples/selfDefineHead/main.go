package main

import (
	"fmt"
	"log"
	"time"

	"github.com/overtalk/qnet/base"
	"github.com/overtalk/qnet/server"
)

type TestHead struct {
	base.BaseNetHead
}

func (head *TestHead) GetMsgID() uint16     { return 1 }
func (head *TestHead) GetMsgLength() uint32 { return 5 }

func TestMsgHeadDeserializer(data []byte) (base.NetHead, error) {
	return &TestHead{}, nil
}

func main() {
	svr, err := server.NewServer(
		server.WithURL("ws://127.0.0.1:9999/ws"),
		server.WithMsgRouter(base.HeadLength(0), TestMsgHeadDeserializer),
		server.WithConnectHook(
			func(session base.Session) {
				sessionID := session.GetSessionID()
				fmt.Println("[ConnectHook] session id = ", sessionID)
			},
		),
		server.WithDisconnectHook(
			func(session base.Session) {
				sessionID := session.GetSessionID()
				fmt.Println("[DisconnectHook] disconnect, session id = ", sessionID)
			},
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := svr.RegisterMsgHandler(1, messageHandler); err != nil {
		log.Fatal(err)
	}

	if err := svr.Start(); err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(time.Second)
	}
}

func messageHandler(session base .Session, msg *base.NetMsg) *base.NetMsg {
	fmt.Println(string(msg.GetMsg()))
	session.Write(msg.GetMsg())
	return nil
}
