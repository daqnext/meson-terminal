package downloader

import (
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/downloadtaskmgr"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"os"
	"time"
)

//func DownloadFunc(task *downloadtaskmgr.DownloadTask) error {
//	dir := global.FileDirPath + "/" + task.BindNameHash
//	if !utils.Exists(dir) {
//		os.Mkdir(dir, 0777)
//	}
//	fileName := utils.FileAddMark(task.FileNameHash, common.RedirectMark)
//	filePath := dir + "/" + fileName
//
//	err := downloadtaskmgr.DownLoadFile(task.TargetUrl, filePath)
//	if err != nil {
//		logger.Error("download file url="+task.TargetUrl+"error", "err", err)
//		return err
//	}
//	go ldb.SetAccessTimeStamp(task.BindNameHash+"/"+task.FileNameHash, time.Now().Unix())
//	//get file size
//	fileInfo, err := os.Stat(filePath)
//	if err == nil {
//		filemgr.CdnSpaceUsed += fileInfo.Size()
//	}
//
//	//post download finish msg to server
//	payload := commonmsg.TerminalDownloadFinishMsg{
//		TransferTag:  task.OriginTag,
//		FileNameHash: task.FileName,
//		BindNameHash: task.BindName,
//		Continent:    task.Continent,
//		Country:      task.Country,
//		Area:         task.Area,
//	}
//	header := map[string]string{
//		"Content-Type":  "application/json",
//		"Authorization": "Bearer " + accountmgr.Token,
//	}
//	_, err = httputils.Request("POST", global.ReportDownloadFinishUrl, payload, header)
//	if err != nil {
//		logger.Error("send downloadfinish msg to server error", "err", err)
//	}
//
//	return nil
//}

func AddToDownloadQueue(downloadCmd commonmsg.DownLoadFileCmdMsg) error {
	dir := global.FileDirPath + "/" + downloadCmd.BindNameHash
	if !utils.Exists(dir) {
		os.Mkdir(dir, 0777)
	}
	savePath := dir + "/" + downloadCmd.FileNameHash

	info := &downloadtaskmgr.DownloadInfo{
		TargetUrl: downloadCmd.DownloadUrl,
		OriginTag: downloadCmd.TransferTag,
		BindName:  downloadCmd.BindNameHash,
		FileName:  downloadCmd.FileNameHash,
		Continent: downloadCmd.Continent,
		Country:   downloadCmd.Country,
		Area:      downloadCmd.Area,
		SavePath:  savePath,
	}

	return downloadtaskmgr.AddGlobalDownloadTask(info)
}

func OnDownloadSuccess(task *downloadtaskmgr.DownloadTask) {
	filePath := task.SavePath
	go ldb.SetAccessTimeStamp(task.BindName+"/"+task.FileName, time.Now().Unix())
	//get file size
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		filemgr.CdnSpaceUsed += fileInfo.Size()
	}

	//post download finish msg to server
	payload := commonmsg.TerminalDownloadFinishMsg{
		TransferTag:  task.OriginTag,
		FileNameHash: task.FileName,
		BindNameHash: task.BindName,
		Continent:    task.Continent,
		Country:      task.Country,
		Area:         task.Area,
	}
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}
	_, err = httputils.Request("POST", global.ReportDownloadFinishUrl, payload, header)
	if err != nil {
		logger.Error("send downloadfinish msg to server error", "err", err)
	}
}

func OnDownloadFailed(task *downloadtaskmgr.DownloadTask) {
	//post failed msg to server
	payload := commonmsg.TerminalDownloadFailedMsg{
		TransferTag:  task.OriginTag,
		FileNameHash: task.FileName,
		BindNameHash: task.BindName,
		Continent:    task.Continent,
		Country:      task.Country,
		Area:         task.Area,
	}
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}
	_, err := httputils.Request("POST", global.ReportDownloadFailedUrl, payload, header)
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
