package filemgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/fixregionmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"github.com/shirou/gopsutil/v3/disk"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"
)

//var CdnSpaceLimit = int64(config.GetInt("spacelimit") * 1000000000)
var CdnSpaceLimit = int64(0)
var CdnSpaceUsed int64 = 0
var SpaceHoldFiles = []string{}
var HoldFileSize = int64(0)
var LeftSpace = int64(0)

const UnitK = 1 << 10
const UnitM = 1 << 20
const UnitG = 1 << 30

const eachHoldFileSize = 100 * UnitM
const headSpace = 200 * UnitM

var lock sync.RWMutex

var array = make([]byte, UnitM)
var isNewStart = true

func Init() error {
	CdnSpaceLimit = int64(config.UsingSpaceLimit * UnitG)

	if CdnSpaceLimit < 40*UnitG && config.GetString("runMode") != "local" {
		logger.Error("40GB disk space is the minimum.")
		return errors.New("disk space error")
	}

	//is dir exist
	if !utils.Exists(global.FileDirPath) {
		err := os.Mkdir(global.FileDirPath, 0777)
		if err != nil {
			logger.Error("tempfile dir create failed, please create dir " + global.FileDirPath + " by manual or try to run program with admin permission.")
			return errors.New("FileDirPath error")
		}
	}

	if !utils.Exists(global.SpaceHolderDir) {
		err := os.Mkdir(global.SpaceHolderDir, 0777)
		if err != nil {
			logger.Error("spaceHolder dir create failed, please create dir " + global.SpaceHolderDir + " by manual or try to run program with admin permission.")
			return errors.New("SpaceHolderDir error")
		}
	}

	if !utils.Exists(global.FileDirPath + "/standardfile") {
		err := os.Mkdir(global.FileDirPath+"/standardfile", 0777)
		if err != nil {
			logger.Error("tempfile dir create failed, please create dir " + global.FileDirPath + "/standardfile" + " by manual or try to run program with admin permission.")
			return errors.New("standardfile error")
		}
	}

	//create std file
	createStdFile(1, "byte")
	createStdFile(1*UnitM, "1")
	createStdFile(2*UnitM, "2")
	createStdFile(5*UnitM, "5")
	createStdFile(10*UnitM, "10")
	createStdFile(50*UnitM, "50")
	createStdFile(100*UnitM, "100")

	SyncCdnDirSize()
	SyncHoldFileDir()

	free := uint64(0)
	d, err := disk.Usage("./")
	if err != nil {
		logger.Error("get disk usage error", "err", err)
	} else {
		free = d.Free
	}

	total := CdnSpaceUsed + HoldFileSize + int64(free)
	if total < CdnSpaceLimit {
		logger.Error("Disk space is smaller than the value you set")
		return errors.New("Disk space error")
	}

	fmt.Println("Initializing system... ")
	//fmt.Println("This process will take several minutes, depending on the size of the cdn space you provide")

	FullSpace()
	isNewStart = false

	return nil
}

func SyncCdnDirSize() {
	defer panichandler.CatchPanicStack()
	size, err := utils.GetDirSize(global.FileDirPath)
	if err != nil {
		logger.Error("get dir size error", "err", err)
	}
	CdnSpaceUsed = int64(size)
}

func SyncHoldFileDir() {
	defer panichandler.CatchPanicStack()
	//disk space holder
	if !utils.Exists(global.SpaceHolderDir) {
		err := os.Mkdir(global.SpaceHolderDir, 0777)
		if err != nil {
			logger.Error("spaceholder dir create failed")
			return
		}
	}

	holdFiles, err := ioutil.ReadDir(global.SpaceHolderDir)
	if err != nil {
		logger.Error("read space holder dir error", "err", err)
		return
	}

	for _, file := range holdFiles {
		SpaceHoldFiles = append(SpaceHoldFiles, file.Name())
		HoldFileSize += file.Size()
	}
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

	//lock.Lock()
	//scan exist holder files
	//SpaceHoldFiles = []string{}
	//HoldFileSize = 0

	//holdFiles, err := ioutil.ReadDir(global.SpaceHolderDir)
	//if err != nil {
	//	lock.Unlock()
	//	logger.Error("read space holder dir error", "err", err)
	//	return
	//}
	//for _, file := range holdFiles {
	//	SpaceHoldFiles = append(SpaceHoldFiles, global.SpaceHolderDir+"/"+file.Name())
	//	HoldFileSize += file.Size()
	//}
	//lock.Unlock()

	LeftSpace = CdnSpaceLimit - CdnSpaceUsed - HoldFileSize

	if LeftSpace > 0 {
		go func() {
			defer panichandler.CatchPanicStack()
			for LeftSpace-eachHoldFileSize > headSpace {
				lock.Lock()
				createSpaceHoldFile()
				LeftSpace = CdnSpaceLimit - CdnSpaceUsed - HoldFileSize

				if isNewStart {
					percent := (float64(CdnSpaceLimit-LeftSpace) / float64(CdnSpaceLimit)) * float64(100)
					fmt.Fprintf(os.Stdout, "scaning... %3.0f%%\r", percent)
				}
				lock.Unlock()
				time.Sleep(500 * time.Millisecond)
			}
		}()
	} else {
		releaseSpace := -LeftSpace + headSpace
		ReleaseDiskSpace(releaseSpace)
	}

}

