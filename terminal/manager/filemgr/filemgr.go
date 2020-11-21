package filemgr

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var CdnSpaceLimit = uint64(config.GetInt("spacelimit") * 1000000000)
var CdnSpaceUsed uint64 = 0

func init() {
	//判断文件夹是否存在
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

	//创建标准文件
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

//扫描过期文件
func ScanExpirationFiles() {
	//向服务器请求过期时间
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

	//扫描过期的文件
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

	//向服务器发送要删除的文件列表
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
		//得到正确回复后删除文件
		logger.Debug("agree to delete files")
		for _, v := range expirationFils {
			//file := strings.Split(v, "/")
			os.Remove(global.FileDirPath + "/" + v)
			ldb.DB.Delete([]byte(v), nil)
		}
		//检查文件夹,如果已经没有文件,就删除文件夹
		dirs, _ := ioutil.ReadDir(global.FileDirPath)
		for _, v := range dirs {
			dirName := v.Name()
			if v.IsDir() {
				//检查文件夹是否已经空了
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
		//统计访问时间
		go ldb.SetAccessTimeStamp(filePath, time.Now().Unix())
	}
}
