package routerpath

import (
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/downloadtaskmgr"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/account"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

func testHandler(ctx *gin.Context) {
	logger.Debug("Get test Request form Server")
	if account.ServerRequestTest != nil {
		account.ServerRequestTest <- true
	}
	ctx.JSON(200, gin.H{
		"status": 0,
		"time":   time.Now().Format("2006-01-02 15:04:05.000"),
	})
}

func healthHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status": 0,
	})
}

func requestCachedFilesHandler(ctx *gin.Context, bindName string, filePath string) {
	storagePath := global.FileDirPath + "/" + bindName + filePath
	exist := utils.Exists(storagePath)
	if exist {
		//fileName := path.Base(filePath)
		filePath = strings.Replace(filePath, "-redirecter456gt", "", 1)
		//set access time
		go ldb.SetAccessTimeStamp(bindName+filePath, time.Now().Unix())
		transferCacheFileFS(ctx, storagePath)
		return
	}

	//if not exist
	//redirect to server
	serverUrl := global.ServerDomain + "/api/cdn/" + bindName + filePath
	ctx.Redirect(302, serverUrl)
	return
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

func pauseHandler(ctx *gin.Context) {
	pauseTimeStr := ctx.Param("time")
	pauseTime, err := strconv.ParseInt(pauseTimeStr, 10, 64)
	if err != nil {
		pauseTime = 4
	}
	global.PauseMoment = time.Now().Unix() + pauseTime
	resp.SuccessResp(ctx, nil)
}
