package main

import (
	"fmt"
	"github.com/overtalk/qnet/gnet"
	"log"
)

func main() {
	svr, err := gnet.NewQNetServer(
		gnet.WithReact(echo),
		gnet.WithOnInitComplete(SetOnInitComplete),
		gnet.WithOnOpened(func(c gnet.GNetConn) ([]byte, gnet.GNetAction) {
			fmt.Println("OnOpened")
			return nil, 0
		}),
		gnet.WithOnClosed(func(c gnet.GNetConn, err error) gnet.GNetAction {
			fmt.Println("OnClosed")
			return 0
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	svr.Start()
}

func echo(frame []byte, c gnet.GNetConn) (out []byte, action gnet.GNetAction) {
	fmt.Println(c.Context(), string(frame))
	out = frame
	return
}

func SetOnInitComplete(server gnet.GNetServer) gnet.GNetAction {
	fmt.Println("OnInitComplete")
	return 0
}
