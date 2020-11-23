package main

import (
	"fmt"
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/account"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/downloader"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/statemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/terminallogger"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"time"

	//api router
	_ "github.com/daqnext/meson-terminal/terminal/routerpath/api"
	_ "github.com/daqnext/meson-terminal/terminal/routerpath/api/v1"
)

func init() {
	terminallogger.InitLogger()
}

var healthPort = 0

func main() {
	config.CheckConfig()

	//login
	account.TerminalLogin(global.TerminalLoginUrl, config.UsingToken)

	go func() {
		select {
		case flag := <-account.ServerRequestTest:
			if flag == true {
				logger.Info("net connect confirmed")
				account.ServerRequestTest = nil
			}
		case <-time.After(30 * time.Second):
			logger.Fatal("net connect error,please make sure your port is open")
		}
	}()

	//设置gin的工作模式
	if config.GetString("ginMode") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	//sync cdn dir size
	filemgr.SyncCdnDirSize()

	//download queue
	downloader.StartDownloadJob()

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

	//go func() {
	//	time.Sleep(time.Second * 10)
	//	statemgr.SendStateToServer()
	//	//start schedule job- upload terminal state
	//	startScheduleJob()
	//}()

	CheckGinStart(func() {
		logger.Info("Terminal Is Running on Port:" + config.UsingPort)
		statemgr.SendStateToServer()
		//start schedule job- upload terminal state
		startScheduleJob()
	})

	addr := fmt.Sprintf(":%s", config.UsingPort)
	//logger.Info("Terminal Is Running on Port:" + config.UsingPort)
	if config.GetString("apiProto") == "http" {
		healthPort, _ = strconv.Atoi(config.UsingPort)
		common.GinRouter.Run(addr) // only in local dev
	} else {
		//
		go func() {
			port, _ := strconv.Atoi(config.UsingPort)
			for true {
				port++
				if port > 65535 {
					port = 19080
				}
				healthPort = port
				httpAddr := fmt.Sprintf(":%d", port)
				err := common.GinRouter.Run(httpAddr)
				if err != nil {
					continue
				}
			}
		}()
		err = common.GinRouter.RunTLS(addr, "./"+crtFileName, "./"+keyFileName)
		if err != nil {
			logger.Error("server start error", "err", err)
		}
	}
}

func CheckGinStart(onStart func()) {
	go func() {
		for true {
			time.Sleep(time.Second)
			header := map[string]string{
				"Content-Type": "application/json",
			}
			url := fmt.Sprintf("http://127.0.0.1:%d/api/testapi/health", healthPort)
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
