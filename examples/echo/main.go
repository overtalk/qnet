package main

import (
	"flag"
	"fmt"
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

	url := ""

	if udp {
		fmt.Println("udp")
		url = "udp://127.0.0.1:9999"
	} else {
		if ws {
			fmt.Println("ws")
			url = "ws://127.0.0.1:9999/ws"
		} else {
			if tcp {
				fmt.Println("tcp")
				url = "tcp://127.0.0.1:9999"
			}
		}
	}

	svr := qnet.NewNServer().
		SetOnClosedFunc(func(c qnet.Conn, err error) qnet.Action {
			fmt.Println("on close ")
			return qnet.None
		}).
		SetOnInitCompleteFunc(func(server interface{}) qnet.Action {
			fmt.Println("init ")
			return qnet.None
		}).
		SetReactFunc(func(frame []byte, c qnet.Conn) ([]byte, qnet.Action) {
			fmt.Println(c.Context(), string(frame))
			return frame, qnet.Close
		}).
		SetTickFunc(func() (time.Duration, qnet.Action) {
			fmt.Println("ticker func")
			return time.Second, qnet.None
		}).SetURL(url)

	svr.Start()
}
