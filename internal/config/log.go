package config

type Log struct {
	To              string `yaml:"to,omitempty"`
	Level           string `yaml:"level,omitempty"`
	MaxDays         int    `yaml:"max_days,omitempty"`
	DisableLogColor bool   `yaml:"disable_log_color,omitempty"`
}
