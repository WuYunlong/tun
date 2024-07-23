package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
	"tun/internal/config"
	"tun/internal/pkg/clog"
	"tun/internal/pkg/conn"
	"tun/internal/pkg/file"
	"tun/internal/pkg/log"
	"tun/internal/pkg/msg"
	"tun/internal/pkg/util"
	"tun/internal/server/proxy"
	"tun/pkg/tmux"
	"tun/pkg/version"
)

type Server struct {
	ln          net.Listener
	pm          *proxy.Manager
	cm          *ControlManager
	cfg         *config.ServerConfig
	ctx         context.Context
	cancel      context.CancelFunc
	OpenClient  chan int
	CloseClient chan int
	OpenTunnel  chan *file.Tunnel
	CloseTunnel chan *file.Tunnel
	RunList     sync.Map
}

func NewServer(cfg *config.ServerConfig) (ts *Server, err error) {
	ts = &Server{
		ctx:         context.Background(),
		pm:          proxy.NewManager(),
		cm:          NewControlManager(),
		cfg:         cfg,
		OpenClient:  make(chan int),
		CloseClient: make(chan int),
		OpenTunnel:  make(chan *file.Tunnel),
		CloseTunnel: make(chan *file.Tunnel),
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
	// 启动所有隧道
	go ts.InitFromFile()
	// go ts.DealTunnel()

	ts.HandleListener(ts.ln)

	<-ts.ctx.Done()

	if ts.ln != nil {
		ts.Close()
	}
}

func (ts *Server) Close() error {
	if ts.ln != nil {
		ts.ln.Close()
		ts.ln = nil
	}
	ts.cm.Close()
	if ts.cancel != nil {
		ts.cancel()
	}
	return nil
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

	if o := ts.cm.Add(loginMsg.Token, ctl); o != nil {
		o.WaitClosed()
	}

	ctl.Start()

	go func() {
		ctl.WaitClosed()
		ts.cm.Del(loginMsg.Token, ctl)
	}()

	return nil
}

func (ts *Server) RegisterWorkConn(workConn net.Conn, newMsg *msg.NewWorkConn) error {
	// log := conn.NewLogFromConn(workConn)
	c, exist := ts.cm.GetByToken(newMsg.Token)
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

// TODO 测试内容
func (ts *Server) GetWorkConn(token string) (workConn net.Conn, err error) {
	c, ok := ts.cm.GetByToken(token)
	if !ok {
		return nil, fmt.Errorf("no client control found for run id [%s]", token)
	}
	// 获取工作链接
	workConn, err = c.GetWorkConn()
	return
}

func (ts *Server) RunTunnel(t *file.Tunnel) (err error) {
	pxy, err := proxy.NewProxy(t, ts.GetWorkConn)
	if err != nil {
		return err
	}

	if t.Mode == "http" {
		ts.pm.SetHttp(pxy)
	} else if t.Mode == "https" {
		ts.pm.SetHttps(pxy)
	} else {
		ts.pm.Add(t.Id, pxy)
	}

	remoteAddr, err := pxy.Run()
	if err != nil {
		return err
	}
	log.Infof("tunnel %s start mode：%s port %d addr %s", t.Remark, t.Mode, t.Port, remoteAddr)
	return nil
}

// TODO 启动隧道
func (ts *Server) InitFromFile() {
	ts.RunTunnel(&file.Tunnel{Id: 0, ClientId: 0, Mode: "http", Port: 80})
	ts.RunTunnel(&file.Tunnel{Id: 0, ClientId: 0, Mode: "https", Port: 443})

	file.GetDB().JsonDB.Tunnels.Range(func(key, value any) bool {
		v := value.(*file.Tunnel)
		ts.RunTunnel(v)
		return true
	})

}
