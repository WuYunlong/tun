package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
)

func init() {
	RegisterProxyFactory("http", NewHttpProxy)
}

type HttpProxy struct {
	*BaseProxy
	httpServer *http.Server
}

func NewHttpProxy(baseProxy *BaseProxy) Proxy {
	return &HttpProxy{
		BaseProxy: baseProxy,
	}
}

func (s *HttpProxy) Run() (remoteAddr string, err error) {
	address := net.JoinHostPort("0.0.0.0", strconv.Itoa(s.BaseProxy.tunnel.Port))
	remoteAddr = address
	s.httpServer = &http.Server{
		Addr: address,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Scheme = "http"
			s.handleTunneling(w, r)
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	go func() {
		var ln net.Listener
		ln, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: s.BaseProxy.tunnel.Port})
		if err != nil {
			fmt.Println("listenTCP error")
			os.Exit(0)
		}
		err = s.httpServer.Serve(ln)
		if err != nil {
			fmt.Println("httpServer serve error")
			os.Exit(0)
		}
	}()
	return
}

func (s *HttpProxy) Close() {
}

func (s *HttpProxy) handleTunneling(w http.ResponseWriter, r *http.Request) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	c, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	s.handleHttp(c, r)
}

func (s *HttpProxy) handleHttp(c net.Conn, r *http.Request) {
	defer func() {
		c.Close()
	}()
	// fmt.Println(r.Host)
	_, _ = c.Write([]byte("Hello World ..."))
}
