package log

import (
	"bytes"
	"os"

	"tun/pkg/log"
)

var (
	TraceLevel = log.TraceLevel
	DebugLevel = log.DebugLevel
	InfoLevel  = log.InfoLevel
	WarnLevel  = log.WarnLevel
	ErrorLevel = log.ErrorLevel
)

var Logger *log.Logger

func init() {
	Logger = log.New(
		log.WithCaller(true),
		log.AddCallerSkip(1),
		log.WithLevel(log.InfoLevel),
	)
}

func InitLogger(logPath string, levelStr string, maxDays int, disableLogColor bool) {
	var options []log.Option
	if logPath == "console" {
		if !disableLogColor {
			options = append(options,
				log.WithOutput(log.NewConsoleWriter(log.ConsoleConfig{
					Colorful: true,
				}, os.Stdout)),
			)
		}
	} else {
		writer := log.NewRotateFileWriter(log.RotateFileConfig{
			FileName: logPath,
			Mode:     log.RotateFileModeDaily,
			MaxDays:  maxDays,
		})
		writer.Init()
		options = append(options, log.WithOutput(writer))
	}

	level, err := log.ParseLevel(levelStr)
	if err != nil {
		level = log.InfoLevel
	}
	options = append(options, log.WithLevel(level))
	Logger = Logger.WithOptions(options...)
}

func Errorf(format string, v ...interface{}) {
	Logger.Errorf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	Logger.Warnf(format, v...)
}

func Infof(format string, v ...interface{}) {
	Logger.Infof(format, v...)
}

func Debugf(format string, v ...interface{}) {
	Logger.Debugf(format, v...)
}

func Tracef(format string, v ...interface{}) {
	Logger.Tracef(format, v...)
}

func Logf(level log.Level, offset int, format string, v ...interface{}) {
	Logger.Logf(level, offset, format, v...)
}

type WriteLogger struct {
	level  log.Level
	offset int
}

func NewWriteLogger(level log.Level, offset int) *WriteLogger {
	return &WriteLogger{
		level:  level,
		offset: offset,
	}
}

func (w *WriteLogger) Write(p []byte) (n int, err error) {
	Logger.Log(w.level, w.offset, string(bytes.TrimRight(p, "\n")))
	return len(p), nil
}
