package security

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/runpath"
	"github.com/daqnext/meson-terminal/terminal/manager/downloader"
	"github.com/daqnext/meson-terminal/terminal/manager/statemgr"
	"os"
	"path/filepath"
	"time"
)

var PublicKey *rsa.PublicKey = nil
var KeyPath = "./meson_PublicKey.pem"

func DownloadAndInitPublicKey() error {
	//download publickey
	publicKeyPath := filepath.Join(runpath.RunPath, KeyPath)
	url := "https://assets.meson.network:10443/static/terminal/publickey/meson_PublicKey.pem"
	err := downloader.DownloadFile(url, publicKeyPath)
	if err != nil {
		logger.Error("download publicKey url="+url+"error", "err", err)
	}

	//publicKey
	err = InitPublicKey(publicKeyPath)
	if err != nil {
		return err
	}
	return nil
}

// ParsePublicKey
func ParsePublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	fp, err := os.Open(publicKeyPath)
	if err != nil {
		logger.Error("open publicKey error", "path", publicKeyPath)
		return nil, err
	}
	defer fp.Close()
	fileinfo, err := fp.Stat()
	if err != nil {
		logger.Error("get publicKey stat error", "path", publicKeyPath)
		return nil, err
	}
	buf := make([]byte, fileinfo.Size())
	_, err = fp.Read(buf)
	if err != nil {
		logger.Error("read publicKey error", "path", publicKeyPath)
		return nil, err
	}

	block, _ := pem.Decode(buf)
	if block == nil {
		return nil, errors.New("publicKey error")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pubKey.(*rsa.PublicKey), nil
}

func InitPublicKey(keyPath string) error {
	var err error
	PublicKey, err = ParsePublicKey(keyPath)
	if err != nil {
		logger.Error("InitPublicKey error", "err", err)
		return err
	}
	return nil
}

func ValidateSignature(signContent string, sign string) bool {
	if PublicKey == nil {
		err := InitPublicKey(KeyPath)
		if err != nil {
			return false
		}
	}

	hashed := sha256.Sum256([]byte(signContent))
	sig, _ := base64.StdEncoding.DecodeString(sign)
	err := rsa.VerifyPKCS1v15(PublicKey, crypto.SHA256, hashed[:], sig)
	if err != nil {
		logger.Error("rsa2 public check sign failed.", "err", err)
		return false
	}
	return true
}

func CheckRequestLegal(timeStamp int64, macAddr string, macSign string) bool {
	//make sure request is in 30s
	if time.Now().Unix() > timeStamp+30 {
		logger.Error("request past due")
		return false
	}

	if statemgr.State.MacAddr != macAddr {
		logger.Error("request mac address error")
		return false
	}

	pass := ValidateSignature(statemgr.State.MacAddr, macSign)
	if pass == false {
		logger.Error("ValidateSignature MacAddr fail")
		return false
	}

	return true
}
