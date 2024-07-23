package proxy

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
	"tun/internal/pkg/conn"
	"tun/internal/pkg/msg"
	"tun/internal/pkg/util"
	"tun/pkg/pool"
)

func init() {
	RegisterProxyFactory("udp", NewUDPProxy)
}

type UDPProxy struct {
	*BaseProxy
	sendCh       chan *msg.UDPPacket
	readCh       chan *msg.UDPPacket
	udpConn      *net.UDPConn
	workConn     net.Conn
	isClosed     bool
	checkCloseCh chan int
}

func NewUDPProxy(baseProxy *BaseProxy) Proxy {
	return &UDPProxy{
		BaseProxy: baseProxy,
	}
}

func (udp *UDPProxy) Run() (remoteAddr string, err error) {
	remoteAddr = net.JoinHostPort("0.0.0,0", strconv.Itoa(udp.BaseProxy.tunnel.Port))
	var addr *net.UDPAddr
	addr, err = net.ResolveUDPAddr("udp", net.JoinHostPort("0.0.0.0", "8080"))
	if err != nil {
		return
	}

	var udpConn *net.UDPConn
	udpConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return
	}

	udp.udpConn = udpConn
	udp.sendCh = make(chan *msg.UDPPacket, 1024)
	udp.readCh = make(chan *msg.UDPPacket, 1024)

	workConnReaderFn := func(c net.Conn) {
		for {
			var (
				rawMsg msg.Message
				errRet error
			)
			_ = c.SetReadDeadline(time.Now().Add(time.Duration(60) * time.Second))
			if rawMsg, errRet = msg.ReadMsg(c); errRet != nil {
				fmt.Println("read from workConn for udp error")
				_ = c.Close()
				_ = util.PanicToError(func() {
					udp.checkCloseCh <- 1
				})
				return
			}
			if err = c.SetReadDeadline(time.Time{}); err != nil {
				fmt.Println("set read deadline error")
			}
			switch m := rawMsg.(type) {
			case *msg.UDPPacket:
				if errRet = util.PanicToError(func() {
					udp.readCh <- m
				}); errRet != nil {
					_ = c.Close()
					return
				}
			}
		}
	}

	workConnSenderFn := func(c net.Conn, ctx context.Context) {
		var errRet error
		for {
			select {
			case udpMsg, ok := <-udp.sendCh:
				if !ok {
					fmt.Println("sender goroutine for udp work connection closed")
					return
				}
				if errRet = msg.WriteMsg(c, udpMsg); errRet != nil {
					fmt.Println("sender goroutine for udp work connection closed")
					_ = c.Close()
					return
				}
				continue
			case <-ctx.Done():
				fmt.Println("sender goroutine for udp work connection closed")
				return
			}
		}
	}

	go func() {
		time.Sleep(500 * time.Millisecond)
		for {
			var workConn net.Conn
			// var err error
			workConn, err = udp.GetWorkConnFromPool(nil, nil)
			if err != nil {
				time.Sleep(1 * time.Second)
				select {
				case _, ok := <-udp.checkCloseCh:
					if !ok {
						return
					}
				default:
				}
				continue
			}
			if udp.workConn != nil {
				udp.workConn.Close()
			}
			var rwc io.ReadWriteCloser = workConn
			ctx, cancel := context.WithCancel(context.Background())
			udp.workConn = conn.WrapReadWriteCloserToConn(rwc, workConn)
			go workConnReaderFn(udp.workConn)
			go workConnSenderFn(udp.workConn, ctx)
			_, ok := <-udp.checkCloseCh
			cancel()
			if !ok {
				return
			}
		}
	}()

	go func() {
		udp.ForwardUserConn(udpConn, udp.readCh, udp.sendCh, 1024)
		udp.Close()
	}()

	return
}

func (udp *UDPProxy) Close() {
	udp.mu.Lock()
	defer udp.mu.Unlock()

	if !udp.isClosed {
		udp.isClosed = true
		udp.BaseProxy.Close()

		if udp.workConn != nil {
			udp.workConn.Close()
		}
		udp.udpConn.Close()
		close(udp.checkCloseCh)
		close(udp.sendCh)
		close(udp.readCh)
	}
}

func (udp *UDPProxy) ForwardUserConn(udpConn *net.UDPConn, readCh <-chan *msg.UDPPacket, sendCh chan<- *msg.UDPPacket, bufSize int) {
	go func() {
		for udpMsg := range readCh {
			bytes, err := base64.StdEncoding.DecodeString(udpMsg.Content)
			if err != nil {
				continue
			}
			_, _ = udpConn.WriteToUDP(bytes, udpMsg.RemoteAddr)
		}
	}()

	buf := pool.GetBuf(bufSize)
	defer pool.PutBuf(buf)
	for {
		n, remoteAddr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		udpMsg := &msg.UDPPacket{
			Content:    base64.StdEncoding.EncodeToString(buf[:n]),
			LocalAddr:  nil,
			RemoteAddr: remoteAddr,
		}
		select {
		case sendCh <- udpMsg:
		default:
		}
	}
}
