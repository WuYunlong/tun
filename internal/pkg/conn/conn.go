package conn

import (
	"context"
	"net"
	"tun/internal/pkg/clog"
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

func NewLogFromConn(conn net.Conn) *clog.Logger {
	if c, ok := conn.(ContextGetter); ok {
		return clog.FromContextSafe(c.Context())
	}
	return clog.New()
}
