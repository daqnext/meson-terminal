package ldb

import (
	"encoding/binary"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
)

var db *leveldb.DB

func init() {
	LevelDBInit()
}

func GetDB() *leveldb.DB {
	if db == nil {
		LevelDBInit()
	}
	return db
}

func LevelDBInit() {
	if !utils.Exists(global.LDBPath) {
		err := os.Mkdir(global.LDBPath, 0700)
		if err != nil {
			logger.Fatal("tempfile dir create failed, please create dir " + global.FileDirPath + " by manual")
		}
	}

	ldb, err := leveldb.OpenFile(global.LDBFile, nil)
	if err != nil {
		logger.Fatal("open level db error", "err", err)
	}
	db = ldb
}

func SetAccessTimeStamp(filePath string, timeStamp int64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(timeStamp))
	err := GetDB().Put([]byte(filePath), b, nil)
	if err != nil {
		logger.Error("leveldb put data error", "err", err, "filePath", filePath)
	}
}

func GetLastAccessTimeStamp(filePath string) int64 {
	data, err := GetDB().Get([]byte(filePath), nil)
	if err != nil {
		logger.Debug("leveldb data not find", "err", err)
		return 0
	} else {
		i := int64(binary.LittleEndian.Uint64(data))
		return i
	}
}

func Close() {
	logger.Info("levelDB close")
	GetDB().Close()
	db = nil
}
