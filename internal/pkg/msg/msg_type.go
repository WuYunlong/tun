package msg

const (
	TypeLogin         = '1'
	TypeLoginResp     = '2'
	TypeReqWorkConn   = '3'
	TypeNewWorkConn   = '4'
	TypeStartWorkConn = '5'
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
	Error string `json:"error,omitempty"`
}

var msgTypeMap = map[byte]interface{}{
	TypeLogin:         Login{},
	TypeLoginResp:     LoginResp{},
	TypeReqWorkConn:   ReqWorkConn{},
	TypeNewWorkConn:   NewWorkConn{},
	TypeStartWorkConn: StartWorkConn{},
}
