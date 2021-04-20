package daemonmgr

import (
	"fmt"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/takama/daemon"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

const (
	// name of the service
	Name        = "meson"
	Description = "meson terminal"
)

// Service has embedded daemon
type Service struct {
	daemon.Daemon
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage(apiServerFun func()) (string, error) {
	usage := "Usage: meson install | remove | start | stop | status"
	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":

			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	// Do something, call your goroutines, etc

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Waiting for interrupt by system signal
	killSignal := <-interrupt
	logger.Debug("Got signal:", "signal", killSignal)
	return "Service exited", nil
}

func AddDaemon(apiServerFun func()) {
	kind := daemon.SystemDaemon
	switch runtime.GOOS {
	case "darwin":
		kind = daemon.UserAgent
	}

	srv, err := daemon.New(Name, Description, kind)
	if err != nil {
		logger.Error("Daemon start error", "err", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage(apiServerFun)
	if err != nil {
		logger.Error(status, "Error", err)
		os.Exit(1)
	}
	fmt.Println(status)
}
