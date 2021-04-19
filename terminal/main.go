package main

import (
	//terminalLogger
	_ "github.com/daqnext/meson-terminal/terminal/manager/terminallogger"
	//Init gin
	"github.com/daqnext/meson-terminal/terminal/routerpath"

	"fmt"
	"github.com/daqnext/meson-common/common/ginrouter"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/job"
	"github.com/daqnext/meson-terminal/terminal/manager/account"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/domainmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/downloader"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/httpserver"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"github.com/daqnext/meson-terminal/terminal/manager/security"
	"github.com/daqnext/meson-terminal/terminal/manager/statemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/versionmgr"
	"golang.org/x/sync/errgroup"
	"strconv"
	"time"

	//api router
	_ "github.com/daqnext/meson-terminal/terminal/routerpath/api"
)

var g errgroup.Group

func main() {
	//domain check
	domainmgr.CheckAvailableDomain()

	//version check
	versionmgr.CheckVersion()

	//download publickey
	url := "https://assets.meson.network:10443/static/terminal/publickey/meson_PublicKey.pem"
	err := downloader.DownloadFile(url, security.KeyPath)
	if err != nil {
		logger.Error("download publicKey url="+url+"error", "err", err)
	}

	config.CheckConfig()
	filemgr.Init()

	//publicKey
	err = security.InitPublicKey(security.KeyPath)
	if err != nil {
		logger.Fatal("InitPublicKey error, try to download key by manual", "err", err)
	}

	//login
	account.TerminalLogin(domainmgr.UsingDomain+global.TerminalLoginUrl, config.UsingToken)

	defer panichandler.CatchPanicStack()

	//waiting for confirm msg
	go func() {
		defer panichandler.CatchPanicStack()
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

	//sync cdn dir size
	filemgr.SyncCdnDirSize()
	//download queue
	downloader.StartDownloadJob()

	CheckGinStart(func() {
		logger.Info("Terminal Is Running on Port:" + config.UsingPort)
		statemgr.SendStateToServer()
		//start schedule job- upload terminal state
		job.StartLoopJob()
		job.StartScheduleJob()
	})

	//start api server
	g.Go(func() error {
		port, _ := strconv.Atoi(config.UsingPort)
		for true {
			port++
			if port > 65535 {
				port = 19080
			}
			global.HealthCheckPort = strconv.Itoa(port)
			httpAddr := fmt.Sprintf(":%d", port)
			testGinServer := ginrouter.GetGinInstance(routerpath.CheckStartGin)
			err := httpserver.StartHttpServer(httpAddr, testGinServer.GinInstance)
			if err != nil {
				continue
			}
		}
		return nil
	})

	//start cdn server
	g.Go(func() error {
		//start https api server
		crtFileName := "./host_chain.crt"
		keyFileName := "./host_key.key"
		httpsAddr := fmt.Sprintf(":%s", config.UsingPort)
		httpsGinServer := ginrouter.GetGinInstance(routerpath.DefaultGin)
		// https server
		err = httpserver.StartHttpsServer(httpsAddr, crtFileName, keyFileName, httpsGinServer.GinInstance)
		if err != nil {
			logger.Error("https server start error", "err", err)
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Fatal("gin server error", "err", err)
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
			url := fmt.Sprintf("http://127.0.0.1:%s/api/testapi/health", global.HealthCheckPort)
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
