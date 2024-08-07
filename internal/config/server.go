package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"tun/internal/pkg/common"
	"tun/pkg/util"
)

type ServerConfig struct {
	BindAddr          string `yaml:"bindAddr,omitempty"`
	BindPort          int    `yaml:"bindPort,omitempty"`
	VhostHttpPort     int    `yaml:"vhostHttpPort,omitempty"`
	VhostHttpsPort    int    `yaml:"vhostHttpsPort,omitempty"`
	SendErrorToClient bool   `yaml:"sendErrorToClient,omitempty"`
	Log               Log    `yaml:"log,omitempty"`
}

func LoadServerConfig(filePath string) (cfg *ServerConfig) {
	cfg = new(ServerConfig)

	if common.FileExists(filePath) {
		bytes, _ := os.ReadFile(filePath)
		_ = yaml.Unmarshal(bytes, &cfg)
	}

	cfg.Complete()

	return
}

func (s *ServerConfig) Complete() {
	s.BindAddr = util.EmptyOr(s.BindAddr, "0.0.0.0")
	s.BindPort = util.EmptyOr(s.BindPort, 10001)
	s.VhostHttpPort = util.EmptyOr(s.VhostHttpPort, 80)
	s.VhostHttpsPort = util.EmptyOr(s.VhostHttpsPort, 443)
	s.SendErrorToClient = util.EmptyOr(s.SendErrorToClient, false)
	s.Log.Complete()
}
