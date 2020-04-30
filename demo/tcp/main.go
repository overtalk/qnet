package main

import (
	"bufio"
	"fmt"
	"github.com/overtalk/qnet/server"
	"io"
	"log"
	"time"
)

func main() {
	svr, err := server.NewServerFromString("tcp://127.0.0.1:9999", handler)
	if err != nil {
		log.Fatal(err)
	}

	svr.AddConnectHook(func(session server.Session) {
		sessionID := session.GetSessionID()
		fmt.Println("session id = ", sessionID)
	}).AddConnectHook(func(session server.Session) {
		sessionID := session.GetSessionID()
		fmt.Println("hhh, session id = ", sessionID)
	}).AddDisconnectHook(func(session server.Session) {
		sessionID := session.GetSessionID()
		fmt.Println("disconnect, session id = ", sessionID)
	})

	if err := svr.Start(); err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(time.Second)
	}
}

func handler(session server.Session) {
	fmt.Println("in handler ")
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
		fmt.Println(string(line))
		session.Write(line)
	}
}
