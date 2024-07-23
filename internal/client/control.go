package client

import (
	"context"
	"net"
	"time"
	"tun/internal/pkg/clog"
	"tun/internal/pkg/conn"
	"tun/internal/pkg/msg"
)

type Control struct {
	ctx           context.Context
	log           *clog.Logger
	sessionCtx    *SessionContext
	doneCh        chan struct{}
	msgDispatcher *msg.Dispatcher
}

func NewControl(ctx context.Context, sessionCtx *SessionContext) (ctl *Control, err error) {
	ctl = &Control{
		ctx:        ctx,
		log:        clog.FromContextSafe(ctx),
		sessionCtx: sessionCtx,
		doneCh:     make(chan struct{}),
	}

	ctl.msgDispatcher = msg.NewDispatcher(sessionCtx.Conn)
	ctl.registerMsgHandlers()

	return
}

func (c *Control) Run() {
	go c.worker()
}

func (c *Control) SetInWorkConnCallback() {
}

func (c *Control) Done() <-chan struct{} {
	return c.doneCh
}

func (c *Control) Close() error {
	return c.GracefulClose(0)
}

func (c *Control) GracefulClose(d time.Duration) error {
	time.Sleep(d)
	c.closeSession()
	return nil
}

func (c *Control) registerMsgHandlers() {
	c.msgDispatcher.RegisterHandler(&msg.ReqWorkConn{}, msg.AsyncHandler(c.handleReqWorkConn))
}

func (c *Control) handleReqWorkConn(_ msg.Message) {
	log := c.log
	workConn, err := c.connectServer()
	if err != nil {
		log.Warnf("start new connection to server error: %v", err)
		return
	}
	m := &msg.NewWorkConn{
		Token:     c.sessionCtx.Token,
		Timestamp: time.Now().Unix(),
	}
	if err = msg.WriteMsg(workConn, m); err != nil {
		log.Warnf("work connection write to server error: %v", err)
		workConn.Close()
		return
	}

	var startWorkConn msg.StartWorkConn
	if err = msg.ReadMsgInto(workConn, &startWorkConn); err != nil {
		log.Tracef("work connection closed before response StartWorkConn message: %v", err)
		workConn.Close()
		return
	}

	if startWorkConn.Error != "" {
		log.Warnf("start new connection to server error: %v", startWorkConn.Error)
		workConn.Close()
		return
	}

	// TODO 转发数据

	dial, err := net.Dial("tcp", startWorkConn.Target)
	if err != nil {
		log.Errorf("connect local [%s] error ...", startWorkConn.Target)
		return
	}
	inCount, outCount, _ := conn.Join(dial, workConn)
	log.Infof("use flow in [%d] out [%d]", inCount, outCount)
	defer dial.Close()
	defer workConn.Close()
}

func (c *Control) connectServer() (net.Conn, error) {
	return c.sessionCtx.Connector.Connect()
}

func (c *Control) worker() {
	go c.heartbeatWorker()
	go c.msgDispatcher.Run()

	<-c.msgDispatcher.Done()
	c.closeSession()
	close(c.doneCh)
}

func (c *Control) heartbeatWorker() {
}

func (c *Control) closeSession() {
	c.sessionCtx.Conn.Close()
	c.sessionCtx.Connector.Close()
}
