package filemgr

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/downloadtaskmgr"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/gvar"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var CdnSpaceLimit = uint64(config.GetInt("spacelimit") * 1000000000)
var CdnSpaceUsed uint64 = 0

func init() {
	//is dir exist
	if !utils.Exists(global.FileDirPath) {
		err := os.Mkdir(global.FileDirPath, 0700)
		if err != nil {
			logger.Fatal("tempfile dir create failed, please create dir " + global.FileDirPath + " by manual")
		}
	}

	if !utils.Exists(global.FileDirPath + "/standardfile") {
		err := os.Mkdir(global.FileDirPath+"/standardfile", 0700)
		if err != nil {
			logger.Fatal("tempfile dir create failed, please create dir " + global.FileDirPath + "/standardfile" + " by manual")
		}
	}

	//create std file
	for i := 5; i <= 30; i = i + 5 {
		fileName := global.FileDirPath + "/" + fmt.Sprintf("standardfile/%d.bin", i)
		if utils.Exists(fileName) {
			continue
		}
		f, err := os.Create(fileName)
		if err != nil {
			logger.Error("Create standardFile error", "err", err, "fileName", fileName)
			continue
		}
		if err := f.Truncate(int64(i * 1000 * 1000)); err != nil {
			logger.Error("Full standardFile error", "err", err, "fileName", fileName)
		}
		f.Close()
	}
}

func SyncCdnDirSize() {
	size, err := utils.GetDirSize(global.FileDirPath)
	if err != nil {
		logger.Error("get dir size error", "err", err)
	}
	CdnSpaceUsed = size
}

func ScanExpirationFiles() {
	//request expiration time from server
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}

	respCtx, err := httputils.Request("GET", global.RequestFileExpirationTimeUrl, nil, header)
	if err != nil {
		logger.Error("Request FileExpirationTime error", "err", err)
		return
	}
	var respBody resp.RespBody
	if err := json.Unmarshal(respCtx, &respBody); err != nil {
		logger.Error("response from terminal unmarshal error", "err", err)
		return
	}
	var fileExpirationTime = int64(0)
	switch respBody.Status {
	case 0:
		fileExpirationTime = int64(respBody.Data.(float64))
		logger.Debug("get FileExpirationTime", "FileExpirationTime", fileExpirationTime)
	default:
		logger.Error("Request FileExpirationTime response ")
		return
	}

	//scan expiration time
	expirationFils := []string{}
	iter := ldb.DB.NewIterator(nil, nil)
	nowTime := time.Now().Unix()
	for iter.Next() {
		key := iter.Key()
		file := string(key)
		if strings.Contains(file, "standardfile") {
			continue
		}
		value := iter.Value()
		lastAccessTimeStamp := int64(binary.LittleEndian.Uint64(value))
		if nowTime-lastAccessTimeStamp > fileExpirationTime {
			expirationFils = append(expirationFils, file)
		}
	}

	if len(expirationFils) == 0 {
		return
	}

	//post delete file list to server
	header = map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}
	payload := commonmsg.TerminalRequestDeleteFilesMsg{
		Files: expirationFils,
	}
	respCtx, err = httputils.Request("POST", global.RequestToDeleteFilsUrl, payload, header)
	if err != nil {
		logger.Error("Request DeleteFils error", "err", err)
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
		logger.Debug("agree to delete files")
		for _, v := range expirationFils {
			//file := strings.Split(v, "/")
			os.Remove(global.FileDirPath + "/" + v)
			ldb.DB.Delete([]byte(v), nil)
		}
		//if dir is empty,delete dir
		dirs, _ := ioutil.ReadDir(global.FileDirPath)
		for _, v := range dirs {
			dirName := v.Name()
			if v.IsDir() {
				//is dir empty
				files, _ := ioutil.ReadDir(global.FileDirPath + "/" + dirName)
				if len(files) == 0 {
					os.Remove(global.FileDirPath + "/" + dirName)
				}
			}
		}
		SyncCdnDirSize()
	default:
		logger.Error("Request FileExpirationTime response ")
		return
	}

}

func AccessTime() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestPath := ctx.Request.URL.String()
		filePath := strings.Replace(requestPath, "/api/static/files/", "", 1)
		//set access time
		go ldb.SetAccessTimeStamp(filePath, time.Now().Unix())
	}
}

func PreHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestPath := ctx.Request.URL.String()
		filePath := strings.Replace(requestPath, "/api/static/files/", "", 1)
		exist := utils.Exists(gvar.RootPath + "/files/" + filePath)
		if exist {
			//set access time
			go ldb.SetAccessTimeStamp(filePath, time.Now().Unix())
			return
		}

		//not exist
		extension := filepath.Ext(filePath)
		fileHash := utils.GetStringHash(filePath)

		// get referer from request header
		referer := ctx.Request.Referer()
		if referer == "" {
			ctx.Abort()
			return
		}
		//https://f39e5277e178ccb4cecfdf40b9ecaf95.shoppynext.com:19091/api/static/files/3776229933cd27f6c629269afeaf5f2f/1f71cc1221ac92cd4d07e01a39abc515.css
		values := strings.Split(referer, "/api/static/files/")
		refererFile := values[1]
		values = strings.Split(refererFile, "/")
		bindName := values[0]

		exist = utils.Exists(gvar.RootPath + "/files/" + bindName + "/" + fileHash + extension)
		if exist {
			ctx.Redirect(302, "/api/static/files/"+bindName+"/"+fileHash+extension)
			ctx.Abort()
			return
		}

		//redirect to server
		url := global.RequestNotExistFileUrl + "/" + bindName + "/" + filePath
		if config.GetString("apiProto") == "http" {
			url = "http://127.0.0.1:9090/api/v1/terminalfindfile" + "/" + bindName + "/" + filePath
		}
		ctx.Redirect(302, url)
		ctx.Abort()

		//下载文件
		localFilePath := gvar.RootPath + "/files/" + bindName + "/" + fileHash + extension
		err := downloadtaskmgr.DownLoadFile(url, localFilePath)
		if err != nil {
			logger.Error("download file url="+url+"error", "err", err)
			ctx.Abort()
		}
		ldb.SetAccessTimeStamp(bindName+"/"+fileHash+extension, time.Now().Unix())
	}
}
