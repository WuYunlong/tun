package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:10004")
	if err != nil {
		panic(err)
	}

	conn.Write([]byte("Hello World ..."))

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			panic(err)
		}
		fmt.Println("服务端响应:", string(buffer[:n]))
		time.Sleep(1 * time.Second)
		conn.Write([]byte("Hello World ..."))
	}
}
