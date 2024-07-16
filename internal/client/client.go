package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"
	"tun/internal/pkg/clog"
	"tun/internal/pkg/msg"
	"tun/internal/pkg/wait"
	pnet "tun/pkg/net"
	"tun/pkg/version"
)

type cancelErr struct {
	Err error
}

func (e cancelErr) Error() string {
	return e.Err.Error()
}

type Client struct {
	token                    string
	ctx                      context.Context
	cancel                   context.CancelCauseFunc
	ctl                      *Control
	ctlMu                    sync.RWMutex
	gracefulShutdownDuration time.Duration
	connectorCreator         func(context.Context, *SeverCfg) Connector
}

func NewClient(token string) *Client {
	tc := &Client{
		token: token,
	}
	tc.connectorCreator = NewConnector
	return tc
}

func (tc *Client) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	tc.ctx = clog.NewContext(ctx, clog.FromContextSafe(ctx))
	tc.cancel = cancel
	pnet.SetDefaultDNSAddress("223.5.5.5")

	tc.loopLoginUntilSuccess()
	if tc.ctl == nil {
		cancelCause := cancelErr{}
		_ = errors.As(context.Cause(tc.ctx), &cancelCause)
		return fmt.Errorf("login to the server failed: %v. With loginFailExit enabled, no additional retries will be attempted", cancelCause.Err)
	}

	go tc.keepControllerWorking()

	<-tc.ctx.Done()
	tc.stop()
	return nil
}

func (tc *Client) Close() {
	tc.GracefulClose(0)
}

func (tc *Client) GracefulClose(d time.Duration) {
	tc.gracefulShutdownDuration = d
	tc.cancel(nil)
}

func (tc *Client) loopLoginUntilSuccess() {
	cl := clog.FromContextSafe(tc.ctx)
	loginFunc := func() (bool, error) {
		cl.Infof("try to connect to server...")
		conn, connector, err := tc.login()
		if err != nil {
			cl.Warnf("connect to server error: %v", err)
			return false, err
		}

		sessionCtx := &SessionContext{
			Conn:      conn,
			Token:     tc.token,
			Connector: connector,
		}
		ctl, err := NewControl(tc.ctx, sessionCtx)
		if err != nil {
			conn.Close()
			cl.Errorf("NewControl error: %v", err)
			return false, nil
		}

		ctl.SetInWorkConnCallback()

		ctl.Run()

		tc.ctlMu.Lock()
		tc.ctl = ctl
		tc.ctlMu.Unlock()
		return true, nil
	}
	bfm := wait.NewFastBackoffManager(wait.FastBackoffOptions{
		Duration:    time.Second,
		Factor:      2,
		Jitter:      0.1,
		MaxDuration: 10 * time.Second,
	})
	wait.BackoffUntil(loginFunc, bfm, true, tc.ctx.Done())
}

func (tc *Client) login() (conn net.Conn, connector Connector, err error) {
	log := clog.FromContextSafe(tc.ctx)
	// TODO 在api服务器获取
	cfg := &SeverCfg{ServerAddr: "127.0.0.1", ServerPort: 10002}
	connector = tc.connectorCreator(tc.ctx, cfg)
	if err = connector.Open(); err != nil {
		return nil, nil, err
	}
	conn, err = connector.Connect()
	if err != nil {
		return
	}

	loginMsg := &msg.Login{
		Version:   version.Full(),
		Arch:      runtime.GOARCH,
		Os:        runtime.GOOS,
		Timestamp: time.Now().Unix(),
		Token:     tc.token,
	}

	if err = msg.WriteMsg(conn, loginMsg); err != nil {
		return
	}
	var loginRespMsg msg.LoginResp
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err = msg.ReadMsgInto(conn, &loginRespMsg); err != nil {
		return
	}
	_ = conn.SetReadDeadline(time.Time{})

	if loginRespMsg.Error != "" {
		err = fmt.Errorf("%s", loginRespMsg.Error)
		return
	}

	tc.token = loginRespMsg.Token
	log.AddPrefix(clog.LogPrefix{Name: "Token", Value: loginRespMsg.Token})
	log.Infof("login to server success, get token is [%s]", loginRespMsg.Token)
	return
}

func (tc *Client) keepControllerWorking() {
	<-tc.ctl.Done()

	wait.BackoffUntil(func() (bool, error) {
		tc.loopLoginUntilSuccess()
		if tc.ctl != nil {
			<-tc.ctl.Done()
			return false, errors.New("control is closed and try another loop")
		}
		return false, nil
	}, wait.NewFastBackoffManager(
		wait.FastBackoffOptions{
			Duration:        time.Second,
			Factor:          2,
			Jitter:          0.1,
			MaxDuration:     20 * time.Second,
			FastRetryCount:  3,
			FastRetryDelay:  200 * time.Millisecond,
			FastRetryWindow: time.Minute,
			FastRetryJitter: 0.5,
		},
	), true, tc.ctx.Done())
}

func (tc *Client) stop() {
	tc.ctlMu.Lock()
	defer tc.ctlMu.Unlock()
	if tc.ctl != nil {
		tc.ctl.GracefulClose(tc.gracefulShutdownDuration)
		tc.ctl = nil
	}
}
