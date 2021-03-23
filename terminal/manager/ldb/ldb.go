package ldb

import (
	"encoding/binary"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"sync"
)

//var db *leveldb.DB

func init() {
	LevelDBInit()
}

//func GetDB() *leveldb.DB {
//	if db == nil {
//		LevelDBInit()
//	}
//	return db
//}

var DBLock sync.Mutex

func LevelDBInit() {
	if !utils.Exists(global.LDBPath) {
		err := os.Mkdir(global.LDBPath, 0700)
		if err != nil {
			logger.Fatal("file dir create failed, please create dir " + global.FileDirPath + " by manual")
		}
	}
}

func OpenDB() (*leveldb.DB, error) {
	db, err := leveldb.OpenFile(global.LDBFile, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func SetAccessTimeStamp(filePath string, timeStamp int64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(timeStamp))
	DBLock.Lock()
	defer DBLock.Unlock()
	db, err := OpenDB()
	if err != nil {
		logger.Error("SetAccessTimeStamp open level db error", "err", err)
		return
	}
	defer db.Close()

	err = db.Put([]byte(filePath), b, nil)
	if err != nil {
		logger.Error("leveldb put data error", "err", err, "filePath", filePath)
	}
}
