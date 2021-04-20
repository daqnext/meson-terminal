package routerpath

import (
	"encoding/json"
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-common/common/runpath"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/account"
	"github.com/daqnext/meson-terminal/terminal/manager/domainmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/downloader"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"github.com/daqnext/meson-terminal/terminal/manager/security"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"path/filepath"
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
		filePath = strings.Replace(filePath, common.RedirectMark, "", 1)
		//set access time

		// mapcount++
		// defer mapcount--

		go func() {
			defer panichandler.CatchPanicStack()
			ldb.SetAccessTimeStamp(bindName+filePath, time.Now().Unix())
		}()
		transferCacheFileFS(ctx, storagePath)
		return
	}

	//if not exist
	//redirect to server
	//todo: cdnDomain
	serverUrl := domainmgr.UsingDomain + "/api/cdn/" + bindName + filePath
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
		respCtx, err := httputils.Request("POST", domainmgr.UsingDomain+global.RequestToDeleteFilesUrl, payload, header)
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
			logger.Debug("notify server delete missing file success", "file", bindName+filePath)
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
	pass := security.CheckRequestLegal(downloadCmd.TimeStamp, downloadCmd.MachineMac, downloadCmd.Sign)
	if pass == false {
		resp.ErrorResp(ctx, resp.ErrInternalError)
		return
	}

	//check disk space
	fileSize := downloadCmd.FileSize
	//todo: handle if can not get file size
	if fileSize == 0 {
		fileSize = 1 * filemgr.UnitG
	}
	filemgr.GenDiskSpace(int64(fileSize))

	//就加入新的下载任务
	err := downloader.AddToDownloadQueue(downloadCmd)
	if err != nil {
		resp.ErrorResp(ctx, resp.ErrAddDownloadTaskFailed)
		return
	}
	resp.SuccessResp(ctx, nil)
}

func deleteFileHandler(ctx *gin.Context) {
	//get cmd msg
	var msg commonmsg.DownLoadFileCmdMsg
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		resp.ErrorResp(ctx, resp.ErrMalParams)
		return
	}

	//check sign
	pass := security.CheckRequestLegal(msg.TimeStamp, msg.MachineMac, msg.Sign)
	if pass == false {
		resp.ErrorResp(ctx, resp.ErrInternalError)
		return
	}

	err := filemgr.DeleteFile(msg.BindName, msg.FileName)
	if err != nil {
		logger.Error("Delete file fail")
		resp.ErrorResp(ctx, resp.ErrInternalError)
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
	pass := security.CheckRequestLegal(msg.TimeStamp, msg.MachineMac, msg.Sign)
	if pass == false {
		resp.ErrorResp(ctx, resp.ErrInternalError)
		return
	}

	pauseTime := 4
	if msg.PauseTime > 0 && msg.PauseTime < 10 {
		pauseTime = msg.PauseTime + 1
	}

	global.PauseMoment = time.Now().Unix() + int64(pauseTime)
	resp.SuccessResp(ctx, nil)
}

func fileRequestLogHandler(ctx *gin.Context) {
	logFiles := []byte{}
	path := filepath.Join(runpath.RunPath, "./requestRecordlog")
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Error("read ./requestRecordlog fail", "err", err, "dir", "./requestRecordlog/")
		resp.ErrorResp(ctx, resp.ErrInternalError)
		return
	}
	for _, fi := range rd {
		if !fi.IsDir() {
			name := "<a href=" + "/api/log/requestRecordlog/" + fi.Name() + ">" + "requestRecordlog/" + fi.Name() + "</a><br/>"
			logFiles = append(logFiles, []byte(name)...)
		}
	}
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", logFiles)
}

func fileDefaultLogHandler(ctx *gin.Context) {
	logFiles := []byte{}
	path := filepath.Join(runpath.RunPath, "./dailylog")
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Error("read ./log fail", "err", err, "dir", "./log/")
		resp.ErrorResp(ctx, resp.ErrInternalError)
		return
	}
	for _, fi := range rd {
		if !fi.IsDir() {
			name := "<a href=" + "/api/log/dailylog/" + fi.Name() + ">" + "log/" + fi.Name() + "</a><br/>"
			logFiles = append(logFiles, []byte(name)...)
		}
	}
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", logFiles)
}
