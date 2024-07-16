package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"tun/internal/pkg/common"
	"tun/pkg/util"
)

type ServerConfig struct {
	BindAddr          string `yaml:"bindAddr,omitempty"`
	BindPort          int    `yaml:"bindPort,omitempty"`
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
	s.SendErrorToClient = util.EmptyOr(s.SendErrorToClient, false)
	s.Log.Complete()
}
