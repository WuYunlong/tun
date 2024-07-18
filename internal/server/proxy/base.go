package proxy

import (
	"sync"
	"tun/internal/pkg/file"
)

type Service interface {
	Start() error
	Close() error
}

type BaseServer struct {
	tunnel *file.Tunnel
	sync.Mutex
}

func NewBaseServer(tunnel *file.Tunnel) *BaseServer {
	return &BaseServer{
		tunnel: tunnel,
		Mutex:  sync.Mutex{},
	}
}

func (svr *BaseServer) CheckFlowAndConnNum(client *file.Client) error {
	return nil
}
