package file

import (
	"golang.org/x/time/rate"
	"sync"
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
	Id          int           `json:"id"`             // id
	Token       string        `json:"token"`          // 唯一标识
	Remark      string        `json:"remark"`         // 备注
	Flow        Flow          `json:"flow"`           // 流量
	Rate        int           `json:"rate,omitempty"` // 限速KB
	RateLimiter *rate.Limiter `json:"-"`              // 限速器
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

type Tunnel struct {
	Id     int     `json:"id,omitempty"`
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
