package terminallogger

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/runpath"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/fixregionmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var FileRequestLogger *logrus.Logger
var FileRequestChan = make(chan string, 1024*100)

func init() {
	InitDefaultLogger()
}

func InitDefaultLogger() {
	log := logrus.New()

	recordPath := filepath.Join(runpath.RunPath, "./daily")
	//fmt.Println("Default log folder ", "path:", recordPath+"log")
	fileWriter := logger.LogFileWriter{
		RootDir: recordPath,
	}
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

	recordPath := filepath.Join(runpath.RunPath, "requestRecord")
	//fmt.Println("FileRequest log folder ", "path:", recordPath+"log")
	fileWriter := logger.LogFileWriter{
		RootDir:         recordPath,
		OnLogFileChange: UploadFileRequestLog,
		//MaxSize:         1024 * 3, //only for test
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

		//fmt.Fprint(FileRequestLogger.Out, FileRequestLogFormatter(param))
		FileRequestChan <- FileRequestLogFormatter(param)
	}
}

func RecordFileRequest() {
	go func() {
		defer panichandler.CatchPanicStack()
		for true {
			record := <-FileRequestChan
			fmt.Fprint(FileRequestLogger.Out, record)
		}
	}()
}

func DeleteTimeoutLog() {
	defer panichandler.CatchPanicStack()

	recordPath := filepath.Join(runpath.RunPath, "./dailylog")
	logger.DeleteLog(recordPath, 7*24*3600)
	recordPath = filepath.Join(runpath.RunPath, "./requestRecordlog")
	logger.DeleteLog(recordPath, 7*24*3600)
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
	defer file.Close()
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
	url := fixregionmgr.FixRegionD + global.UploadFileRequestLog
	//fmt.Println(url)
	_, err = req.Post(url, req.File(logFilePath), authHeader)
	if err != nil {
		logger.Error("upload fileRequestLog error", "err", err, "file", logFilePath)
	}
}

func GetLatestLog(lineCount int) ([]string, error) {
	recordPath := filepath.Join(runpath.RunPath, "./dailylog")
	rds, err := ioutil.ReadDir(recordPath)
	if err != nil {
		return nil, err
	}
	if len(rds) == 0 {
		return nil, errors.New("no log file")
	}
	//rds:=sortByTime(rd)
	logArray := []string{}

	for i := len(rds) - 1; i >= 0; i-- {
		logFileInfo := rds[i]
		logFile, err := os.Open(filepath.Join(recordPath, logFileInfo.Name()))
		if err != nil {
			continue
		}

		logArrayTemp := []string{}
		buf := bufio.NewReader(logFile)
		for {
			s, _, c := buf.ReadLine()
			if c == io.EOF {
				break
			}
			logArrayTemp = append(logArrayTemp, string(s)+"\n")
			if len(logArrayTemp) > lineCount {
				logArrayTemp = logArrayTemp[1:]
			}
		}
		logFile.Close()
		logArray = append(logArrayTemp, logArray...)
		if len(logArray) > lineCount {
			start := len(logArray) - lineCount
			logArray = logArray[start:]
			return logArray, nil
		}
	}
	return logArray, nil
}

func lineCounter(r io.Reader) (int, error) {

	var readSize int
	var err error
	var count int

	buf := make([]byte, 1024)

	for {
		readSize, err = r.Read(buf)
		if err != nil {
			break
		}

		var buffPosition int
		for {
			i := bytes.IndexByte(buf[buffPosition:], '\n')
			if i == -1 || readSize == buffPosition {
				break
			}
			buffPosition += i + 1
			count++
		}
	}
	if readSize > 0 && count == 0 || count > 0 {
		count++
	}
	if err == io.EOF {
		return count, nil
	}

	return count, err
}

func sortByTime(pl []os.FileInfo) []os.FileInfo {
	sort.Slice(pl, func(i, j int) bool {
		flag := true
		if pl[i].ModTime().After(pl[j].ModTime()) {
			flag = false
		} else if pl[i].ModTime().Equal(pl[j].ModTime()) {
			if pl[i].Name() < pl[j].Name() {
				flag = false
			}
		}
		return flag
	})
	return pl
}
