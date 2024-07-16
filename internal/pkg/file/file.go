package file

import (
	"encoding/json"
	"errors"
	"golang.org/x/time/rate"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
)

type JsonDB struct {
	Clients        sync.Map
	Tunnels        sync.Map
	Hosts          sync.Map
	RunPath        string
	LastClientId   int32
	LastTunnelId   int32
	LastHostId     int32
	ClientFilePath string
	TunnelFilePath string
	HostFilePath   string
	clientMu       sync.Mutex
	tunnelMu       sync.Mutex
	hostMu         sync.Mutex
}

func NewJsonDB(runPath string) *JsonDB {
	return &JsonDB{
		Clients:        sync.Map{},
		Tunnels:        sync.Map{},
		Hosts:          sync.Map{},
		RunPath:        runPath,
		LastClientId:   0,
		LastTunnelId:   0,
		LastHostId:     0,
		ClientFilePath: filepath.Join(runPath, "conf", "clients.json"),
		TunnelFilePath: filepath.Join(runPath, "conf", "tunnels.json"),
		HostFilePath:   filepath.Join(runPath, "conf", "hosts.json"),
	}
}

func (s *JsonDB) LoadClients() {
	var posts []*Client
	bytes, err := s.loadJson(s.ClientFilePath)
	if err != nil {
		return
	}
	json.Unmarshal(bytes, &posts)

	for _, post := range posts {
		if post.Rate > 0 {
			post.RateLimiter = rate.NewLimiter(rate.Limit(post.Rate), post.Rate)
		} else {
			post.RateLimiter = rate.NewLimiter(rate.Limit(2<<32), 2<<32)
		}

		s.Clients.Store(post.Id, post)
		if post.Id > int(s.LastClientId) {
			s.LastClientId = int32(post.Id)
		}
	}
}

func (s *JsonDB) LoadTunnels() {
	var posts []*Tunnel
	bytes, err := s.loadJson(s.TunnelFilePath)
	if err != nil {
		return
	}
	json.Unmarshal(bytes, &posts)

	for _, post := range posts {
		if post.Client, err = s.GetClient(post.Client.Id); err != nil {
			return
		}
		s.Tunnels.Store(post.Id, post)
		if post.Id > int(s.LastTunnelId) {
			s.LastTunnelId = int32(post.Id)
		}
	}
}

func (s *JsonDB) LoadHosts() {
	var posts []*Host
	bytes, err := s.loadJson(s.HostFilePath)
	if err != nil {
		return
	}
	json.Unmarshal(bytes, &posts)

	for _, post := range posts {
		if post.Client, err = s.GetClient(post.Client.Id); err != nil {
			return
		}
		s.Hosts.Store(post.Id, post)
		if post.Id > int(s.LastHostId) {
			s.LastHostId = int32(post.Id)
		}
	}
}

func (s *JsonDB) SaveClients() {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	var posts []*Client
	s.Clients.Range(func(key, value interface{}) bool {
		posts = append(posts, value.(*Client))
		return true
	})
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Id < posts[j].Id
	})
	bytes, _ := json.MarshalIndent(posts, "", "  ")
	s.saveJson(s.ClientFilePath, bytes)
}

func (s *JsonDB) SaveTunnels() {
	s.tunnelMu.Lock()
	defer s.tunnelMu.Unlock()

	var posts []*Tunnel
	s.Tunnels.Range(func(key, value interface{}) bool {
		posts = append(posts, value.(*Tunnel))
		return true
	})
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Id < posts[j].Id
	})
	bytes, _ := json.MarshalIndent(posts, "", "  ")
	s.saveJson(s.TunnelFilePath, bytes)
}

func (s *JsonDB) SaveHosts() {
	s.hostMu.Lock()
	defer s.hostMu.Unlock()

	var posts []*Host
	s.Hosts.Range(func(key, value interface{}) bool {
		posts = append(posts, value.(*Host))
		return true
	})
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Id < posts[j].Id
	})
	bytes, _ := json.MarshalIndent(posts, "", "  ")
	s.saveJson(s.HostFilePath, bytes)
}

func (s *JsonDB) GetClient(id int) (c *Client, err error) {
	if v, ok := s.Clients.Load(id); ok {
		c = v.(*Client)
		return
	}
	return nil, errors.New("未找到客户端")
}

func (s *JsonDB) GetClientID() int {
	return int(atomic.AddInt32(&s.LastClientId, 1))
}

func (s *JsonDB) GetTunnelID() int {
	return int(atomic.AddInt32(&s.LastTunnelId, 1))
}

func (s *JsonDB) GetHostID() int {
	return int(atomic.AddInt32(&s.LastHostId, 1))
}

func (s *JsonDB) loadJson(file string) ([]byte, error) {
	return os.ReadFile(file)
}

func (s *JsonDB) saveJson(file string, bytes []byte) {
	tmpFile := file + ".tmp"
	defer os.Remove(tmpFile)

	if err := os.WriteFile(tmpFile, bytes, 0644); err != nil {
		return
	}

	if err := os.Rename(tmpFile, file); err != nil {
		return
	}
}
