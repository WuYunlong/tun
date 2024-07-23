package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	serverAddr := net.UDPAddr{
		Port: 12345,
		IP:   net.ParseIP("127.0.0.1"), // 服务端IP地址
	}
	conn, err := net.DialUDP("udp", nil, &serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("Connected to server")
	message := []byte("Hello World ...")
	_, err = conn.Write(message)
	if err != nil {
		log.Fatal(err)
	}
	for {
		buffer := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("服务端响应:", string(buffer[:n]))
		time.Sleep(1 * time.Second)
		message = []byte("Hello World ...")
		_, err = conn.Write(message)
		if err != nil {
			log.Fatal(err)
		}
	}
}
