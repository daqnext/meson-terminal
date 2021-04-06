package terminallogger

import (
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

func init() {
	InitLogger()
}

func InitLogger() {
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
