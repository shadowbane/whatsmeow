package dblogger

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type CustomLogLevel logger.LogLevel

const (
	// Silent silent log level
	Silent logger.LogLevel = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
	// Debug debug log level
	Debug
)

type GormLoggerInterface logger.Interface

type ZapLogger struct {
	logger.Writer
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *ZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level

	return &newlogger
}

// Info print info
func (l ZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		zap.S().Infof(fmt.Sprintf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...))
	}
}

// Warn print warn messages
func (l ZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		zap.S().Warnf(fmt.Sprintf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...))
	}
}

// Error print error messages
func (l ZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		zap.S().Errorf(fmt.Sprintf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...))
	}
}

// Trace print sql message
func (l ZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			rows = 0
		}
		zap.S().Errorf("%s\n\t\t\t\tTrace: %s\n\t\t\t\t\tFile: %s\n\t\t\t\t\tTime: %fms\n\t\t\t\t\tRows:%d\n\t\t\t\t\tQuery: %s", err.Error(), l.traceErrStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			rows = 0
		}
		zap.S().Warnf("%s\n\t\t\t\tTrace: %s\n\t\t\t\t\tFile: %s\n\t\t\t\t\tTime: %fms\n\t\t\t\t\tRows:%d\n\t\t\t\t\tQuery: %s", slowLog, l.traceWarnStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
	case l.LogLevel == Debug:
		sql, rows := fc()
		if rows == -1 {
			rows = 0
		}
		zap.S().Debugf("\n\t\t\t\tTrace: %s\n\t\t\t\t\tFile: %s\n\t\t\t\t\tTime: %fms\n\t\t\t\t\tRows:%d\n\t\t\t\t\tQuery: %s", l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}

// ParamsFilter filters params to be logged
func (l ZapLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}

func LogLevelToGormLevel(level string) logger.LogLevel {
	if l, ok := levelToGormLevel[level]; ok {
		return l
	}

	return logger.Silent
}

var levelToGormLevel = map[string]logger.LogLevel{
	"":      Silent,
	"DEBUG": Debug,
	"INFO":  Info,
	"WARN":  Warn,
	"ERROR": Error,
}
