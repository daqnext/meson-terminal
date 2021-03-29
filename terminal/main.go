package main

import (
	"fmt"
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/job"
	"github.com/daqnext/meson-terminal/terminal/manager/account"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/downloader"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/httpserver"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"github.com/daqnext/meson-terminal/terminal/manager/security"
	"github.com/daqnext/meson-terminal/terminal/manager/statemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/terminallogger"
	"github.com/daqnext/meson-terminal/terminal/manager/versionmgr"
	"github.com/daqnext/meson-terminal/terminal/routerpath"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"

	//api router
	_ "github.com/daqnext/meson-terminal/terminal/routerpath/api"
)

func main() {
	terminallogger.InitLogger()

	//domain check

	//version check
	versionmgr.CheckVersion()

	//download publickey
	//url := "https://assets.meson.network:10443/static/terminal/publickey/meson_PublicKey.pem"
	//err := downloadtaskmgr.DownLoadFile(url, security.KeyPath)
	//if err != nil {
	//	logger.Error("download publicKey url="+url+"error", "err", err)
	//}

	config.CheckConfig()
	filemgr.Init()

	//publicKey
	err := security.InitPublicKey(security.KeyPath)
	if err != nil {
		logger.Fatal("InitPublicKey error, try to download key by manual", "err", err)
	}

	//login
	account.TerminalLogin(global.TerminalLoginUrl, config.UsingToken)

	defer panichandler.CatchPanicStack()

	//waiting for confirm msg
	go func() {
		select {
		case flag := <-account.ServerRequestTest:
			if flag == true {
				logger.Info("net connect confirmed")
				account.ServerRequestTest = nil
				global.TerminalIsRunning = true
			}
		case <-time.After(45 * time.Second):
			logger.Fatal("net connect error,please make sure your port is open")
		}
	}()

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
		job.StartScheduleJob()
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
	err = httpserver.StartHttpsServer(httpsAddr, crtFileName, keyFileName, httpsGinServer)
	if err != nil {
		logger.Error("https server start error", "err", err)
	}
}

func CheckGinStart(onStart func()) {
	go func() {
		defer panichandler.CatchPanicStack()
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
