package common

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"time"
)

const (
	logDir  = "~/.qlcchain/logs"
	logfile = "qlc.log"
)

var lumlog lumberjack.Logger
var logger *zap.Logger

func NewLogger(name string) (*zap.SugaredLogger) {
	if logger == nil {
		logFolder, _ := homedir.Expand(logDir)
		createDirIfNotExist(logFolder)

		logfile := filepath.Join(logFolder, logfile)
		lumlog = lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    10, // megabytesÒ
			MaxBackups: 10,
			MaxAge:     28, // days
		}

		logger, _ = zap.NewDevelopment(zap.Hooks(lumberjackZapHook))
	}
	return logger.Sugar().Named(name)
}

func lumberjackZapHook(e zapcore.Entry) error {
	lumlog.Write([]byte(fmt.Sprintf("%s %s [%s] %s %s\n", e.Time.Format(time.RFC3339Nano), e.Level.CapitalString(), e.LoggerName, e.Caller.TrimmedPath(), e.Message)))
	return nil
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("create dir failed: %s", err)
		}
	}
}
