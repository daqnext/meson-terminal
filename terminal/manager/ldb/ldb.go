package ldb

import (
	"encoding/binary"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strings"
	"sync"
	"time"
)

//var db *leveldb.DB

const LDBPath = "./ldb"
const LDBFile = "./ldb/index"

const AccessTimePrefix = "AccessTime-"
const TempTransferPrefix = "TempTransfer-"

var DBLock sync.Mutex

func init() {
	LevelDBInit()
}

func LevelDBInit() {
	if !utils.Exists(LDBPath) {
		err := os.Mkdir(LDBPath, 0700)
		if err != nil {
			logger.Fatal("file dir create failed, please create dir " + LDBPath + " by manual")
		}
	}
}

func OpenDB() (*leveldb.DB, error) {
	db, err := leveldb.OpenFile(LDBFile, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func setValue(key string, value []byte) error {
	DBLock.Lock()
	defer DBLock.Unlock()
	db, err := OpenDB()
	if err != nil {
		logger.Error("SetValue open level db error", "err", err)
		return err
	}
	defer db.Close()

	err = db.Put([]byte(key), value, nil)
	if err != nil {
		logger.Error("leveldb put data error", "err", err, "key", key, "value", value)
		return err
	}
	return nil
}

func deleteValue(key string) error {
	DBLock.Lock()
	defer DBLock.Unlock()
	db, err := OpenDB()
	if err != nil {
		logger.Error("deleteValue open level db error", "err", err)
		return err
	}
	defer db.Close()
	err = db.Delete([]byte(key), nil)
	if err != nil {
		logger.Error("deleteValue db.Delete error", "err", err)
		return err
	}
	return nil
}

func FindExpirationFiles(fileExpirationTime int64) ([]string, error) {
	//scan expiration time
	expirationFiles := []string{}

	DBLock.Lock()
	defer DBLock.Unlock()
	db, err := OpenDB()
	if err != nil {
		logger.Error("ScanExpirationFiles open level db error", "err", err)
		return nil, err
	}
	defer db.Close()
	iter := db.NewIterator(nil, nil)
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
			expirationFiles = append(expirationFiles, file)
		}
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		logger.Error("ScanExpirationFiles iter error", "err", err)
		//return nil,err
	}

	return expirationFiles, nil
}

//func FindExpirationTempTransferFile() ([]string,error){
//	expirationFiles := []string{}
//
//	DBLock.Lock()
//	defer DBLock.Unlock()
//	db, err := OpenDB()
//	if err != nil {
//		logger.Error("FindExpirationTempTransferFile open level db error", "err", err)
//		return nil,err
//	}
//	defer db.Close()
//	iter := db.NewIterator(util.BytesPrefix([]byte(TempTransferPrefix)), nil)
//	nowTime := time.Now().Unix()
//	for iter.Next() {
//		key := iter.Key()
//		file := string(key)
//		value := iter.Value()
//		deleteTime := int64(binary.LittleEndian.Uint64(value))
//		if deleteTime > nowTime {
//			expirationFiles = append(expirationFiles, file)
//		}
//	}
//	iter.Release()
//	err=iter.Error()
//	if err!=nil {
//		logger.Error("FindExpirationTempTransferFile iter error","err",err)
//		//return nil,err
//	}
//
//	return expirationFiles,nil
//}

func BatchRemoveKey(keys []string) error {
	DBLock.Lock()
	defer DBLock.Unlock()
	db, err := OpenDB()
	if err != nil {
		logger.Error("BatchRemoveKey open level db error", "err", err)
		return err
	}
	defer db.Close()
	batch := new(leveldb.Batch)
	for _, v := range keys {
		batch.Delete([]byte(v))
	}
	err = db.Write(batch, nil)
	if err != nil {
		logger.Error("BatchRemoveKey db.Write error", "err", err)
		return err
	}
	return nil
}

func SetAccessTimeStamp(key string, timeStamp int64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(timeStamp))
	err := setValue(key, b)
	if err != nil {
		logger.Error("SetAccessTimeStamp error", "err", err, "key", key)
	}
}

func DeleteAccessTimeStamp(key string) {
	err := deleteValue(key)
	if err != nil {
		logger.Error("DeleteAccessTimeStamp leveldb delete data error", "err", err, "key", key)
	}
}

//func SetDeleteTime(key string,timeStamp int64){
//	b := make([]byte, 8)
//	binary.LittleEndian.PutUint64(b, uint64(timeStamp))
//	err := setValue(key,b)
//	if err != nil {
//		logger.Error("leveldb put data error", "err", err, "key", key)
//	}
//}
//
//func CancelDeleteTime(key string){
//	err:=deleteValue(key)
//	if err != nil {
//		logger.Error("CancelDeleteTime leveldb delete data error", "err", err, "key", key)
//	}
//}
