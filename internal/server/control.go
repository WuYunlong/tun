package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime/debug"
	"sync"
	"time"
	"tun/internal/pkg/clog"
	"tun/internal/pkg/msg"
	"tun/pkg/version"
)

type Control struct {
	log           *clog.Logger
	ctx           context.Context
	token         string
	sessionCtx    *SessionContext
	msgDispatcher *msg.Dispatcher
	workConnCh    chan net.Conn
	doneCh        chan struct{}
	mu            sync.RWMutex
}

func NewControl(ctx context.Context, sessionCtx *SessionContext) (c *Control, err error) {
	c = &Control{
		log:        clog.FromContextSafe(ctx),
		ctx:        ctx,
		token:      sessionCtx.Token,
		sessionCtx: sessionCtx,
		doneCh:     make(chan struct{}),
		workConnCh: make(chan net.Conn, 10),
	}
	c.msgDispatcher = msg.NewDispatcher(sessionCtx.Conn)
	c.registerMsgHandlers()
	return
}

func (c *Control) Start() {
	loginRespMsg := &msg.LoginResp{
		Version: version.Full(),
		Token:   c.sessionCtx.Token,
		Error:   "",
	}
	_ = msg.WriteMsg(c.sessionCtx.Conn, loginRespMsg)
	go func() {
		for i := 0; i < 7; i++ {
			_ = c.msgDispatcher.Send(&msg.ReqWorkConn{})
		}
	}()
	go c.worker()
}

func (c *Control) RegisterWorkConn(conn net.Conn) error {
	log := c.log
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("panic error: %v", err)
			c.log.Errorf(string(debug.Stack()))
		}
	}()

	select {
	case c.workConnCh <- conn:
		log.Debugf("new work connection registered")
		return nil
	default:
		log.Debugf("work connection pool is full, discarding")
		return fmt.Errorf("work connection pool is full, discarding")
	}
}

func (c *Control) GetWorkConn() (workConn net.Conn, err error) {
	defer func() {
		if errRes := recover(); errRes != nil {
			c.log.Errorf("panic error: %v", err)
			c.log.Errorf(string(debug.Stack()))
		}
	}()
	var ok bool
	select {
	case workConn, ok = <-c.workConnCh:
		if !ok {
			err = errors.New("control is closed")
			return
		}
		c.log.Debugf("get work connection from pool")
	default:
		if err = c.msgDispatcher.Send(&msg.ReqWorkConn{}); err != nil {
			return nil, fmt.Errorf("control is already closed")
		}
		select {
		case workConn, ok = <-c.workConnCh:
			if !ok {
				err = errors.New("control is closed")
				c.log.Warnf("no work connections available, %v", err)
				return
			}
		case <-time.After(10 * time.Second):
			err = fmt.Errorf("timeout trying to get work connection")
			c.log.Warnf("%v", err)
			return
		}
	}
	_ = c.msgDispatcher.Send(&msg.ReqWorkConn{})
	return
}

func (c *Control) Replaced(newCtl *Control) {
	c.log.Infof("Replaced by client [%s]", newCtl.token)
	c.token = ""
	c.sessionCtx.Conn.Close()
}

func (c *Control) WaitClosed() {
	<-c.doneCh
}

func (c *Control) Close() error {
	c.sessionCtx.Conn.Close()
	return nil
}

func (c *Control) registerMsgHandlers() {

}

func (c *Control) worker() {
	go c.heartbeatWorker()
	go c.msgDispatcher.Run()

	<-c.msgDispatcher.Done()
	c.sessionCtx.Conn.Close()

	c.mu.Lock()
	defer c.mu.Unlock()

	close(c.workConnCh)
	for workConn := range c.workConnCh {
		workConn.Close()
	}

	c.log.Infof("client exit success")
	close(c.doneCh)
}

func (c *Control) heartbeatWorker() {

}
