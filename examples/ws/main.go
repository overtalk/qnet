package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/overtalk/qnet/server"
)

func main() {
	svr, err := server.NewServer(
		server.WithURL("ws://127.0.0.1:9999"),
		server.WithHandler(handler),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := svr.Start(); err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(time.Second)
	}
}

func handler(session server.Session) {
	id := session.GetSessionID()

	fmt.Println("[handler ", id, "]")

	for {
		packet, err := session.ReadPacket()
		//line, err := r.ReadBytes(byte('\n'))
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
		fmt.Printf("[%d] %s\n", id, string(packet))
		session.Write(packet)
	}
}
