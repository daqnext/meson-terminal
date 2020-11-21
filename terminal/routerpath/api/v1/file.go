package v1

import (
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/downloadtaskmgr"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/gin-gonic/gin"
)

func init() {
	common.AutoConfigRouter()

	// /api/v1/file/save
	common.GetMyRouter().POST("/save", saveNewFileHandler)
	// /api/v1/file/delete
	common.GetMyRouter().POST("/delete", deleteFileHandler)
}

func saveNewFileHandler(ctx *gin.Context) {
	//接收到下载文件的命令,生成下载任务,加入到下载任务队列中
	var downloadCmd commonmsg.DownLoadFileCmdMsg
	if err := ctx.ShouldBindJSON(&downloadCmd); err != nil {
		resp.ErrorResp(ctx, resp.ErrMalParams)
		return
	}

	//先检查文件是否存在
	filePath := global.FileDirPath + "/" + downloadCmd.BindNameHash + "/" + downloadCmd.FileNameHash
	if utils.Exists(filePath) {
		//如果文件存在,返回任务接收成功
		resp.SuccessResp(ctx, nil)

		//并且同时发送请求给server,告诉server下载完成
		payload := commonmsg.TerminalDownloadFinishMsg{
			TransferTag:  downloadCmd.TransferTag,
			FileNameHash: downloadCmd.FileNameHash,
			BindNameHash: downloadCmd.BindNameHash,
			Continent:    downloadCmd.Continent,
			Country:      downloadCmd.Country,
			Area:         downloadCmd.Area,
		}
		header := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + accountmgr.Token,
		}
		_, err := httputils.Request("POST", global.ReportDownloadFinishUrl, payload, header)
		if err != nil {
			logger.Error("send downloadfinish msg to server error", "err", err)
		}

		return
	}

	//文件不存在,就加入新的下载任务
	err := downloadtaskmgr.AddTask(
		downloadCmd.DownloadUrl,
		downloadCmd.TransferTag,
		downloadCmd.Continent,
		downloadCmd.Country,
		downloadCmd.Area,
		downloadCmd.BindNameHash,
		downloadCmd.FileNameHash,
		0,
	)
	if err != nil {
		resp.ErrorResp(ctx, resp.ErrAddDownloadTaskFailed)
		return
	}
	resp.SuccessResp(ctx, nil)
}

func deleteFileHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status": 0,
		"msg":    "deleteFileHandler",
	})
}
