// +build  !debug

/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package log

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
)

const (
	logfile = "qlc.log"
)

var (
	logger *zap.Logger
	Root   *zap.SugaredLogger
)

func init() {
	//atom := zap.NewAtomicLevelAt(zap.WarnLevel)
	//encoderCfg := zap.NewDevelopmentEncoderConfig()
	//
	//defaultLogger := zap.New(zapcore.NewCore(
	//	zapcore.NewJSONEncoder(encoderCfg),
	//	zapcore.Lock(os.Stdout),
	//	atom,
	//))
	defaultLogger, _ := zap.NewDevelopment()
	logger = defaultLogger
	Root = defaultLogger.Sugar().Named("chain")
}

func Setup(cfg *config.Config) (err error) {
	logFolder := cfg.LogDir()
	err = util.CreateDirIfNotExist(logFolder)
	if err != nil {
		return
	}
	logfile, _ := filepath.Abs(filepath.Join(logFolder, logfile))
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logfile,
		MaxSize:    10, // megabytes
		MaxBackups: 10,
		MaxAge:     28, // days
		Compress:   true,
		LocalTime:  true,
	})
	l := zap.ErrorLevel
	if err := l.Set(cfg.LogLevel); err != nil {
		fmt.Println(err)
	}
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleDebugging, l),
		zapcore.NewCore(zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}), w, l),
	)

	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return nil
}

func Teardown() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

//NewLogger create logger by name
func NewLogger(name string) *zap.SugaredLogger {
	return logger.Sugar().Named(name)
}
