package downloader

import (
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/downloadtaskmgr"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/domainmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"io/ioutil"
	"net/http"
	"os"
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
	if !utils.Exists(dir) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			logger.Error("AddToDownloadQueue os.MkdirAll error", "err", err, "dir", dir)
			return err
		}
	}
	fileName := utils.FileAddMark(downloadCmd.FileName, common.RedirectMark)
	savePath := dir + "/" + fileName

	info := &downloadtaskmgr.DownloadInfo{
		OriginTag:    downloadCmd.TransferTag,
		TargetUrl:    downloadCmd.DownloadUrl,
		BindName:     downloadCmd.BindName,
		FileName:     downloadCmd.FileName,
		Continent:    downloadCmd.RequestContinent,
		Country:      downloadCmd.RequestCountry,
		Area:         downloadCmd.RequestArea,
		SavePath:     savePath,
		DownloadType: downloadCmd.DownloadType,
		OriginRegion: downloadCmd.OriginRegion,
		TargetRegion: downloadCmd.TargetRegion,

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
		FileName:         task.FileName,
		BindName:         task.BindName,
		RequestContinent: task.Continent,
		RequestCountry:   task.Country,
		RequestArea:      task.Area,
		DownloadType:     task.DownloadType,
		OriginRegion:     task.OriginRegion,
		TargetRegion:     task.TargetRegion,
		DownloadUrl:      task.TargetUrl,
		FileSize:         uint64(fileInfo.Size()),
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
		FileName:         task.FileName,
		BindName:         task.BindName,
		RequestContinent: task.Continent,
		RequestCountry:   task.Country,
		RequestArea:      task.Area,
		DownloadType:     task.DownloadType,
		OriginRegion:     task.OriginRegion,
		TargetRegion:     task.TargetRegion,
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

func OnDownloadStart(task *downloadtaskmgr.DownloadTask) {
	defer panichandler.CatchPanicStack()
	//send down load start msg
	logger.Debug("Download Start", "task", task)
	payload := commonmsg.TerminalDownloadStartMsg{
		BindName:         task.BindName,
		FileName:         task.FileName,
		RequestContinent: task.Continent,
		RequestCountry:   task.Country,
		RequestArea:      task.Area,
	}
	logger.Debug("report start to server", "msg", payload)
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}
	_, err := httputils.Request("POST", domainmgr.UsingDomain+global.ReportDownloadStartUrl, payload, header)
	if err != nil {
		logger.Error("send downloadStart msg to server error", "err", err)
	}
}

func OnDownloading(task *downloadtaskmgr.DownloadTask, usedTimeSec int) {
	defer panichandler.CatchPanicStack()
	//send download process info
	logger.Debug("Download Process", "downloaded", task.DownloadedSize, "usedTimeSec", usedTimeSec)
	if usedTimeSec%60000 != 0 {
		return
	}

	payload := commonmsg.TerminalDownloadProcessMsg{
		BindName:         task.BindName,
		FileName:         task.FileName,
		RequestContinent: task.Continent,
		RequestCountry:   task.Country,
		RequestArea:      task.Area,
		Downloaded:       task.DownloadedSize,
	}
	logger.Debug("report process to server", "msg", payload)
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}
	_, err := httputils.Request("POST", domainmgr.UsingDomain+global.ReportDownloadProcessUrl, payload, header)
	if err != nil {
		logger.Error("send downloadprocess msg to server error", "err", err)
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
	downloadtaskmgr.SetOnDownloadStart(OnDownloadStart)
	downloadtaskmgr.SetOnDownloading(OnDownloading)

	//start loop
	downloadtaskmgr.Run()
}
