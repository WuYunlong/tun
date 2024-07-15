package server

type ControlManager struct {
	ctls map[string]*Control
}

func NewControlManager() *ControlManager {
	pm := new(ControlManager)
	pm.ctls = make(map[string]*Control)
	return pm
}

func (cm *ControlManager) Add(runID string, c *Control) (ctl *Control) {
	return nil
}

func (cm *ControlManager) Del(runID string, c *Control) {

}

func (cm *ControlManager) GetByID(runID string) (c *Control) {
	return nil
}
