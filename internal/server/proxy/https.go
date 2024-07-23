package proxy

func init() {
	RegisterProxyFactory("https", NewHttpsProxy)
}

type HttpsProxy struct {
	*BaseProxy
}

func NewHttpsProxy(baseProxy *BaseProxy) Proxy {
	return &HttpsProxy{
		BaseProxy: baseProxy,
	}
}

func (https *HttpsProxy) Run() (remoteAddr string, err error) {
	return
}

func (https *HttpsProxy) Close() {
}
