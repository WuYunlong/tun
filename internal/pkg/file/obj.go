package file

import (
	"golang.org/x/time/rate"
	"sync"
	"sync/atomic"
)

type Flow struct {
	In    int64 `json:"in"`    // 流入
	Out   int64 `json:"out"`   // 流出
	Total int64 `json:"total"` // 总流量
	sync.RWMutex
}

func (f *Flow) Add(in int64, out int64) {
	f.Lock()
	defer f.Unlock()
	f.In += in
	f.Out += out
	f.Total += in + out
}

type Target struct {
	TargetStr string   `json:"target_str,omitempty"`
	TargetArr []string `json:"target_arr,omitempty"`
}

type Client struct {
	Id          int           `json:"id"`                 // id
	Token       string        `json:"token"`              // 唯一标识
	Remark      string        `json:"remark"`             // 备注
	Flow        Flow          `json:"flow"`               // 流量
	Rate        int           `json:"rate,omitempty"`     // 限速KB
	Version     string        `json:"version,omitempty"`  // 客户端版本号
	MaxConn     int           `json:"max_conn,omitempty"` // 最大连接数
	NowConn     int32         `json:"now_conn,omitempty"` // 当前连接数
	RateLimiter *rate.Limiter `json:"-"`                  // 限速器
	sync.RWMutex
}

func NewClient(token string) *Client {
	return &Client{
		Id:          0,
		Token:       token,
		Remark:      "",
		Rate:        0,
		RateLimiter: nil,
		RWMutex:     sync.RWMutex{},
	}
}

func (c *Client) CutConn() {
	atomic.AddInt32(&c.NowConn, 1)
}

func (c *Client) AddConn() {
	atomic.AddInt32(&c.NowConn, -1)
}

func (c *Client) GetConn() bool {
	if c.MaxConn == 0 || int(c.NowConn) < c.MaxConn {
		c.CutConn()
		return true
	}
	return false
}

func (c *Client) HasTunnel(t *Tunnel) (exist bool) {
	GetDB().JsonDB.Tunnels.Range(func(key, value interface{}) bool {
		v := value.(*Tunnel)
		if v.Client.Id == t.Id {
			exist = true
			return false
		}
		return true
	})
	return
}

func (c *Client) HasHost(h *Host) (exist bool) {
	GetDB().JsonDB.Hosts.Range(func(key, value any) bool {
		v := value.(*Host)
		if v.Client.Id == h.Id {
			exist = true
			return false
		}
		return true
	})

	return
}

func (c *Client) GetTunnelNum() (num int) {
	GetDB().JsonDB.Tunnels.Range(func(key, value any) bool {
		v := value.(*Tunnel)
		if v.Client.Id == c.Id {
			num++
		}
		return true
	})

	GetDB().JsonDB.Hosts.Range(func(key, value any) bool {
		v := value.(*Host)
		if v.Client.Id == c.Id {
			num++
		}
		return true
	})
	return
}

type Tunnel struct {
	Id     int     `json:"id,omitempty"`
	Mode   string  `json:"mode,omitempty"`
	Port   int     `json:"port,omitempty"`
	Remark string  `json:"remark,omitempty"`
	Target Target  `json:"target,omitempty"`
	Client *Client `json:"client,omitempty"`
}

type Host struct {
	Id     int     `json:"id,omitempty"`
	Remark string  `json:"remark,omitempty"`
	Target Target  `json:"target,omitempty"`
	Client *Client `json:"client,omitempty"`
}
