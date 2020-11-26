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
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/gin-gonic/gin"
	"os"
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
	createStdFile(1, "byte")
	createStdFile(1*1000*1000, "1")
	createStdFile(2*1000*1000, "2")
	createStdFile(3*1000*1000, "3")
	createStdFile(4*1000*1000, "4")
	createStdFile(5*1000*1000, "5")
	createStdFile(10*1000*1000, "10")
	createStdFile(15*1000*1000, "15")
	createStdFile(20*1000*1000, "20")
}

func createStdFile(size int, name string) {
	fileName := global.FileDirPath + "/" + fmt.Sprintf("standardfile/%s.bin", name)
	if utils.Exists(fileName) {
		return
	}
	f, err := os.Create(fileName)
	if err != nil {
		logger.Error("Create standardFile error", "err", err, "fileName", fileName)
		return
	}
	if err := f.Truncate(int64(size)); err != nil {
		logger.Error("Full standardFile error", "err", err, "fileName", fileName)
	}
	f.Close()
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

func IsFileExist(filePath string) bool {
	time := ldb.GetLastAccessTimeStamp(filePath)
	if time == 0 {
		return false
	}

	return true
}

func PreHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		requestPath := ctx.Request.URL.String()                               ///api/static/files/wr1cs5/vendor/@fortawesome/fontawesome-free/webfonts/fa-brands-400.woff2
		filePath := strings.Replace(requestPath, "/api/static/files/", "", 1) // wr1cs5/vendor/@fortawesome/fontawesome-free/webfonts/fa-brands-400.woff2
		exist := utils.Exists(global.FileDirPath + "/" + filePath)
		//exist:=IsFileExist(filePath)
		if exist {
			//set access time
			go ldb.SetAccessTimeStamp(filePath, time.Now().Unix())
			return
		}

		//if not exist
		strs := strings.Split(filePath, "/")
		bindName := strs[0]
		fileName := strings.Replace(filePath, bindName+"/", "", 1)

		//redirect to server
		url := global.RequestNotExistFileUrl + "/" + bindName + "/" + fileName
		logger.Debug("back to server", "url", url)
		if config.GetString("apiProto") == "http" {
			url = "http://127.0.0.1:9090/api/v1/terminalfindfile" + "/" + bindName + "/" + fileName
		}
		ctx.Redirect(302, url)
		ctx.Abort()

		//download file
		localFilePath := global.FileDirPath + "/" + bindName + "/" + fileName
		err := downloadtaskmgr.DownLoadFile(url, localFilePath)
		if err != nil {
			logger.Error("download file url="+url+"error", "err", err)
			ctx.Abort()
			return
		}
		ldb.SetAccessTimeStamp(bindName+"/"+fileName, time.Now().Unix())

	}
}

func DeleteEmptyFolder() {
	utils.DeleteEmptyFolders(global.FileDirPath)
}

func DeleteFolder(folderPath string) error {
	return os.RemoveAll(global.FileDirPath + "/" + folderPath)
}
