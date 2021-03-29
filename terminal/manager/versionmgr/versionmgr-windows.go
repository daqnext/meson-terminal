// +build windows

package versionmgr

import (
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/utils"
)

func CheckVersion() {
	defer panichandler.CatchPanicStack()
	//check is there new version or not
	latestVersion, allowVersion, err := GetTerminalVersionFromServer()
	if err != nil {
		logger.Info("Version check error, please check version on meson.network")
		return
	}

	vResult := utils.VersionCompare(Version, latestVersion)
	if vResult != -1 {
		logger.Info("Already Latest Version")
		return
	}

	// is meet allowVersion
	vResult = utils.VersionCompare(Version, allowVersion)
	if vResult != -1 {
		logger.Info("New version detected, please download new version. This version will be deprecated soon.")
		return
	}

	//need upgrade
	logger.Warn("This version is deprecated, please download new version on meson.network.")
}

//func DownloadNewVersion(fileName string,downloadUrl string,newVersion string) {
//	//download xxx.zip
//	err := downloadtaskmgr.DownLoadFile(downloadUrl, fileName)
//	if err != nil {
//
//	}
//
//	//unzip
//	targetDir := "./" + strings.Replace("fileName", ".zip", "", 1)
//	zipReader, err := zip.OpenReader(fileName)
//	if err != nil {
//		fmt.Println("OpenReader failed: ", err)
//		return
//	}
//	defer zipReader.Close()
//
//	for _, file := range zipReader.Reader.File {
//		if file.Name != "meson.exe" {
//			continue
//		}
//		zippedFile, err := file.Open()
//		if err != nil {
//			fmt.Println("Open error: ", err)
//			return
//		}
//		defer zippedFile.Close()
//		extractedFilePath := filepath.Join(targetDir, file.Name)
//		if file.FileInfo().IsDir() {
//			fmt.Println("mkdir: ", extractedFilePath)
//			os.MkdirAll(extractedFilePath, 777)
//		} else {
//			// 如果文件在目录中间，那么file.Name也会包含目录的
//			fmt.Println("unzip file: ", file.Name)
//			outputFile, err := os.OpenFile(extractedFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 777)
//			if err != nil {
//				fmt.Println(err)
//				return
//			}
//			defer outputFile.Close()
//			_, err = io.Copy(outputFile, zippedFile)
//			if err != nil {
//				fmt.Println("Err: ", err)
//				return
//			}
//		}
//	}
//
//	//cover old version file
//	input, err := ioutil.ReadFile(targetDir+"/meson.exe")
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	os.Remove("./meson.exe")
//	err = ioutil.WriteFile("./meson.exe", input, 777)
//	if err != nil {
//		fmt.Println("Error creating", "./meson.exe")
//		fmt.Println(err)
//		return
//	}
//	os.Remove("./v"+Version)
//	os.Create("./v"+newVersion)
//
//	os.RemoveAll(targetDir)
//	os.Remove(fileName)
//}
