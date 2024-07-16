package file

import (
	"github.com/wuyunlong/tun/pkg/util"
	"golang.org/x/time/rate"
	"sync"
)

type DBUtils struct {
	JsonDB *JsonDB
}

var (
	once sync.Once
	db   *DBUtils
)

func GetDB() *DBUtils {
	once.Do(func() {
		jsonDB := NewJsonDB(".")
		jsonDB.LoadClients()
		jsonDB.LoadTunnels()
		jsonDB.LoadHosts()
		db = &DBUtils{JsonDB: jsonDB}
	})
	return db
}

func (d *DBUtils) NewClient(c *Client) {
	if c.Id == 0 {
		c.Id = d.JsonDB.GetClientID()
	}

reset:

	if c.Key == "" {
		c.Key, _ = util.RandID()
	}

	if _, ok := d.JsonDB.Clients.Load(c.Id); ok {
		c.Key = ""
		goto reset
	}

	if c.Rate > 0 {
		c.RateLimiter = rate.NewLimiter(rate.Limit(c.Rate), c.Rate)
	} else {
		c.RateLimiter = rate.NewLimiter(rate.Limit(2<<32), 2<<32)
	}

	d.JsonDB.Clients.Store(c.Id, c)
	d.JsonDB.SaveClients()
}

func (d *DBUtils) NewTunnel(t *Tunnel) {
	if t.Id == 0 {
		t.Id = d.JsonDB.GetTunnelID()
	}

	d.JsonDB.Tunnels.Store(t.Id, t)
	d.JsonDB.SaveTunnels()
}

func (d *DBUtils) GetIdByKey(k string) (id int, ok bool) {
	d.JsonDB.Clients.Range(func(key, value any) bool {
		v := value.(*Client)
		if v.Key == k {
			id = v.Id
			ok = true
			return false
		}
		return true
	})
	return
}

func (d *DBUtils) NewHost(t *Host) {
	if t.Id == 0 {
		t.Id = d.JsonDB.GetHostID()
	}

	d.JsonDB.Hosts.Store(t.Id, t)
	d.JsonDB.SaveHosts()
}