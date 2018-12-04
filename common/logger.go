/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package common

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/qlcchain/go-qlc/common/util"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logfile = "qlc.log"
)

var lumlog lumberjack.Logger
var logger *zap.Logger

//NewLogger create logger by name
func NewLogger(name string) *zap.SugaredLogger {
	if logger == nil {
		logFolder := util.QlcDir("log")
		err := util.CreateDirIfNotExist(logFolder)
		if err != nil {
			fmt.Println(err)
		}
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
	_, err := lumlog.Write([]byte(fmt.Sprintf("%s %s [%s] %s %s\n", e.Time.Format(time.RFC3339Nano), e.Level.CapitalString(), e.LoggerName, e.Caller.TrimmedPath(), e.Message)))
	return err
}
