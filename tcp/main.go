//package main
//
//import (
//	"bufio"
//	"fmt"
//	"io"
//	"net"
//	"os"
//)
//
//var port = "0.0.0.0:9001"
//
//func echo(conn net.Conn) {
//	r := bufio.NewReader(conn)
//	for {
//		line, err := r.ReadBytes(byte('\n'))
//		switch err {
//		case nil:
//			break
//		case io.EOF:
//		default:
//			fmt.Println("ERROR", err)
//		}
//		conn.Write(line)
//	}
//}
//
//func main() {
//	l, err := net.Listen("tcp", port)
//	if err != nil {
//		fmt.Println("ERROR", err)
//		os.Exit(1)
//	}
//
//	for {
//		conn, err := l.Accept()
//		if err != nil {
//			fmt.Println("ERROR", err)
//			continue
//		}
//		go echo(conn)
//	}
//
//}

package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

var (
	wg sync.WaitGroup = sync.WaitGroup{} // 等待各个socket连接处理
)

func main() {

	stop_chan := make(chan os.Signal) // 接收系统中断信号
	signal.Notify(stop_chan, os.Interrupt)

	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		<-stop_chan
		fmt.Println("Get Stop Command. Now Stoping...")
		if err = listen.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println("Start listen :8080 ... ")
	for {
		conn, err := listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				fmt.Println(err)
				break
			}
			fmt.Println(err)
			continue
		}
		fmt.Println("Accept ", conn.RemoteAddr())
		wg.Add(1)
		go Handler(conn)
	}

	fmt.Println("xxx")
	wg.Wait() // 等待是否有未处理完socket处理
}

func Handler(conn net.Conn) {
	defer wg.Done()
	defer conn.Close()

	time.Sleep(5 * time.Second)

	conn.Write([]byte("Hello!"))
	fmt.Println("Send hello")
}
