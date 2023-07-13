package wmeow

import (
	"fmt"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/zap"
	"strings"
)

type CustomLogger waLog.Logger

type zapLogger struct {
	module  string
	minimum int
}

func (s *zapLogger) Errorf(msg string, args ...interface{}) {
	if levelToInt["ERROR"] < s.minimum {
		return
	}
	zap.S().Errorf("[WhatsApp]\t"+msg, args...)
}
func (s *zapLogger) Warnf(msg string, args ...interface{}) {
	if levelToInt["WARN"] < s.minimum {
		return
	}
	zap.S().Warnf("[WhatsApp]\t"+msg, args...)
}
func (s *zapLogger) Infof(msg string, args ...interface{}) {
	if levelToInt["INFO"] < s.minimum {
		return
	}
	zap.S().Infof("[WhatsApp]\t"+msg, args...)
}
func (s *zapLogger) Debugf(msg string, args ...interface{}) {
	if levelToInt["DEBUG"] < s.minimum {
		return
	}
	zap.S().Debugf("[WhatsApp]\t"+msg, args...)
}
func (s *zapLogger) Sub(module string) waLog.Logger {
	return &zapLogger{module: fmt.Sprintf("%s/%s", s.module, module), minimum: s.minimum}
}

var levelToInt = map[string]int{
	"":      -1,
	"DEBUG": 0,
	"INFO":  1,
	"WARN":  2,
	"ERROR": 3,
}

func InitZapLogger(module string, minLevel string) waLog.Logger {
	return &zapLogger{
		module:  module,
		minimum: levelToInt[strings.ToUpper(minLevel)],
	}
}
