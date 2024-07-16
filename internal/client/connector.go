package client

import (
	"context"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
	"tun/internal/pkg/clog"
	"tun/pkg/tmux"
)

type SeverCfg struct {
	ServerAddr string
	ServerPort int
}

type Connector interface {
	Open() error
	Connect() (net.Conn, error)
	Close() error
}

type TmuxConnector struct {
	ctx       context.Context
	cfg       *SeverCfg
	session   *tmux.Session
	closeOnce sync.Once
}

func (t *TmuxConnector) Open() error {
	log := clog.FromContextSafe(t.ctx)
	conn, err := t.realConnect()
	if err != nil {
		return err
	}
	tmuxCfg := tmux.DefaultConfig()
	tmuxCfg.KeepAliveInterval = 10 * time.Second
	tmuxCfg.LogOutput = io.Discard
	tmuxCfg.MaxStreamWindowSize = 10 * 1024 * 1024
	session, err := tmux.Client(conn, tmuxCfg)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}
	t.session = session
	return nil
}

func (t *TmuxConnector) Connect() (net.Conn, error) {
	log := clog.FromContextSafe(t.ctx)
	if t.session != nil {
		stream, err := t.session.OpenStream()
		if err != nil {
			log.Errorf("%v", err)
			return nil, err
		}
		return stream, nil
	}

	return t.realConnect()
}

func (t *TmuxConnector) Close() error {
	t.closeOnce.Do(func() {
		if t.session != nil {
			_ = t.session.Close()
		}
	})
	return nil
}

func (t *TmuxConnector) realConnect() (net.Conn, error) {
	log := clog.FromContextSafe(t.ctx)
	address := net.JoinHostPort(t.cfg.ServerAddr, strconv.Itoa(t.cfg.ServerPort))
	log.Infof("server address is [%s]", address)
	return net.Dial("tcp", address)
}

func NewConnector(ctx context.Context, cfg *SeverCfg) Connector {
	return &TmuxConnector{
		ctx: ctx,
		cfg: cfg,
	}
}
