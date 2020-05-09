package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn, err := net.Dial("udp", "127.0.0.1:9999")
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			p := make([]byte, 2048)
			_, err = bufio.NewReader(conn).Read(p)
			if err != nil {
				log.Println("read:", err)
				time.Sleep(time.Second)
			} else {
				log.Printf("recv: %s", p)
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			_, err := conn.Write([]byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			return
		}
	}
}
