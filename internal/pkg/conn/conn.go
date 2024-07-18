package conn

import (
	"context"
	"net"
)

type ContextConn struct {
	net.Conn
	ctx context.Context
}

func NewContextConn(ctx context.Context, c net.Conn) *ContextConn {
	return &ContextConn{
		Conn: c,
		ctx:  ctx,
	}
}

type ContextGetter interface {
	Context() context.Context
}

func NewContextFromConn(conn net.Conn) context.Context {
	if c, ok := conn.(ContextGetter); ok {
		return c.Context()
	}
	return context.Background()
}

type Conn struct {
	Conn net.Conn
	Rb   []byte
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{Conn: conn}
}
