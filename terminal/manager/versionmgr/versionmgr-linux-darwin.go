// +build linux darwin

package versionmgr

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/runpath"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

func CheckVersion() {
	defer panichandler.CatchPanicStack()
	//check is there new version or not
	latestVersion, _, err := GetTerminalVersionFromServer()
	if err != nil {
		logger.Info("Version check error, please check version on meson.network")
		return
	}
	vResult := utils.VersionCompare(Version, latestVersion)
	if vResult != -1 {
		logger.Info("Already Latest Version")
		return
	}

	//need upgrade
	logger.Info("New version detected, start to upgrade... ")
	//check arch and os
	arch, osInfo := GetOSInfo()

	// 'https://meson.network/static/terminal/v0.1.2/meson-darwin-amd64.tar.gz'
	fileName := "meson" + "-" + osInfo + "-" + arch + ".tar.gz"
	newVersionDownloadUrl := "https://assets.meson.network:10443/static/terminal/v" + latestVersion + "/" + fileName
	logger.Debug("new version download url", "url", newVersionDownloadUrl)
	//download new version
	fileName = path.Join(runpath.RunPath, fileName)
	err = DownloadNewVersion(fileName, newVersionDownloadUrl, latestVersion)
	if err != nil {
		logger.Error("auto upgrade error", "err", err)
		logger.Info("auto download new version error. Please download new version by manual.")
		return
	}

	//restart
	RestartTerminal()
}

func DownloadNewVersion(fileName string, downloadUrl string, newVersion string) error {
	//get
	response, err := http.Get(downloadUrl)
	if err != nil {
		logger.Error("get file url "+downloadUrl+" error", "err", err)
		return err
	}
	//creat folder and file
	distDir := path.Dir(fileName)
	err = os.MkdirAll(distDir, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	//defer file.Close()
	if response.Body == nil {
		file.Close()
		return errors.New("body is null")
	}
	defer response.Body.Close()
	_, err = io.Copy(file, response.Body)
	if err != nil {
		os.Remove(fileName)
		file.Close()
		return err
	}
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		os.Remove(fileName)
		file.Close()
		return err
	}
	size := fileInfo.Size()
	logger.Debug("donwload file,fileInfo", "size", size)

	if size == 0 {
		os.Remove(fileName)
		file.Close()
		return errors.New("download file size error")
	}
	file.Close()

	//unzip tar.gz
	targetDir := strings.Replace(fileName, ".tar.gz", "", 1)
	// file read
	fr, err := os.Open(fileName)
	if err != nil {
		logger.Error("open version file error", "err", err)
		return err
	}
	defer fr.Close()
	// gzip read
	gr, err := gzip.NewReader(fr)
	if err != nil {
		logger.Error("gzip read new version file error", "err", err)
		return err
	}
	defer gr.Close()
	// tar read
	tr := tar.NewReader(gr)
	// 读取文件
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("unzip new version file error", "err", err)
			return err
		}
		fileName := runpath.RunPath + "/" + h.Name
		dirName := string([]rune(fileName)[0:strings.LastIndex(fileName, "/")])
		err = os.MkdirAll(dirName, 0777)
		if err != nil {
			logger.Error("unzip new version file error-create dir", "err", err)
			return err
		}
		if utils.IsDir(fileName) {
			continue
		}
		fw, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0777 /*os.FileMode(h.Mode)*/)
		//fw,err:=os.Create(fileName)
		if err != nil {
			logger.Error("unzip new version file error-create file", "err", err)
			return err
		}
		defer fw.Close()
		// 写文件
		_, err = io.Copy(fw, tr)
		if err != nil {
			logger.Error("unzip new version file error-copy file", "err", err)
			return err
		}
	}
	logger.Debug("un tar.gz ok")

	//cover old version file
	files, _ := ioutil.ReadDir(targetDir)
	for _, v := range files {
		fileName := v.Name()
		fmt.Println(v.Name())
		if fileName == "config.txt" {
			continue
		}
		oldFile := path.Join(runpath.RunPath, fileName)
		newFile := path.Join(targetDir, fileName)
		err := coverOldFile(newFile, oldFile)
		if err != nil {
			logger.Error("new version file error-cover file", "err", err)
			continue
		}
	}

	versionFile := path.Join(runpath.RunPath, "./v"+Version)
	os.Remove(versionFile)

	os.RemoveAll(targetDir)
	os.Remove(fileName)

	return nil
}

func coverOldFile(newFile string, oldFile string) error {
	input, err := ioutil.ReadFile(newFile)
	if err != nil {
		return err
	}
	os.Remove(oldFile)
	err = ioutil.WriteFile(oldFile, input, 777)
	if err != nil {
		fmt.Println("Error creating", oldFile)
		fmt.Println(err)
		return err
	}
	os.Chmod(oldFile, 0777)
	return nil
}

func RestartTerminal() {
	if global.TerminalIsRunning {
		command := fmt.Sprintf("kill -1 %d", syscall.Getpid())
		cmd := exec.Command("/bin/bash", "-c", command)
		cmd.Run()
	} else {
		logger.Error("New version download finish.Please restart")
	}

}
