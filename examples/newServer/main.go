package main

import (
	"fmt"

	"github.com/overtalk/qnet"
)

func main() {
	s := qnet.NewNServer().
		SetURL("tcp://localhost:9999").
		SetOnClosedFunc(func(c qnet.Conn, err error) qnet.Action {
			fmt.Println("on close ")
			return 0
		}).
		SetOnInitCompleteFunc(func(server interface{}) qnet.Action {
			fmt.Println("init ")
			return 0
		}).
		SetReactFunc(func(frame []byte, c qnet.Conn) ([]byte, qnet.Action) {
			fmt.Println(c.Context(), string(frame))
			return frame, 0
		})
	s.Start()
}
