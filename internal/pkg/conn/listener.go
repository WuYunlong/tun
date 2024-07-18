package conn

import "net"

func NewTcpListenerAndProcess(address string, f func(conn net.Conn), listener *net.Listener) (err error) {
	*listener, err = net.Listen("tcp", address)
	if err != nil {
		return
	}
	Accept(*listener, f)
	return
}

func Accept(ln net.Listener, f func(net.Conn)) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		if conn == nil {
			break
		}
		go f(conn)
	}
}
