package main

import (
	"fmt"
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/downloader"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/statemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/terminallogger"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

func init() {
	config.ReadConfig()
	terminallogger.InitLogger()
}

func main() {
	config.CheckConfig()

	logger.Debug("test run")

	////login to get token
	//username := config.GetString("username")
	//password := config.GetString("password")
	//accountmgr.SLogin(global.SLoginUrl, username, password)
	//
	//设置gin的工作模式
	if config.GetString("ginMode") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	//sync cdn dir size
	filemgr.SyncCdnDirSize()

	//download queue
	downloader.StartDownloadJob()

	//start schedule job- upload terminal state
	startScheduleJob()

	//开启api服务器
	//查找目录下的证书文件
	crtFileName := ""
	keyFileName := ""
	rd, err := ioutil.ReadDir("./")
	if err != nil {
		logger.Error("ReadDir error", "err", err)
	}
	for _, fi := range rd {
		if !fi.IsDir() {
			filename := fi.Name()
			if strings.Contains(filename, ".pem") || strings.Contains(filename, ".crt") {
				crtFileName = filename
			}
			if strings.Contains(filename, ".key") {
				keyFileName = filename
			}
		}
	}

	logger.Info("Terminal Is Running...")

	addr := fmt.Sprintf(":%s", config.UsingPort)
	if config.GetString("apiProto") == "http" {
		common.GinRouter.Run(addr) // only in local dev
	} else {
		//
		err = common.GinRouter.RunTLS(addr, "./"+crtFileName, "./"+keyFileName)
		if err != nil {
			logger.Error("server start error", "err", err)
		}
	}
}

func startScheduleJob() {
	c := cron.New(cron.WithSeconds())
	rand.Seed(time.Now().Unix())

	//发送心跳包
	randSecond := rand.Intn(30)
	schedule := fmt.Sprintf("%d,%d * * * * *", randSecond, randSecond+30)
	jobId, err := c.AddFunc(schedule, statemgr.SendStateToServer)
	if err != nil {
		logger.Error("ScheduleJob-"+"SendStateToServer"+" start error", "err", err)
	} else {
		logger.Info("ScheduleJob-"+"SendStateToServer"+" start", "ID", jobId, "Schedule", schedule)
	}

	//同步file文件夹大小
	jobId, err = c.AddFunc("0 0 * * * *", filemgr.SyncCdnDirSize)
	if err != nil {
		logger.Error("ScheduleJob-"+"SyncCdnDirSize"+" start error", "err", err)
	} else {
		logger.Info("ScheduleJob-"+"SyncCdnDirSize"+" start", "ID", jobId, "Schedule", "0 0 * * * *")
	}

	//扫描过期文件 6小时一次
	schedule = fmt.Sprintf("%d 0 0,6,12,18 * * *", rand.Intn(60))
	jobId, err = c.AddFunc(schedule, filemgr.ScanExpirationFiles)
	if err != nil {
		logger.Error("ScheduleJob-"+"ScanExpirationFiles"+" start error", "err", err)
	} else {
		logger.Info("ScheduleJob-"+"ScanExpirationFiles"+" start", "ID", jobId, "Schedule", schedule)
	}

	c.Start()
}
