package server

import (
	"context"
	"net"
)

type Server struct {
	ln     net.Listener
	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer() *Server {
	ts := new(Server)
	ts.ctx = context.Background()

	return ts
}

func (ts *Server) Run(ctx context.Context) {
	ts.ctx, ts.cancel = context.WithCancel(ctx)

	<-ts.ctx.Done()
	ts.stop()
}

func (ts *Server) stop() {

}
