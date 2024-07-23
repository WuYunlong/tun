package msg

import "net"

const (
	TypeLogin         = '1'
	TypeLoginResp     = '2'
	TypeReqWorkConn   = '3'
	TypeNewWorkConn   = '4'
	TypeStartWorkConn = '5'
	TypeUdpPacket     = '6'
)

type Login struct {
	Version   string `json:"version,omitempty"`
	Token     string `json:"token,omitempty"`
	Os        string `json:"os,omitempty"`
	Arch      string `json:"arch,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

type LoginResp struct {
	Version string `json:"version,omitempty"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ReqWorkConn struct{}

type NewWorkConn struct {
	Token     string `json:"token,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

type StartWorkConn struct {
	Id      int    `json:"id,omitempty"`
	Remark  string `json:"remark,omitempty"`
	SrcAddr string `json:"src_addr,omitempty"`
	SrcPort int16  `json:"src_port,omitempty"`
	DstAddr string `json:"dst_addr,omitempty"`
	DstPort int16  `json:"dst_port,omitempty"`
	Target  string `json:"target,omitempty"`
	Error   string `json:"error,omitempty"`
}

type UDPPacket struct {
	Content    string       `json:"c,omitempty"`
	LocalAddr  *net.UDPAddr `json:"l,omitempty"`
	RemoteAddr *net.UDPAddr `json:"r,omitempty"`
}

var msgTypeMap = map[byte]interface{}{
	TypeLogin:         Login{},
	TypeLoginResp:     LoginResp{},
	TypeReqWorkConn:   ReqWorkConn{},
	TypeNewWorkConn:   NewWorkConn{},
	TypeStartWorkConn: StartWorkConn{},
	TypeUdpPacket:     UDPPacket{},
}
