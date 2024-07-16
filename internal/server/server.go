package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
	"tun/internal/config"
	"tun/internal/pkg/clog"
	"tun/internal/pkg/conn"
	"tun/internal/pkg/file"
	"tun/internal/pkg/log"
	"tun/internal/pkg/msg"
	"tun/internal/pkg/util"
	"tun/pkg/tmux"
	"tun/pkg/version"
)

type Server struct {
	ln     net.Listener
	pm     *ControlManager
	cfg    *config.ServerConfig
	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(cfg *config.ServerConfig) (ts *Server, err error) {
	ts = &Server{
		ctx: context.Background(),
		pm:  NewControlManager(),
		cfg: cfg,
	}

	address := net.JoinHostPort(cfg.BindAddr, strconv.Itoa(cfg.BindPort))
	ts.ln, err = net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("create server listener error, %v", err)
	}
	log.Infof("tuns tcp listen on %s", address)

	return
}

func (ts *Server) Run(ctx context.Context) {
	ts.ctx, ts.cancel = context.WithCancel(ctx)
	ts.HandleListener(ts.ln)
	<-ts.ctx.Done()
	ts.stop()
}

func (ts *Server) HandleListener(ln net.Listener) {
	for {
		var c net.Conn
		var err error
		c, err = ln.Accept()
		if err != nil {
			log.Warnf("Listener for incoming connections from client closed")
			return
		}
		cl := clog.New()
		ctx := context.Background()
		c = conn.NewContextConn(clog.NewContext(ctx, cl), c)
		go func(ctx context.Context, tunConn net.Conn) {
			tmuxCnf := tmux.DefaultConfig()
			tmuxCnf.KeepAliveInterval = time.Duration(10) * time.Second
			tmuxCnf.LogOutput = io.Discard
			tmuxCnf.MaxStreamWindowSize = 10 * 1024 * 1024
			var session *tmux.Session
			session, err = tmux.Server(tunConn, tmuxCnf)
			if err != nil {
				log.Warnf("Failed to create mux connection: %v", err)
				tunConn.Close()
				return
			}
			for {
				var stream *tmux.Stream
				stream, err = session.AcceptStream()
				if err != nil {
					log.Debugf("Accept new mux stream error: %v", err)
					session.Close()
					return
				}
				go ts.handleConnection(ctx, stream)
			}
		}(ctx, c)
	}
}

func (ts *Server) RegisterControl(ctlConn net.Conn, loginMsg *msg.Login) error {
	if err := ts.checkToken(loginMsg); err != nil {
		return err
	}

	ctx := conn.NewContextFromConn(ctlConn)
	cl := clog.FromContextSafe(ctx)
	cl.AppendPrefix(loginMsg.Token)
	ctx = clog.NewContext(ctx, cl)

	cl.Infof(
		"client login info: ip [%s] version [%s] os [%s] arch[%s]",
		ctlConn.RemoteAddr().String(),
		loginMsg.Token,
		loginMsg.Os,
		loginMsg.Arch)

	sessionCtx := &SessionContext{
		Conn:  ctlConn,
		Token: loginMsg.Token,
	}
	ctl, err := NewControl(ctx, sessionCtx)
	if err != nil {
		cl.Warnf("create new controller error: %v", err)
		return fmt.Errorf("unexpected error when creating new controller")
	}

	if o := ts.pm.Add(loginMsg.Token, ctl); o != nil {
		o.WaitClosed()
	}

	ctl.Start()

	return nil
}

func (ts *Server) RegisterWorkConn(workConn net.Conn, newMsg *msg.NewWorkConn) error {
	// log := conn.NewLogFromConn(workConn)
	c, exist := ts.pm.GetByToken(newMsg.Token)
	if !exist {
		log.Warnf("No client control found for run id [%s]", newMsg.Token)
		return fmt.Errorf("no client control found for run id [%s]", newMsg.Token)
	}
	return c.RegisterWorkConn(workConn)
}

func (ts *Server) checkToken(loginMsg *msg.Login) error {
	if loginMsg.Token == "" {
		return fmt.Errorf("token is empty")
	}
	if _, ok := file.GetDB().GetIdByToken(loginMsg.Token); !ok {
		return fmt.Errorf("token is invalid")
	}
	return nil
}

func (ts *Server) handleConnection(ctx context.Context, conn net.Conn) {
	cl := clog.FromContextSafe(ctx)
	var (
		rawMsg msg.Message
		err    error
	)

	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	if rawMsg, err = msg.ReadMsg(conn); err != nil {
		log.Tracef("Failed to read message: %v", err)
		conn.Close()
		return
	}
	_ = conn.SetReadDeadline(time.Time{})

	switch m := rawMsg.(type) {
	case *msg.Login:
		err = ts.RegisterControl(conn, m)
		if err != nil {
			cl.Warnf("register control error: %v", err)
			_ = msg.WriteMsg(conn, &msg.LoginResp{
				Version: version.Full(),
				Error:   util.GenerateResponseErrorString("register control error", err, ts.cfg.SendErrorToClient),
			})
			conn.Close()
		}
	case *msg.NewWorkConn:
		ts.RegisterWorkConn(conn, m)
	default:
		log.Warnf("Error message type for the new connection [%s]", conn.RemoteAddr().String())
		conn.Close()
	}
}

func (ts *Server) stop() {
	if ts.ln != nil {
		ts.ln.Close()
		ts.ln = nil
	}
}
