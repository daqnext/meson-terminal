package main

import (
	"fmt"
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/downloader"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/httpserver"
	"github.com/daqnext/meson-terminal/terminal/manager/statemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/terminallogger"
	"github.com/daqnext/meson-terminal/terminal/manager/versionmgr"
	"github.com/daqnext/meson-terminal/terminal/routerpath"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"math/rand"
	"strconv"
	"time"

	//api router
	_ "github.com/daqnext/meson-terminal/terminal/routerpath/api"
	_ "github.com/daqnext/meson-terminal/terminal/routerpath/api/v1"
)

func main() {
	terminallogger.InitLogger()

	//version check
	versionmgr.CheckVersion()

	config.CheckConfig()
	filemgr.Init()

	//login
	//account.TerminalLogin(global.TerminalLoginUrl, config.UsingToken)

	//waiting for confirm msg
	//go func() {
	//	select {
	//	case flag := <-account.ServerRequestTest:
	//		if flag == true {
	//			logger.Info("net connect confirmed")
	//			account.ServerRequestTest = nil
	//			global.TerminalIsRunning = true
	//		}
	//	case <-time.After(45 * time.Second):
	//		logger.Fatal("net connect error,please make sure your port is open")
	//	}
	//}()

	//set gin mode
	if config.GetString("ginMode") == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	//sync cdn dir size
	filemgr.SyncCdnDirSize()
	//download queue
	downloader.StartDownloadJob()

	CheckGinStart(func() {
		logger.Info("Terminal Is Running on Port:" + config.UsingPort)
		statemgr.SendStateToServer()
		//start schedule job- upload terminal state
		startScheduleJob()
	})

	//start http api for server
	go func() {
		port, _ := strconv.Atoi(config.UsingPort)
		for true {
			port++
			if port > 65535 {
				port = 19080
			}
			global.ApiPort = strconv.Itoa(port)
			httpAddr := fmt.Sprintf(":%d", port)
			httpGinServer := common.GinRouter
			err := httpserver.StartHttpServer(httpAddr, httpGinServer)
			if err != nil {
				continue
			}
		}
	}()

	//start https api server
	//looking for ssl files
	crtFileName := "./host_chain.crt"
	keyFileName := "./host_key.key"
	httpsAddr := fmt.Sprintf(":%s", config.UsingPort)
	httpsGinServer := routerpath.RequestServer()
	// https server
	err := httpserver.StartHttpsServer(httpsAddr, crtFileName, keyFileName, httpsGinServer)
	if err != nil {
		logger.Error("https server start error", "err", err)
	}
}

func CheckGinStart(onStart func()) {
	go func() {
		for true {
			time.Sleep(time.Second)
			header := map[string]string{
				"Content-Type": "application/json",
			}
			url := fmt.Sprintf("http://127.0.0.1:%s/api/testapi/health", global.ApiPort)
			_, err := httputils.Request("GET", url, nil, header)
			if err != nil {
				logger.Debug("health check error", "err", err)
				continue
			}
			if onStart != nil {
				onStart()
			}
			break
		}
	}()
}

func startScheduleJob() {
	c := cron.New(cron.WithSeconds())
	rand.Seed(time.Now().Unix())

	//heartbeat
	randSecond := rand.Intn(30)
	schedule := fmt.Sprintf("%d,%d * * * * *", randSecond, randSecond+30)
	jobId, err := c.AddFunc(schedule, statemgr.SendStateToServer)
	if err != nil {
		logger.Error("ScheduleJob-"+"SendStateToServer"+" start error", "err", err)
	} else {
		logger.Info("ScheduleJob-"+"SendStateToServer"+" start", "ID", jobId, "Schedule", schedule)
	}

	//version check
	randSecond = rand.Intn(60)
	schedule = fmt.Sprintf("%d %d * * * *", randSecond, randSecond)
	jobId, err = c.AddFunc(schedule, versionmgr.CheckVersion)
	if err != nil {
		logger.Error("ScheduleJob-"+"VersionCheck"+" start error", "err", err)
	} else {
		logger.Info("ScheduleJob-"+"VersionCheck"+" start", "ID", jobId, "Schedule", schedule)
	}

	//sync folder size
	jobId, err = c.AddFunc("0 0 * * * *", filemgr.SyncCdnDirSize)
	if err != nil {
		logger.Error("ScheduleJob-"+"SyncCdnDirSize"+" start error", "err", err)
	} else {
		logger.Info("ScheduleJob-"+"SyncCdnDirSize"+" start", "ID", jobId, "Schedule", "0 0 * * * *")
	}

	//scan expiration files  every 6 hours
	schedule = fmt.Sprintf("%d 0 0,6,12,18 * * *", rand.Intn(60))
	//schedule = fmt.Sprintf("%d * * * * *", rand.Intn(60))
	jobId, err = c.AddFunc(schedule, filemgr.ScanExpirationFiles)
	if err != nil {
		logger.Error("ScheduleJob-"+"ScanExpirationFiles"+" start error", "err", err)
	} else {
		logger.Info("ScheduleJob-"+"ScanExpirationFiles"+" start", "ID", jobId, "Schedule", schedule)
	}

	//delete empty folder 1time/hour
	schedule = fmt.Sprintf("%d 0 * * * *", rand.Intn(60))
	jobId, err = c.AddFunc(schedule, filemgr.DeleteEmptyFolder)
	if err != nil {
		logger.Error("ScheduleJob-"+"DeleteEmptyFolder"+" start error", "err", err)
	} else {
		logger.Info("ScheduleJob-"+"DeleteEmptyFolder"+" start", "ID", jobId, "Schedule", schedule)
	}

	c.Start()
}
