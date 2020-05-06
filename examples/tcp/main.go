package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/overtalk/qnet/base"
	"github.com/overtalk/qnet/server"
)

func main() {
	svr, err := server.NewServer(
		server.WithURL("tcp://127.0.0.1:9999"),
		//server.WithHandler(handler),
		server.WithMsgRouter(base.CSHeadLength, base.CSMsgHeadDeserializer),
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

	svr.RegisterMsgHandler(1, messageHandler)

	if err := svr.Start(); err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(time.Second)
	}
}

func messageHandler(session base.Session, msg *base.NetMsg) *base.NetMsg {
	return nil
}

func handler(session base.Session) {
	id := session.GetSessionID()

	fmt.Println("[handler ", id, "]")
	r := bufio.NewReader(session)
	for {
		line, err := r.ReadBytes(byte('\n'))
		switch err {
		case nil:
			break
		case io.EOF:
			fmt.Println("EOF")
			return
		default:
			fmt.Println("ERROR", err)
			return
		}
		fmt.Printf("[%d] %s", id, string(line))
		session.Write(line)
	}
}
