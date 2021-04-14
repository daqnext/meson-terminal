package terminallogger

import (
	"fmt"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/domainmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"time"
)

var FileRequestLogger *logrus.Logger

func init() {
	InitDefaultLogger()
}

func InitDefaultLogger() {
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

func InitFileRequestLogger() {
	log := logrus.New()

	fileWriter := logger.LogFileWriter{
		RootDir:         "./requestRecord",
		OnLogFileChange: UploadFileRequestLog,
		MaxSize:         1024 * 3, //only for test
	}
	log.SetOutput(&fileWriter)

	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000000",
		FullTimestamp:   true,
	})

	// must disable when production env
	log.SetReportCaller(false)

	// set log level
	log.Level = logrus.Level(5)

	FileRequestLogger = log
}

var FileRequestLogFormatter = func(param gin.LogFormatterParams) string {
	//var statusColor, methodColor, resetColor string
	//if param.IsOutputColor() {
	//	statusColor = param.StatusCodeColor()
	//	methodColor = param.MethodColor()
	//	resetColor = param.ResetColor()
	//}

	spendTimeUs := param.Latency.Microseconds()
	bindName, _ := param.Keys["bindName"]

	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}

	return fmt.Sprintf("{\"requestTime\":\"%v\",\"statusCode\":%3d,\"spendTimeUs\":%d,\"latency\":\"%s\",\"clientIp\":\"%s\",\"method\":\"%s\",\"bindName\":%#v, \"path\":%#v,\"errorMessage\":\"%s\"}\n",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		param.StatusCode,
		spendTimeUs,
		param.Latency,
		param.ClientIP,
		param.Method,
		bindName,
		param.Path,
		param.ErrorMessage,
	)

	//return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
	//	param.TimeStamp.Format("2006/01/02 - 15:04:05"),
	//	statusColor, param.StatusCode, resetColor,
	//	param.Latency,
	//	param.ClientIP,
	//	methodColor, param.Method, resetColor,
	//	param.Path,
	//	param.ErrorMessage,
	//)
}

func FileRequestLoggerMiddleware() gin.HandlerFunc {
	if FileRequestLogger == nil {
		InitFileRequestLogger()
	}

	return func(c *gin.Context) {
		hostName := strings.Split(c.Request.Host, ".")[0]
		hostInfo := strings.Split(hostName, "-")
		bindName := hostInfo[0]
		c.Set("bindName", bindName)
		if bindName == "0" {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped

		param := gin.LogFormatterParams{
			Request: c.Request,
			Keys:    c.Keys,
		}

		// Stop timer
		param.TimeStamp = time.Now()
		param.Latency = param.TimeStamp.Sub(start)

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

		param.BodySize = c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		param.Path = path

		fmt.Fprint(FileRequestLogger.Out, FileRequestLogFormatter(param))
	}
}

func DeleteTimeoutLog() {
	defer panichandler.CatchPanicStack()

	logger.DeleteLog("./log", 7*24*3600)
	logger.DeleteLog("./requestRecordlog", 7*24*3600)
}

func UploadFileRequestLog(fileName string) {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}

	file, err := os.Open(fileName)
	if err != nil {
		logger.Error("UploadFileRequestLog open log file error", "err", err, "fileName", fileName)
		return
	}
	stat, err := file.Stat()
	if err != nil {
		logger.Error("UploadFileRequestLog file.Stat() error", "err", err, "fileName", fileName)
		file.Close()
		return
	}
	if stat.Size() <= 0 {
		file.Close()
		return
	}
	file.Close()

	logFilePath := fileName
	url := domainmgr.UsingDomain + global.UploadFileRequestLog
	fmt.Println(url)
	_, err = req.Post(url, req.File(logFilePath), authHeader)
	if err != nil {
		logger.Error("upload fileRequestLog error", "err", err, "file", logFilePath)
	}
}
