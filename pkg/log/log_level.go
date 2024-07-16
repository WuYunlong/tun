package log

import (
	"bytes"
	"errors"
	"fmt"
)

var errUnmarshalNilLevel = errors.New("can't unmarshal a nil *Level")

type Level int8

const (
	TraceLevel Level = iota + 1
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

func ParseLevel(text string) (Level, error) {
	var level Level
	err := level.UnmarshalText([]byte(text))
	return level, err
}

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

func (l Level) LogPrefix() string {
	switch l {
	case TraceLevel:
		return "[T]"
	case DebugLevel:
		return "[D]"
	case InfoLevel:
		return "[I]"
	case WarnLevel:
		return "[W]"
	case ErrorLevel:
		return "[E]"
	default:
		return ""
	}
}

func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errUnmarshalNilLevel
	}
	if !l.unmarshalText(text) && !l.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized level: %q", text)
	}
	return nil
}

func (l *Level) unmarshalText(text []byte) bool {
	switch string(text) {
	case "trace", "TRACE":
		*l = TraceLevel
	case "debug", "DEBUG":
		*l = DebugLevel
	case "info", "INFO", "": // make the zero value useful
		*l = InfoLevel
	case "warn", "WARN":
		*l = WarnLevel
	case "error", "ERROR":
		*l = ErrorLevel
	default:
		return false
	}
	return true
}

func (l Level) Enabled(lvl Level) bool {
	return lvl >= l
}
