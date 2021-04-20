package main

import (
	"github.com/daqnext/meson-common/common/runpath"
	"github.com/takama/daemon"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	//terminalLogger
	_ "github.com/daqnext/meson-terminal/terminal/manager/terminallogger"
	"path/filepath"

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

var systemDConfig = `[Unit]
Description={{.Description}}
Requires={{.Dependencies}}
After={{.Dependencies}}
[Service]
PIDFile=/var/run/{{.Name}}.pid
ExecStartPre=/bin/rm -f /var/run/{{.Name}}.pid
ExecStart={{.Path}} {{.Args}}
Restart=always
[Install]
WantedBy=multi-user.target
`

func run() {
	//domain check
	domainmgr.CheckAvailableDomain()

	//version check
	versionmgr.CheckVersion()

	//download publickey
	publicKeyPath := filepath.Join(runpath.RunPath, security.KeyPath)
	url := "https://assets.meson.network:10443/static/terminal/publickey/meson_PublicKey.pem"
	err := downloader.DownloadFile(url, publicKeyPath)
	if err != nil {
		logger.Error("download publicKey url="+url+"error", "err", err)
	}

	config.CheckConfig()
	//publicKey
	err = security.InitPublicKey(publicKeyPath)
	if err != nil {
		logger.Fatal("InitPublicKey error, try to download key by manual", "err", err)
	}

	defer panichandler.CatchPanicStack()

	//login
	account.TerminalLogin(domainmgr.UsingDomain+global.TerminalLoginUrl, config.UsingToken)
	filemgr.Init()

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
			logger.Fatal("Net connect error. Please confirm that your machine can be accessed by the external network and the port is opened on the firewall.")
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
		crtFileName := filepath.Join(runpath.RunPath, "./host_chain.crt")
		keyFileName := filepath.Join(runpath.RunPath, "./host_key.key")
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

	//if err := g.Wait(); err != nil {
	//	logger.Fatal("gin server error", "err", err)
	//}
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

// Service is the daemon service struct
type Service struct {
	daemon.Daemon
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {

	usage := "Usage: meson install | remove | start | stop | status"
	// If received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			//domain check
			domainmgr.CheckAvailableDomain()
			config.CheckConfig()
			account.TerminalLogin(domainmgr.UsingDomain+global.TerminalLoginUrl, config.UsingToken)
			return service.Install()
		case "remove":
			newConfigs := map[string]string{
				config.Token:      "",
				config.Port:       "",
				config.SpaceLimit: "",
			}
			err := config.RecordConfigToFile(newConfigs)
			if err != nil {
				logger.Error("RecordConfigToFile error", "err", err)
			}
			return service.Remove()
		case "start":
			//domain check
			domainmgr.CheckAvailableDomain()
			config.CheckConfig()
			account.TerminalLogin(domainmgr.UsingDomain+global.TerminalLoginUrl, config.UsingToken)
			return service.Start()
		case "stop":
			// No need to explicitly stop cron since job will be killed
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	wg := sync.WaitGroup{}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGKILL, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Waiting for interrupt by system signal
		killSignal := <-interrupt
		logger.Debug("Got signal", "signal", killSignal)
		//return "Service exited", nil
	}()

	//run terminal
	run()

	wg.Wait()
	return "Service exited", nil
}

const (
	// name of the service
	name        = "meson"
	description = "meson terminal"
)

func main() {
	kind := daemon.SystemDaemon
	switch runtime.GOOS {
	case "darwin":
		kind = daemon.UserAgent
	}

	srv, err := daemon.New(name, description, kind)
	if err != nil {
		logger.Error("New daemon Error: ", "err", err)
		os.Exit(1)
	}
	service := &Service{srv}
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		service.SetTemplate(systemDConfig)
	}

	status, err := service.Manage()
	if err != nil {
		fmt.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
