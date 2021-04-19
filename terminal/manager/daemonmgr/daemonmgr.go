package daemonmgr

//import (
//	"github.com/takama/daemon"
//	"net"
//	"os"
//	"os/signal"
//	"syscall"
//)
//
//const (
//
//	// name of the service
//	Name        = "meson"
//	Description = "meson terminal"
//)
//
//// Service has embedded daemon
//type Service struct {
//	daemon.Daemon
//}
//
//// Manage by daemon commands or run the daemon
//func (service *Service) Manage() (string, error) {
//
//	usage := "Usage: myservice install | remove | start | stop | status"
//
//	// if received any kind of command, do it
//	if len(os.Args) > 1 {
//		command := os.Args[1]
//		switch command {
//		case "install":
//			return service.Install()
//		case "remove":
//			return service.Remove()
//		case "start":
//			return service.Start()
//		case "stop":
//			return service.Stop()
//		case "status":
//			return service.Status()
//		default:
//			return usage, nil
//		}
//	}
//
//	// Do something, call your goroutines, etc
//
//	// Set up channel on which to send signal notifications.
//	// We must use a buffered channel or risk missing the signal
//	// if we're not ready to receive when the signal is sent.
//	interrupt := make(chan os.Signal, 1)
//	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
//
//	// Set up listener for defined host and port
//	listener, err := net.Listen("tcp", port)
//	if err != nil {
//		return "Possibly was a problem with the port binding", err
//	}
//
//	// set up channel on which to send accepted connections
//	listen := make(chan net.Conn, 100)
//	go acceptConnection(listener, listen)
//
//	// loop work cycle with accept connections or interrupt
//	// by system signal
//	for {
//		select {
//		case conn := <-listen:
//			go handleClient(conn)
//		case killSignal := <-interrupt:
//			stdlog.Println("Got signal:", killSignal)
//			stdlog.Println("Stoping listening on ", listener.Addr())
//			listener.Close()
//			if killSignal == os.Interrupt {
//				return "Daemon was interruped by system signal", nil
//			}
//			return "Daemon was killed", nil
//		}
//	}
//
//	// never happen, but need to complete code
//	return usage, nil
//}
