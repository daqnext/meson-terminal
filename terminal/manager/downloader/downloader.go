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
	"os"
	"time"
)

func DownloadFunc(task *downloadtaskmgr.DownloadTask) error {
	dir := global.FileDirPath + "/" + task.BindNameHash
	if !utils.Exists(dir) {
		os.Mkdir(dir, 0777)
	}
	filePath := dir + "/" + task.FileNameHash
	//下载文件
	err := downloadtaskmgr.DownLoadFile(task.TargetUrl, filePath)
	if err != nil {
		logger.Error("download file url="+task.TargetUrl+"error", "err", err)
		return err
	}
	ldb.SetAccessTimeStamp(task.BindNameHash+"/"+task.FileNameHash, time.Now().Unix())
	//获取文件的大小
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		filemgr.CdnSpaceUsed += uint64(fileInfo.Size())
	}

	//将下载完成的消息 发送给server,告诉server下载完成
	payload := commonmsg.TerminalDownloadFinishMsg{
		TransferTag:  task.OriginTag,
		FileNameHash: task.FileNameHash,
		BindNameHash: task.BindNameHash,
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

	return nil
}

func OnDownloadFailed(task *downloadtaskmgr.DownloadTask) {
	//通知服务器下载失败
	payload := commonmsg.TerminalDownloadFailedMsg{
		TransferTag:  task.OriginTag,
		FileNameHash: task.FileNameHash,
		BindNameHash: task.BindNameHash,
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
	//如果文件夹不存在 先创建文件夹
	if !utils.Exists(global.FileDirPath) {
		os.Mkdir(global.FileDirPath, 0666)
	}

	//初始化下载任务管理
	downloadtaskmgr.InitTaskMgr("./task")
	//设置如何处理下载任务
	downloadtaskmgr.SetExecTaskFunc(DownloadFunc)
	//下载失败处理
	downloadtaskmgr.SetOnTaskFailed(OnDownloadFailed)
	//开始任务循环
	downloadtaskmgr.Run()
}