func createSpaceHoldFile() {
	holdFileCount := len(SpaceHoldFiles)
	name := fmt.Sprintf("%010d%d", holdFileCount+1, rand.Intn(99999999))
	name = utils.GetStringHash(name)
	name = name + ".bin"
	fileName := global.SpaceHolderDir + "/" + name
	f, err := os.Create(fileName)
	if err != nil {
		logger.Error("Create holderFile error", "err", err, "fileName", fileName)
		return
	}
	defer f.Close()
	fileSize := int64(0)
	for i := 0; i < 100; i++ {
		size, err := f.Write(array)
		if err != nil {
			logger.Error("createSpaceHoldFile error", "err", err)
			return
		}
		fileSize += int64(size)
	}

	SpaceHoldFiles = append(SpaceHoldFiles, name)
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

func ScanExpirationFiles() {
	defer panichandler.CatchPanicStack()
	//request expiration time from server
	logger.Info("Start ScanExpirationFiles")
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}

	respCtx, err := httputils.Request("GET", fixregionmgr.FixRegionD+global.RequestFileExpirationTimeUrl, nil, header)
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
		logger.Error("Request FileExpirationTime response err", "respBody.Status", respBody.Status)
		return
	}

	expirationFiles, err := ldb.FindExpirationFiles(fileExpirationTime)
	if err != nil {
		return
	}
	if len(expirationFiles) == 0 {
		return
	}

	//post delete file list to server
	header = map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}
	payload := commonmsg.TerminalRequestDeleteFilesMsg{
		Files: expirationFiles,
	}
	respCtx, err = httputils.Request("POST", fixregionmgr.FixRegionD+global.RequestToDeleteFilesUrl, payload, header)
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
		//delay 15 minutes delete
		time.Sleep(15 * time.Minute)

		err := ldb.BatchRemoveKey(expirationFiles)
		if err != nil {
			logger.Error("ScanExpirationFiles leveldb batch delete error", "err", err)
		}
		for _, v := range expirationFiles {
			os.Remove(global.FileDirPath + "/" + v)
		}
		DeleteEmptyFolder()
		FullSpace()
	default:
		logger.Error("Request FileExpirationTime response", "response", respBody2)
		return
	}

}

func DeleteEmptyFolder() {
	defer panichandler.CatchPanicStack()
	utils.DeleteEmptyFolders(global.FileDirPath)
}

func DeleteFile(bindName string, fileName string) error {
	fixFileName := utils.FileAddMark(fileName, common.RedirectMark)
	dir := global.FileDirPath + "/" + bindName

	savePath := dir + "/" + fixFileName
	if !utils.Exists(savePath) {
		return nil
	}

	//delete ldb record
	go func() {
		panichandler.CatchPanicStack()
		ldb.DeleteAccessTimeStamp(bindName + "/" + fixFileName)
	}()

	err := os.Remove(savePath)
	if err != nil {
		logger.Error("delete file error", "err", err, "file", savePath)
		return err
	}

	return nil
}

func ReleaseDiskSpace(releaseSize int64) bool {
	released := int64(0)
	for released < releaseSize {
		holdFileCount := len(SpaceHoldFiles)
		if holdFileCount <= 0 {
			logger.Error("ReleaseDiskSpace space not enough")
			return false
		}
		name := SpaceHoldFiles[holdFileCount-1]
		fileName := global.SpaceHolderDir + "/" + name
		if utils.Exists(fileName) {
			fileStat, err := os.Stat(fileName)
			size := int64(0)
			if err != nil {
				logger.Error("ReleaseDiskSpace get file stat error", "err", err)
			} else {
				size = fileStat.Size()
			}

			err = os.Remove(fileName)
			if err != nil {
				logger.Error("GenDiskSpace delete space hold file error", "err", err)
			} else {
				LeftSpace += size
				released += size
			}
		}
		SpaceHoldFiles = SpaceHoldFiles[:holdFileCount-1]
	}
	return true
}

func GenDiskSpace(fileSize int64) bool {
	lock.Lock()
	defer lock.Unlock()
	for LeftSpace <= fileSize+headSpace {
		holdFileCount := len(SpaceHoldFiles)
		if holdFileCount <= 0 {
			logger.Error("space not enough")
			return false
		}
		name := SpaceHoldFiles[holdFileCount-1]
		fileName := global.SpaceHolderDir + "/" + name
		if utils.Exists(fileName) {
			fileStat, err := os.Stat(fileName)
			size := int64(0)
			if err != nil {
				logger.Error("GenDiskSpace get file stat error", "err", err)
			} else {
				size = fileStat.Size()
			}

			err = os.Remove(fileName)
			if err != nil {
				logger.Error("GenDiskSpace delete space hold file error", "err", err)
			} else {
				LeftSpace += size
			}
		}
		SpaceHoldFiles = SpaceHoldFiles[:holdFileCount-1]
	}
	return true
}
