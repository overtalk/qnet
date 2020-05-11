package main

import (
	"fmt"
	"github.com/overtalk/qnet/ggnet"
	"github.com/panjf2000/gnet"
	"log"
)

func main() {
	svr, err := ggnet.NewQNetServer(
		ggnet.WithReact(echo),
		ggnet.WithURL("udp://127.0.0.1:9999"),
		ggnet.WithOnInitComplete(func(server gnet.Server) gnet.Action {
			fmt.Println("OnInitComplete")
			return 0
		}),
		ggnet.WithOnOpened(func(c gnet.Conn) ([]byte, gnet.Action) {
			fmt.Printf("[%v] OnOpened\n", c.Context())
			return nil, 0
		}),
		ggnet.WithOnClosed(func(c gnet.Conn, err error) gnet.Action {
			fmt.Println("OnClosed")
			return 0
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	svr.Start()
}

func echo(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	fmt.Println(c.LocalAddr(), c.RemoteAddr())
	fmt.Println(c.Context(), string(frame))
	out = frame
	return
}
