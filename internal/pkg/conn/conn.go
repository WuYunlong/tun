package conn

import (
	"context"
	"net"
	"time"
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

func (c *Conn) Read(b []byte) (n int, err error) {
	if c.Rb != nil {
		if len(c.Rb) > 0 {
			n = copy(b, c.Rb)
			c.Rb = c.Rb[n:]
			return
		}
		c.Rb = nil
	}
	return c.Conn.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return c.Conn.Write(b)
}

func (c *Conn) Close() error {
	return c.Conn.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.Conn.SetWriteDeadline(t)
}
