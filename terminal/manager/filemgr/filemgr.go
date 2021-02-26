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
	"github.com/shirou/gopsutil/disk"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

//var CdnSpaceLimit = int64(config.GetInt("spacelimit") * 1000000000)
var CdnSpaceLimit = int64(0)
var CdnSpaceUsed int64 = 0
var SpaceHoldFiles = []string{}
var HoldFileSize = int64(0)
var LeftSpace = int64(0)

const eachHoldFileSize = 100 * 1000 * 1000
const headSpace = 200 * 1000 * 1000

var lock sync.RWMutex

var array = make([]byte, 1000*1000)
var isNewStart = true

func Init() {
	CdnSpaceLimit = int64(config.UsingSpaceLimit * 1000000000)

	if CdnSpaceLimit < 40*1000000000 && config.GetString("runMode") != "local" {
		logger.Fatal("40GB disk space is the minimum.")
	}

	//is dir exist
	if !utils.Exists(global.FileDirPath) {
		err := os.Mkdir(global.FileDirPath, 0777)
		if err != nil {
			logger.Fatal("tempfile dir create failed, please create dir " + global.FileDirPath + " by manual")
		}
	}

	if !utils.Exists(global.SpaceHolderDir) {
		err := os.Mkdir(global.SpaceHolderDir, 0777)
		if err != nil {
			logger.Fatal("spaceHolder dir create failed, please create dir " + global.SpaceHolderDir + " by manual")
		}
	}

	if !utils.Exists(global.FileDirPath + "/standardfile") {
		err := os.Mkdir(global.FileDirPath+"/standardfile", 0777)
		if err != nil {
			logger.Fatal("tempfile dir create failed, please create dir " + global.FileDirPath + "/standardfile" + " by manual")
		}
	}

	//create std file
	createStdFile(1, "byte")
	createStdFile(1*1000*1000, "1")
	createStdFile(2*1000*1000, "2")
	createStdFile(5*1000*1000, "5")
	createStdFile(10*1000*1000, "10")
	createStdFile(50*1000*1000, "50")
	createStdFile(100*1000*1000, "100")

	SyncCdnDirSize()
	SyncHoldFileDirSize()

	d, err := disk.Usage("./")
	if err != nil {
		logger.Error("get disk usage error", "err", err)
	}
	free := d.Free

	total := CdnSpaceUsed + HoldFileSize + int64(free)
	if total < CdnSpaceLimit {
		logger.Fatal("Disk space is smaller than the value you set")
	}

	fmt.Println("Initializing system... ")
	//fmt.Println("This process will take several minutes, depending on the size of the cdn space you provide")

	FullSpace()
	isNewStart = false

}

func FullSpace() {
	//disk space holder
	if !utils.Exists(global.SpaceHolderDir) {
		err := os.Mkdir(global.SpaceHolderDir, 0777)
		if err != nil {
			logger.Error("spaceholder dir create failed")
		}
	}

	SyncCdnDirSize()

	lock.Lock()
	//scan exist holder files
	SpaceHoldFiles = []string{}
	HoldFileSize = 0

	holdFiles, err := ioutil.ReadDir(global.SpaceHolderDir)
	if err != nil {
		logger.Error("read space holder dir error", "err", err)
		return
	}
	for _, file := range holdFiles {
		SpaceHoldFiles = append(SpaceHoldFiles, global.SpaceHolderDir+"/"+file.Name())
		HoldFileSize += file.Size()
	}
	lock.Unlock()

	LeftSpace = CdnSpaceLimit - CdnSpaceUsed - HoldFileSize
	go func() {
		for LeftSpace-eachHoldFileSize > headSpace {
			lock.Lock()
			createSpaceHoldFile()
			LeftSpace = CdnSpaceLimit - CdnSpaceUsed - HoldFileSize

			if isNewStart {
				percent := (float64(CdnSpaceLimit-LeftSpace) / float64(CdnSpaceLimit)) * float64(100)
				fmt.Fprintf(os.Stdout, "scaning... %3.0f%%\r", percent)
			}
			lock.Unlock()
		}
	}()

}

func createSpaceHoldFile() {
	holdFileCount := len(SpaceHoldFiles)
	fileName := global.SpaceHolderDir + fmt.Sprintf("/%010d.bin", holdFileCount+1)
	f, err := os.Create(fileName)
	if err != nil {
		logger.Error("Create holderFile error", "err", err, "fileName", fileName)
		return
	}
	defer f.Close()
	//if err := f.Truncate(int64(eachHoldFileSize)); err != nil {
	//	logger.Error("Full holderFile error", "err", err, "fileName", fileName)
	//}
	for i := 0; i < 100; i++ {
		f.Write(array)
	}

	SpaceHoldFiles = append(SpaceHoldFiles, f.Name())
	HoldFileSize += int64(eachHoldFileSize)

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
	defer f.Close()
	if err := f.Truncate(int64(size)); err != nil {
		logger.Error("Full standardFile error", "err", err, "fileName", fileName)
	}

	//content:=make([]byte,size)
	//f.Write(content)
}

func SyncHoldFileDirSize() {
	size, err := utils.GetDirSize(global.SpaceHolderDir)
	if err != nil {
		logger.Error("get dir size error", "err", err)
	}
	HoldFileSize = int64(size)
}

func SyncCdnDirSize() {
	size, err := utils.GetDirSize(global.FileDirPath)
	if err != nil {
		logger.Error("get dir size error", "err", err)
	}
	CdnSpaceUsed = int64(size)
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
	iter := ldb.GetDB().NewIterator(nil, nil)
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
		//delay 5 minutes delete
		time.Sleep(5 * time.Minute)
		for _, v := range expirationFils {
			os.Remove(global.FileDirPath + "/" + v)
			ldb.GetDB().Delete([]byte(v), nil)
		}
		DeleteEmptyFolder()
		FullSpace()
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
		//hostName:=strings.Split(ctx.Request.Host,".")[0]
		//hostInfo:=strings.Split(hostName,"-")
		//bindName:=hostInfo[0]

		requestPath := ctx.Request.URL.String()
		filePath := strings.Replace(requestPath, "/api/static/files/", "", 1)
		exist := utils.Exists(global.FileDirPath + "/" + filePath)
		if exist {
			fileName := path.Base(filePath)
			fileName = strings.Replace(fileName, "-redirecter456gt", "", 1)
			ctx.Writer.Header().Add("Content-Disposition", "attachment; filename="+fileName)
			//set access time
			go ldb.SetAccessTimeStamp(filePath, time.Now().Unix())
			return
		}

		//if not exist
		//redirect to server
		serverUrl := global.ServerDomain + "/api/cdn/" + filePath
		ctx.Redirect(302, serverUrl)
		ctx.Abort()
	}
}

func DeleteEmptyFolder() {
	utils.DeleteEmptyFolders(global.FileDirPath)
}

func DeleteFolder(folderPath string) error {
	return os.RemoveAll(global.FileDirPath + "/" + folderPath)
}

func GenDiskSpace(fileSize int64) {
	lock.Lock()
	for LeftSpace <= fileSize+headSpace {
		holdFileCount := len(SpaceHoldFiles)
		fileName := global.SpaceHolderDir + fmt.Sprintf("/%010d.bin", holdFileCount)
		if utils.Exists(fileName) {
			err := os.Remove(fileName)
			if err != nil {
				logger.Error("delete space hold file error", "err", err)
			}
			SpaceHoldFiles = SpaceHoldFiles[:holdFileCount-1]
			LeftSpace += eachHoldFileSize
		}
	}
	lock.Unlock()
}
