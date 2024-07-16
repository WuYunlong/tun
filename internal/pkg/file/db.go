package file

import (
	"golang.org/x/time/rate"
	"sync"
	"tun/pkg/util"
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

	if c.Token == "" {
		c.Token, _ = util.RandID()
	}

	if _, ok := d.JsonDB.Clients.Load(c.Id); ok {
		c.Token = ""
		goto reset
	}

	if c.Rate > 0 {
		c.RateLimiter = rate.NewLimiter(rate.Limit(c.Rate), c.Rate)
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

func (d *DBUtils) GetIdByToken(token string) (id int, ok bool) {
	d.JsonDB.Clients.Range(func(key, value any) bool {
		v := value.(*Client)
		if v.Token == token {
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
