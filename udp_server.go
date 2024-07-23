package main

import (
	"log"
	"net"
	"time"
)

func main() {
	addr := net.UDPAddr{
		Port: 12345,
		IP:   net.ParseIP("0.0.0.0"), // 监听所有网络接口
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("UDP server is listening on :12345")

	for {
		buffer := make([]byte, 1024)
		var remoteAddr *net.UDPAddr
		_, remoteAddr, err = conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading from UDP:", err)
			continue
		}

		log.Printf("Received message from %s: %s\n", remoteAddr, buffer)

		// 延迟一秒
		time.Sleep(1 * time.Second)

		// 发送接收到的消息回给发送者
		var b = append([]byte(time.Now().Format("2006-01-02 15:04:05")+" "), buffer...)
		// _, err = conn.WriteToUDP(), remoteAddr)
		_, err = conn.WriteToUDP(b, remoteAddr)
		if err != nil {
			log.Println("Error writing to UDP:", err)
		}
	}

}
