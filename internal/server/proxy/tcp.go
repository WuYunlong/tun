package proxy

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
	"tun/internal/pkg/conn"
	"tun/internal/pkg/log"
)

func init() {
	RegisterProxyFactory("tcp", NewTCPProxy)
}

type TCPProxy struct {
	*BaseProxy
}

func NewTCPProxy(baseProxy *BaseProxy) Proxy {
	return &TCPProxy{
		BaseProxy: baseProxy,
	}
}

func (tcp *TCPProxy) Run() (remoteAddr string, err error) {
	var listen net.Listener
	listen, err = net.Listen("tcp", net.JoinHostPort("0.0.0.0", strconv.Itoa(tcp.tunnel.Port)))
	if err != nil {
		return
	}
	tcp.listeners = append(tcp.listeners, listen)
	remoteAddr = fmt.Sprintf("0.0.0.0:%d", tcp.tunnel.Port)
	tcp.Start()
	return
}

func (tcp *TCPProxy) Close() {
	tcp.BaseProxy.Close()
}

func (tcp *TCPProxy) Start() {
	for _, ln := range tcp.listeners {
		go func(ln net.Listener) {
			var tempDelay time.Duration
			for {
				c, err := ln.Accept()
				if err != nil {
					if err, ok := err.(interface{ Temporary() bool }); ok && err.Temporary() {
						if tempDelay == 0 {
							tempDelay = 5 * time.Millisecond
						} else {
							tempDelay *= 2
						}
						if mx := 1 * time.Second; tempDelay > mx {
							tempDelay = mx
						}
						time.Sleep(tempDelay)
						continue
					}
					fmt.Println("listener is closed ...")
					return
				}
				go tcp.handleUserTCPConnection(c)
			}
		}(ln)
	}
}

func (tcp *TCPProxy) handleUserTCPConnection(userConn net.Conn) {
	defer userConn.Close()

	workConn, err := tcp.GetWorkConn(userConn)
	if err != nil {
		return
	}
	defer workConn.Close()

	var local io.ReadWriteCloser = workConn
	inCount, outCount, _ := conn.Join(local, userConn)
	log.Infof("run in [%d], out [%d]", inCount, outCount)
}
