package server

import "sync"

type ControlManager struct {
	ctls map[string]*Control
	mu   sync.RWMutex
}

func NewControlManager() *ControlManager {
	cm := new(ControlManager)
	cm.ctls = make(map[string]*Control)
	return cm
}

func (cm *ControlManager) Add(token string, c *Control) (old *Control) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	var ok bool
	old, ok = cm.ctls[token]
	if ok {
		old.Replaced(c)
	}
	cm.ctls[token] = c
	return nil
}

func (cm *ControlManager) Del(token string, c *Control) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if o, ok := cm.ctls[token]; ok && o == c {
		o.Close()
		delete(cm.ctls, token)
	}
}

func (cm *ControlManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, c := range cm.ctls {
		c.Close()
	}
	cm.ctls = make(map[string]*Control)
	return nil
}

func (cm *ControlManager) GetByToken(token string) (c *Control, ok bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	c, ok = cm.ctls[token]
	return
}
