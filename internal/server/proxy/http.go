package proxy

type HttpServer struct {
}

func NewHttpServer() *HttpServer {
	return &HttpServer{}
}

func (http *HttpServer) Start() error {
	return nil
}

func (http *HttpServer) Close() error {
	return nil
}
