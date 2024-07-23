package conn

import (
	"context"
	"errors"
	"io"
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

type WrapReadWriteCloserConn struct {
	io.ReadWriteCloser
	underConn  net.Conn
	remoteAddr net.Addr
}

func WrapReadWriteCloserToConn(rwc io.ReadWriteCloser, underConn net.Conn) *WrapReadWriteCloserConn {
	return &WrapReadWriteCloserConn{
		ReadWriteCloser: rwc,
		underConn:       underConn,
	}
}

func (w *WrapReadWriteCloserConn) LocalAddr() net.Addr {
	if w.underConn != nil {
		return w.underConn.LocalAddr()
	}
	return (*net.TCPAddr)(nil)
}

func (w *WrapReadWriteCloserConn) SetRemoteAddr(addr net.Addr) {
	w.remoteAddr = addr
}

func (w *WrapReadWriteCloserConn) RemoteAddr() net.Addr {
	if w.remoteAddr != nil {
		return w.remoteAddr
	}
	if w.underConn != nil {
		return w.underConn.RemoteAddr()
	}
	return (*net.TCPAddr)(nil)
}

func (w *WrapReadWriteCloserConn) SetDeadline(t time.Time) error {
	if w.underConn != nil {
		return w.underConn.SetDeadline(t)
	}
	return &net.OpError{Op: "set", Net: "wrap", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}

func (w *WrapReadWriteCloserConn) SetReadDeadline(t time.Time) error {
	if w.underConn != nil {
		return w.underConn.SetReadDeadline(t)
	}
	return &net.OpError{Op: "set", Net: "wrap", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}

func (w *WrapReadWriteCloserConn) SetWriteDeadline(t time.Time) error {
	if w.underConn != nil {
		return w.underConn.SetWriteDeadline(t)
	}
	return &net.OpError{Op: "set", Net: "wrap", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}
