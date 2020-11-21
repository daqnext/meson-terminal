package terminallogger

import (
	//baselogger "github.com/daqnext/meson-common/common/terminallogger"
	//"github.com/daqnext/meson-terminal/terminal/manager/config"
	//"github.com/sirupsen/logrus"
	//"io"
	//"os"
	//"testing"

	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"testing"
)

func InitLogger() {
	testing.Init()

	log := logrus.New()

	fileWriter := logger.LogFileWriter{}
	log.SetOutput(io.MultiWriter(&fileWriter, os.Stdout))

	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000000",
		FullTimestamp:   true,
	})

	// must disable when production env
	log.SetReportCaller(false)

	// set log level
	loglevel := config.GetInt("loglevel")
	log.Level = logrus.Level(loglevel)

	logger.BaseLogger = log
}

//func Debug(msg string, params ...interface{}) {
//	if terminallogger == nil {
//		return
//	}
//	terminallogger.WithFields(baselogger.SliceToFields(params)).Debug(msg)
//}
//
//func Info(msg string, params ...interface{}) {
//	if terminallogger == nil {
//		return
//	}
//	terminallogger.WithFields(baselogger.SliceToFields(params)).Info(msg)
//}
//
//func Warn(msg string, params ...interface{}) {
//	if terminallogger == nil {
//		return
//	}
//	terminallogger.WithFields(baselogger.SliceToFields(params)).Warn(msg)
//}
//
//func Error(msg string, params ...interface{}) {
//	if terminallogger == nil {
//		return
//	}
//	terminallogger.WithFields(baselogger.SliceToFields(params)).Error(msg)
//}
//
//func Fatal(msg string, params ...interface{}) {
//	if terminallogger == nil {
//		return
//	}
//	terminallogger.WithFields(baselogger.SliceToFields(params)).Fatal(msg)
//}
