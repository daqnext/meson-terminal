package downloader

import (
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/downloadtaskmgr"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/runpath"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/domainmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func DownloadFile(url string, savePath string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(savePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func AddToDownloadQueue(downloadCmd commonmsg.DownLoadFileCmdMsg) error {
	dir := global.FileDirPath + "/" + downloadCmd.BindName
	dir = filepath.Join(runpath.RunPath, dir)
	if !utils.Exists(dir) {
		os.MkdirAll(dir, 0777)
	}
	fileName := utils.FileAddMark(downloadCmd.FileName, common.RedirectMark)
	savePath := dir + "/" + fileName

	info := &downloadtaskmgr.DownloadInfo{
		TargetUrl: downloadCmd.DownloadUrl,
		BindName:  downloadCmd.BindName,
		FileName:  downloadCmd.FileName,
		Continent: downloadCmd.RequestContinent,
		Country:   downloadCmd.RequestCountry,
		Area:      downloadCmd.RequestArea,
		SavePath:  savePath,
		//CacheTime: downloadCmd.CacheTime,
	}

	return downloadtaskmgr.AddGlobalDownloadTask(info)
}

func OnDownloadSuccess(task *downloadtaskmgr.DownloadTask) {
	logger.Debug("download success", "task", task)

	filePath := task.SavePath
	fileName := utils.FileAddMark(task.FileName, common.RedirectMark)
	go func() {
		defer panichandler.CatchPanicStack()
		ldb.SetAccessTimeStamp(task.BindName+"/"+fileName, time.Now().Unix())
	}()
	//get file size
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		filemgr.CdnSpaceUsed += fileInfo.Size()
	}

	//post download finish msg to server
	payload := commonmsg.TerminalDownloadFinishMsg{
		FileNameHash:     task.FileName,
		BindNameHash:     task.BindName,
		RequestContinent: task.Continent,
		RequestCountry:   task.Country,
		RequestArea:      task.Area,
	}
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}
	_, err = httputils.Request("POST", domainmgr.UsingDomain+global.ReportDownloadFinishUrl, payload, header)
	if err != nil {
		logger.Error("send downloadfinish msg to server error", "err", err)
	}
}

func OnDownloadFailed(task *downloadtaskmgr.DownloadTask) {
	logger.Debug("download fail", "task", task)

	//post failed msg to server
	payload := commonmsg.TerminalDownloadFailedMsg{
		FileNameHash:     task.FileName,
		BindNameHash:     task.BindName,
		RequestContinent: task.Continent,
		RequestCountry:   task.Country,
		RequestArea:      task.Area,
		DownloadUrl:      task.TargetUrl,
		FileSize:         uint64(task.FileSize),
	}
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}
	_, err := httputils.Request("POST", domainmgr.UsingDomain+global.ReportDownloadFailedUrl, payload, header)
	if err != nil {
		logger.Error("send downloadfailed msg to server error", "err", err)
	}
}

func StartDownloadJob() {
	//create folder
	if !utils.Exists(global.FileDirPath) {
		os.Mkdir(global.FileDirPath, 0777)
	}

	downloadtaskmgr.InitTaskMgr("./task")
	downloadtaskmgr.SetPanicCatcher(panichandler.CatchPanicStack)
	downloadtaskmgr.SetOnTaskSuccess(OnDownloadSuccess)
	downloadtaskmgr.SetOnTaskFailed(OnDownloadFailed)

	//start loop
	downloadtaskmgr.Run()
}
