package routerpath

import (
	"encoding/json"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/account"
	"github.com/daqnext/meson-terminal/terminal/manager/downloader"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"github.com/daqnext/meson-terminal/terminal/manager/security"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"time"
)

func testHandler(ctx *gin.Context) {
	//logger.Debug("Get test Request form Server")
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

		// mapcount++
		// defer mapcount--

		go ldb.SetAccessTimeStamp(bindName+filePath, time.Now().Unix())
		transferCacheFileFS(ctx, storagePath)
		return
	}

	//if not exist
	//redirect to server
	serverUrl := global.ServerDomain + "/api/cdn/" + bindName + filePath
	ctx.Redirect(302, serverUrl)

	//notify server delete cache state
	go func() {
		defer panichandler.CatchPanicStack()

		header := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + accountmgr.Token,
		}
		payload := commonmsg.TerminalRequestDeleteFilesMsg{
			Files: []string{bindName + filePath},
		}
		respCtx, err := httputils.Request("POST", global.RequestToDeleteFilsUrl, payload, header)
		if err != nil {
			logger.Error("Request DeleteFiles error", "err", err)
			return
		}
		var respBody2 resp.RespBody
		if err := json.Unmarshal(respCtx, &respBody2); err != nil {
			logger.Error("response from terminal unmarshal error", "err", err)
			return
		}
		switch respBody2.Status {
		case 0:
			//get right request
			logger.Debug("notify server delete missing file success", "file")
		default:
			logger.Error("notify server delete missing file fail")
		}
	}()
	return
}

func saveNewFileHandler(ctx *gin.Context) {
	//get cmd msg
	var downloadCmd commonmsg.DownLoadFileCmdMsg
	if err := ctx.ShouldBindJSON(&downloadCmd); err != nil {
		resp.ErrorResp(ctx, resp.ErrMalParams)
		return
	}

	//check sign
	timeStamp := downloadCmd.TimeStamp
	//make sure request is in 30s
	if time.Now().Unix() > timeStamp+30 {
		logger.Error("save file request past due")
		resp.ErrorResp(ctx, resp.ErrInternalError)
		return
	}

	timeStampStr := strconv.FormatInt(timeStamp, 10)
	pass := security.ValidateSignature(timeStampStr, downloadCmd.Sign)
	if pass == false {
		logger.Error("ValidateSignature fail")
		resp.ErrorResp(ctx, resp.ErrInternalError)
		return
	}

	//check disk space
	fileSize := downloadCmd.FileSize
	filemgr.GenDiskSpace(fileSize)

	//就加入新的下载任务
	err := downloader.AddToDownloadQueue(downloadCmd)
	if err != nil {
		resp.ErrorResp(ctx, resp.ErrAddDownloadTaskFailed)
		return
	}
	resp.SuccessResp(ctx, nil)
}

func pauseHandler(ctx *gin.Context) {
	var msg commonmsg.TransferPauseMsg
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		resp.ErrorResp(ctx, resp.ErrMalParams)
		return
	}

	//check sign
	timeStamp := msg.TimeStamp
	//make sure request is in 30s
	if time.Now().Unix() > timeStamp+30 {
		logger.Error("save file request past due")
		resp.ErrorResp(ctx, resp.ErrInternalError)
		return
	}

	timeStampStr := strconv.FormatInt(timeStamp, 10)
	pass := security.ValidateSignature(timeStampStr, msg.Sign)
	if pass == false {
		logger.Error("ValidateSignature fail")
		resp.ErrorResp(ctx, resp.ErrInternalError)
		return
	}

	pauseTime := 4
	if msg.PauseTime > 0 && msg.PauseTime < 10 {
		pauseTime = msg.PauseTime
	}

	global.PauseMoment = time.Now().Unix() + int64(pauseTime)
	resp.SuccessResp(ctx, nil)
}
