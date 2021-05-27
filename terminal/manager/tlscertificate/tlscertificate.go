package tlscertificate

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/runpath"
	"github.com/daqnext/meson-terminal/terminal/manager/downloader"
	"github.com/daqnext/meson-terminal/terminal/manager/fixregionmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

var CrtFileName = filepath.Join(runpath.RunPath, "./host_chain.crt")
var KeyFileName = filepath.Join(runpath.RunPath, "./host_key.key")

func DownloadTlsFile() error {
	crtUrl := "https://assets.meson.network:10443/static/terminal/publickey/tls/host_chain.crt"
	err := downloader.DownloadFile(crtUrl, CrtFileName)
	if err != nil {
		logger.Error("download crtFile url="+crtUrl+" error", "err", err)
		return err
	}

	keyUrl := "https://assets.meson.network:10443/static/terminal/publickey/tls/host_key.key"
	err = downloader.DownloadFile(keyUrl, KeyFileName)
	if err != nil {
		logger.Error("download crtFile url="+keyUrl+" error", "err", err)
		return err
	}
	return nil
}

func CheckTlsCertificate() {
	defer panichandler.CatchPanicStack()

	//load tls file
	cert, err := tls.LoadX509KeyPair(CrtFileName, KeyFileName)
	if err != nil {
		logger.Error("CheckTlsCertificate tls.LoadX509KeyPair error", "err", err)
		//download new
		err := DownloadTlsFile()
		if err != nil {
			return
		}
		cert, err = tls.LoadX509KeyPair(CrtFileName, KeyFileName)
		if err != nil {
			return
		}
	}

	//parse
	c, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		logger.Error("CheckTlsCertificate x509.ParseCertificate error", "err", err)
		//download new
		err := DownloadTlsFile()
		if err != nil {
			return
		}
		cert, err = tls.LoadX509KeyPair(CrtFileName, KeyFileName)
		if err != nil {
			return
		}
		c, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return
		}
	}

	//check host correct
	err = c.VerifyHostname(fixregionmgr.TerminalTrack)
	if err != nil {
		logger.Error("CheckTlsCertificate c.VerifyHostname error", "err", err)
		//download new
		err := DownloadTlsFile()
		if err != nil {
			return
		}
		cert, err = tls.LoadX509KeyPair(CrtFileName, KeyFileName)
		if err != nil {
			return
		}
		c, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return
		}
		err = c.VerifyHostname(fixregionmgr.TerminalTrack)
		if err != nil {
			return
		}
	}

	//if pastDue time after 1 week
	if time.Now().Unix()+7*24*3600 < c.NotAfter.Unix() {
		return
	}

	err = DownloadTlsFile()
	if err != nil {
		return
	}
	cert, err = tls.LoadX509KeyPair(CrtFileName, KeyFileName)
	if err != nil {
		return
	}
	c, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return
	}
	err = c.VerifyHostname(fixregionmgr.TerminalTrack)
	if err != nil {
		return
	}

	if time.Now().Unix()+7*24*3600 > c.NotAfter.Unix() {
		if global.TerminalIsRunning {
			//restart
			switch runtime.GOOS {
			case "windows":
				logger.Error("TLS Certificate refreshed, please restart terminal")
			default:
				command := fmt.Sprintf("kill -1 %d", syscall.Getpid())
				cmd := exec.Command("/bin/bash", "-c", command)
				cmd.Run()
			}
		} else {
			logger.Error("TLS Certificate refreshed, please restart terminal")
		}
	}
}
