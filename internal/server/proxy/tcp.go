package proxy

import (
	"net"
	"tun/internal/pkg/conn"
	"tun/internal/pkg/log"
)

type TcpServer struct {
	BaseServer
	listener net.Listener
}

func NewTcpServer() *TcpServer {
	return &TcpServer{}
}

func (tcp *TcpServer) Start() error {
	address := net.JoinHostPort("0.0.0.0", "99")
	return conn.NewTcpListenerAndProcess(address, func(c net.Conn) {
		// 检查流量和连接数
		if err := tcp.CheckFlowAndConnNum(tcp.tunnel.Client); err != nil {
			log.Warnf("client id %d, task id %d,error %s, when tcp connection", tcp.tunnel.Client.Id, tcp.tunnel.Id, err.Error())
			c.Close()
		}
	}, &tcp.listener)
}

func (tcp *TcpServer) Close() error {
	return nil
}
