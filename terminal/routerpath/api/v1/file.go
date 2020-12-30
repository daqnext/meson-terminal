package v1

import (
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/downloadtaskmgr"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

func init() {
	common.AutoConfigRouter()

	// /api/v1/file/save
	common.GetMyRouter().POST("/save", saveNewFileHandler)
	// /api/v1/file/delete
	common.GetMyRouter().POST("/delete", deleteFileHandler)

	// /api/v1/file/deletefolder
	common.GetMyRouter().POST("/deletefolder", deleteFolderHandler)

	// /api/v1/file/pause
	common.GetMyRouter().GET("/pause/:time", pauseHandler)

}

func saveNewFileHandler(ctx *gin.Context) {
	//get cmd msg
	var downloadCmd commonmsg.DownLoadFileCmdMsg
	if err := ctx.ShouldBindJSON(&downloadCmd); err != nil {
		resp.ErrorResp(ctx, resp.ErrMalParams)
		return
	}

	//check disk space
	fileSize := downloadCmd.FileSize
	filemgr.GenDiskSpace(fileSize)

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

func deleteFolderHandler(ctx *gin.Context) {
	var msg commonmsg.DeleteFolderCmdMsg
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		resp.ErrorResp(ctx, resp.ErrMalParams)
		return
	}

	err := filemgr.DeleteFolder(msg.FolderName)
	if err != nil {
		resp.ErrorResp(ctx, resp.ErrInternalError)
		return
	}

	resp.SuccessResp(ctx, nil)
}

func pauseHandler(ctx *gin.Context) {
	pauseTimeStr := ctx.Param("time")
	pauseTime, err := strconv.Atoi(pauseTimeStr)
	if err != nil {
		pauseTime = 4
	}
	global.PauseTransfer = true
	resp.SuccessResp(ctx, nil)
	go func() {
		time.Sleep(time.Second * time.Duration(pauseTime))
		global.PauseTransfer = false
	}()
}
