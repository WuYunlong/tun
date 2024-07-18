package proxy

type HttpsServer struct {
}

func NewHttpsServer() *HttpsServer {
	return &HttpsServer{}
}

func (https *HttpsServer) Start() error {
	return nil
}

func (https *HttpsServer) Close() error {
	return nil
}
