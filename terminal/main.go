package main

import (
	"github.com/daqnext/meson-common/common/runpath"
	"github.com/daqnext/meson-terminal/terminal/manager/fixregionmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/terminallogger"
	"github.com/daqnext/meson-terminal/terminal/manager/tlscertificate"
	"github.com/takama/daemon"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

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

//var systemDConfig = `[Unit]
//Description={{.Description}}
//Requires={{.Dependencies}}
//After={{.Dependencies}}
//[Service]
//PIDFile=/var/run/{{.Name}}.pid
//ExecStartPre=/bin/rm -f /var/run/{{.Name}}.pid
//ExecStart=/bin/bash -c '{{.Path}} {{.Args}}'
//Restart=always
//[Install]
//WantedBy=multi-user.target
//`

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
	fmt.Println("Terminal starting...")
	//check
	fixregionmgr.CheckAvailable()

	//version check
	versionmgr.CheckVersion()

	config.CheckConfig()

	err := security.DownloadAndInitPublicKey()
	if err != nil {
		logger.Error("InitPublicKey error, try to download key by manual url=\"https://assets.meson.network:10443/static/terminal/publickey/meson_PublicKey.pem\"", "err", err)
		if MesonService != nil {
			MesonService.Stop()
		}
		logger.Fatal("Terminal Stopped")
	}

	defer panichandler.CatchPanicStack()

	//login
	account.TerminalLogin(fixregionmgr.Using+global.TerminalLoginUrl, config.UsingToken)
	err = filemgr.Init()
	if err != nil {
		if MesonService != nil {
			MesonService.Stop()
		}
		logger.Fatal("Terminal Stopped")
	}

	//waiting for confirm msg
	go func() {
		defer panichandler.CatchPanicStack()
		select {
		case flag := <-account.ServerRequestTest:
			if flag == true {
				logger.Info("net connect confirmed")
				logger.Info("Terminal start success")
				account.ServerRequestTest = nil
				global.TerminalIsRunning = true
			}
		case <-time.After(45 * time.Second):
			logger.Error("Net connect error. Please confirm that your machine can be accessed by the external network and the port is opened on the firewall.")
			if MesonService != nil {
				MesonService.Stop()
			}
			logger.Fatal("Terminal Stopped")
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

	//get host
	fixregionmgr.SyncTrackHost()

	//check TlsCertificate
	tlscertificate.CheckTlsCertificate()

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
		httpsAddr := fmt.Sprintf(":%s", config.UsingPort)
		httpsGinServer := ginrouter.GetGinInstance(routerpath.DefaultGin)
		// https server
		err = httpserver.StartHttpsServer(httpsAddr, tlscertificate.CrtFileName, tlscertificate.KeyFileName, httpsGinServer.GinInstance)
		if err != nil {
			logger.Error("https server start error", "err", err)
			return err
		}
		return nil
	})

	//don't hold main here
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

var MesonService *Service

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {

	usage := "command error. Usage: sudo ./meson service-install | service-remove | service-start | service-stop | service-status"
	// If received any kind of command, do it
	if len(os.Args) > 1 && !strings.Contains(os.Args[1], "-config=") {
		command := os.Args[1]
		switch command {
		case "service-install":
			//domain check
			fixregionmgr.CheckAvailable()
			config.CheckConfig()
			account.TerminalLogin(fixregionmgr.Using+global.TerminalLoginUrl, config.UsingToken)
			//errorLog := "2>>"+filepath.Join(runpath.RunPath, "./error.log")
			return service.Install()
		case "service-remove":
			newConfigs := map[string]string{
				config.Token:      "",
				config.Port:       "",
				config.SpaceLimit: "",
			}
			err := config.RecordConfigToFile(newConfigs)
			if err != nil {
				logger.Error("RecordConfigToFile error", "err", err)
			}
			service.Stop()
			return service.Remove()
		case "service-start":
			//domain check
			fixregionmgr.CheckAvailable()
			config.CheckConfig()
			account.TerminalLogin(fixregionmgr.Using+global.TerminalLoginUrl, config.UsingToken)
			return service.Start()
		case "service-stop":
			// No need to explicitly stop cron since job will be killed
			return service.Stop()
		case "service-status":
			fmt.Println("--- log ---")
			//latest 30 logs
			logArray, err := terminallogger.GetLatestLog(40)
			if err != nil {
				//do something
				fmt.Println("no any logs")
				fmt.Println("--- log end ---")
				logPath := filepath.Join(runpath.RunPath, "dailylog")
				fmt.Println("--- you can try to check logs in ", logPath)
			} else {
				for _, v := range logArray {
					fmt.Print(v)
				}
				fmt.Println("--- log end ---")
				logPath := filepath.Join(runpath.RunPath, "dailylog")
				fmt.Println("--- more logs in ", logPath)
			}
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
	MesonService = &Service{srv}
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		err := MesonService.SetTemplate(systemDConfig)
		if err != nil {
			logger.Error("MesonService SetTemplate error", "err", err)
		}
	}

	status, err := MesonService.Manage()
	if err != nil {
		fmt.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
