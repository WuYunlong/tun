package proxy

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"tun/internal/pkg/file"
	"tun/internal/pkg/msg"
)

var proxyFactoryRegistry = map[string]func(*BaseProxy) Proxy{}

func RegisterProxyFactory(proxyConfType string, factory func(*BaseProxy) Proxy) {
	proxyFactoryRegistry[proxyConfType] = factory
}

type GetWorkConnFn func(token string) (net.Conn, error)

type Proxy interface {
	Run() (remoteAddr string, err error)
	Close()
}

func NewProxy(t *file.Tunnel, f GetWorkConnFn) (pxy Proxy, err error) {
	factory := proxyFactoryRegistry[t.Mode]
	if factory == nil {
		return nil, fmt.Errorf("proxy type not support")
	}
	baseProxt := &BaseProxy{
		id:            t.Id,
		tunnel:        t,
		listeners:     make([]net.Listener, 0),
		getWorkConnFn: f,
	}
	pxy = factory(baseProxt)
	return
}

// BaseProxy 基础
type BaseProxy struct {
	id            int
	tunnel        *file.Tunnel
	listeners     []net.Listener
	getWorkConnFn GetWorkConnFn
	mu            sync.RWMutex
}

func (b *BaseProxy) GetId() int {
	return b.id
}

func (b *BaseProxy) GetRemark() string {
	return b.tunnel.Remark
}

func (b *BaseProxy) GetToken() string {
	return b.tunnel.Client.Token
}

func (b *BaseProxy) GetWorkConnFromPool(src, dst net.Addr) (workConn net.Conn, err error) {
	// 从所有的链接中找到链接
	for i := 0; i < 7; i++ {
		workConn, err = b.getWorkConnFn(b.GetToken())
		var (
			srcAddr    string
			dstAddr    string
			srcPortStr string
			dstPortStr string
			srcPort    int
			dstPort    int
		)

		if src != nil {
			srcAddr, srcPortStr, _ = net.SplitHostPort(src.String())
			srcPort, _ = strconv.Atoi(srcPortStr)
		}
		if dst != nil {
			dstAddr, dstPortStr, _ = net.SplitHostPort(dst.String())
			dstPort, _ = strconv.Atoi(dstPortStr)
		}

		err = msg.WriteMsg(workConn, &msg.StartWorkConn{
			Id:      b.GetId(),
			Remark:  b.GetRemark(),
			SrcAddr: srcAddr,
			SrcPort: int16(srcPort),
			DstAddr: dstAddr,
			DstPort: int16(dstPort),
			Target:  b.tunnel.Target.TargetStr, // TODO 进行负载
			Error:   "",
		})

		if err != nil {
			fmt.Println("failed to send message to work connection from pool")
			workConn.Close()
		} else {
			break
		}
	}
	if err != nil {
		fmt.Println("try to get work connection failed in the end")
		return
	}

	return
}

func (b *BaseProxy) GetWorkConn(userConn net.Conn) (workConn net.Conn, err error) {
	return b.GetWorkConnFromPool(userConn.RemoteAddr(), userConn.LocalAddr())
}

func (b *BaseProxy) Close() {
	for _, ln := range b.listeners {
		ln.Close()
	}
}

// Manager 管理器
type Manager struct {
	proxys map[int]Proxy
	http   Proxy
	https  Proxy
	mu     sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		proxys: make(map[int]Proxy),
	}
}

func (pm *Manager) Add(id int, pxy Proxy) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if _, ok := pm.proxys[id]; ok {
		return fmt.Errorf("proxy id [%d] is already in use", id)
	}
	pm.proxys[id] = pxy
	return nil
}

func (pm *Manager) Exist(id int) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	_, ok := pm.proxys[id]
	return ok
}

func (pm *Manager) Del(id int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.proxys, id)
}

func (pm *Manager) GetById(id int) (pxy Proxy, ok bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	pxy, ok = pm.proxys[id]
	return
}

func (pm *Manager) SetHttp(http Proxy) {
	pm.http = http
}

func (pm *Manager) SetHttps(https Proxy) {
	pm.https = https
}
