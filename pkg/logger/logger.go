package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"time"
)

var (
	Info  = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	Debug = log.New(os.Stdout, "DEBUG\t", log.Ldate|log.Ltime)
	Error = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
)

type LogFile struct {
	file string
}

func createLogFile(l LogFile) {
	f, err := os.OpenFile(l.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		return
	}

}

func initZapLog() *zap.Logger {
	t := time.Now()
	formattedTime := t.Format("2006-01-02")
	logfile := LogFile{
		"log/" + formattedTime + ".log",
	}

	createLogFile(logfile)

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{
		logfile.file,
		"stdout",
	}
	config.ErrorOutputPaths = []string{
		logfile.file,
		"stderr",
	}
	zapLogger, _ := config.Build()
	return zapLogger
}

func Init() {
	logManager := initZapLog()
	zap.ReplaceGlobals(logManager)
	defer func(logManager *zap.Logger) {
		err := logManager.Sync()
		if err != nil {

		}
	}(logManager) // flushes buffer, if any
}
