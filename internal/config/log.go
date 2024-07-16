package config

import "tun/pkg/util"

type Log struct {
	To              string `yaml:"to,omitempty"`
	Level           string `yaml:"level,omitempty"`
	MaxDays         int    `yaml:"maxDays,omitempty"`
	DisableLogColor bool   `yaml:"disableLogColor,omitempty"`
}

func (l *Log) Complete() {
	l.To = util.EmptyOr(l.To, "console")
	l.Level = util.EmptyOr(l.Level, "info")
	l.MaxDays = util.EmptyOr(l.MaxDays, 30)
	l.DisableLogColor = util.EmptyOr(l.DisableLogColor, false)
}
