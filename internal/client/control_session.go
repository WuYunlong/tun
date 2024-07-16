package client

import "net"

type SessionContext struct {
	Token     string
	Conn      net.Conn
	Connector Connector
}
