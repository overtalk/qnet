package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/overtalk/qnet"
)

type TestHead struct{ qnet.BasicNetHead }

func (head *TestHead) GetMsgID() uint16     { return 1 }
func (head *TestHead) GetMsgLength() uint32 { return 1 }

func TestMsgHeadDeserializer(data []byte) (qnet.NetHead, error) { return &TestHead{}, nil }
func TestMsgSerializeFunc(msg qnet.NetHead) []byte              { return nil }

var (
	ws  bool
	tcp bool
	udp bool
)

func main() {
	flag.BoolVar(&tcp, "tcp", true, "tcp")
	flag.BoolVar(&udp, "udp", false, "udp")
	flag.BoolVar(&ws, "ws", false, "websocket")
	flag.Parse()

	serverOptions := []qnet.Option{
		qnet.WithMsgRouter(qnet.HeadLength(0), TestMsgHeadDeserializer, TestMsgSerializeFunc),
		qnet.WithConnectHook(
			func(session qnet.Session) {
				sessionID := session.GetSessionID()
				fmt.Println("[ConnectHook] session id = ", sessionID)
			},
		),
		qnet.WithDisconnectHook(
			func(session qnet.Session) {
				sessionID := session.GetSessionID()
				fmt.Println("[DisconnectHook] disconnect, session id = ", sessionID)
			},
		),
	}

	if udp {
		fmt.Println("udp")
		serverOptions = append(serverOptions, qnet.WithURL("udp://127.0.0.1:9999"))
	} else {
		if ws {
			fmt.Println("ws")
			serverOptions = append(serverOptions, qnet.WithURL("ws://127.0.0.1:9999/ws"))
		} else {
			if tcp {
				fmt.Println("tcp")
				serverOptions = append(serverOptions, qnet.WithURL("tcp://127.0.0.1:9999"))
			}
		}
	}

	svr, err := qnet.NewServer(serverOptions...)
	if err != nil {
		log.Fatal(err)
	}

	if err := svr.RegisterMsgHandler(1, echo); err != nil {
		log.Fatal(err)
	}

	if err := svr.Start(); err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(time.Second)
	}
}

func echo(session qnet.Session, msg *qnet.NetMsg) *qnet.NetMsg {
	fmt.Printf("[%d] - %s\n", session.GetSessionID(), string(msg.GetMsg()))
	return msg
}
