// +build  debug

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package log

import (
	"github.com/qlcchain/go-qlc/config"
	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewDevelopment()
	Root      *zap.SugaredLogger
)

func init() {
	Root = logger.Sugar()
}

func InitLog(config *config.Config) error {
	//logFolder := config.LogDir()
	//err := util.CreateDirIfNotExist(logFolder)
	//if err != nil {
	//	return err
	//}
	//logfile, _ := filepath.Abs(filepath.Join(logFolder, logfile))
	//lumlog := &lumberjack.Logger{
	//	Filename:   logfile,
	//	MaxSize:    10, // megabytes
	//	MaxBackups: 10,
	//	MaxAge:     28, // days
	//	Compress:   true,
	//	LocalTime:  true,
	//}
	//w := zapcore.AddSync(lumlog)
	//
	//core := zapcore.NewCore(
	//	zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
	//	w,
	//	zap.InfoLevel,
	//)
	//logger = zap.New(core)
	//logger.Info("log service init successful")
	return nil
}

//NewLogger create logger by name
func NewLogger(name string) *zap.SugaredLogger {
	return logger.Sugar().Named(name)
}
