package server

import "net"

type SessionContext struct {
	Conn  net.Conn
	Token string
}
