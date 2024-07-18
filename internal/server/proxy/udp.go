package proxy

type UdpServer struct {
}

func NewUdpServer() *UdpServer {
	return &UdpServer{}
}

func (udp *UdpServer) Start() error {
	return nil
}

func (udp *UdpServer) Close() error {
	return nil
}
