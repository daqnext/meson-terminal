package versionmgr

import (
	"encoding/json"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-terminal/terminal/manager/domainmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"runtime"
)

const Version = "2.1.0"

func GetOSInfo() (arch string, osInfo string) {
	arch = "amd64"
	switch runtime.GOARCH {
	case "386":
		arch = "386"
	}

	osInfo = "linux"
	switch runtime.GOOS {
	case "windows":
		osInfo = "windows"
	case "darwin":
		osInfo = "darwin"
	}

	return arch, osInfo
}

func GetTerminalVersionFromServer() (latestVersion string, allowVersion string, err error) {
	//check is there new version or not
	logger.Info("Check Version...")
	header := map[string]string{
		"Content-Type": "application/json",
	}
	respCtx, err := httputils.Request("GET", domainmgr.UsingDomain+global.RequestCheckVersion, nil, header)
	if err != nil {
		logger.Error("Request FileExpirationTime error", "err", err)
		return "", "", err
	}
	var respBody resp.RespBody
	if err := json.Unmarshal(respCtx, &respBody); err != nil {
		logger.Error("response from terminal unmarshal error", "err", err)
		return "", "", err
	}
	latestVersion = ((respBody.Data.(map[string]interface{}))["latestVersion"]).(string)
	allowVersion = ((respBody.Data.(map[string]interface{}))["allowVersion"]).(string)
	return latestVersion, allowVersion, nil
}
