package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	var listen net.Listener
	var err error
	listen, err = net.Listen("tcp", "0.0.0.0:8989")
	for {
		var conn net.Conn
		conn, err = listen.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	for {
		// 读取客户端发送的数据
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println("读取数据失败:", err)
			break
		}
		fmt.Println(string(buffer[:n]))
		time.Sleep(1 * time.Second)
		// 将接收到的数据发送回客户端
		_, _ = conn.Write([]byte(time.Now().Format("2006-01-02 15:04:05") + " "))
		_, err = conn.Write(buffer[:n])
		if err != nil {
			log.Println("发送数据失败:", err)
			break
		}
	}
}
