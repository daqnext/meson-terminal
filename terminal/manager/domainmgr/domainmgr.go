package domainmgr

import (
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/imroc/req"
	"time"
)

var healthPath = "/api/testapi/health"
var UsingDomain = config.UsingServerDomain

var backupDomain = map[string]int{
	"http://mesonbackup1.com:9090": 1,
	"http://mesonbackup2.com:9090": 1,
	"http://coldcdn.com":           1,
	"http://mesonbackup4.com:9090": 1,
}

func CheckAvailableDomain() {
	logger.Debug("checking available domain")

	usingUrl := UsingDomain + healthPath
	checkResult := CheckDomain(usingUrl)
	if checkResult {
		return
	}

	logger.Info("domain not available, start to check backup domain")
	backupDomain[UsingDomain] = 1
	time.Sleep(5 * time.Second)
	for i := 0; i < 2; i++ {
		for k, _ := range backupDomain {
			checkUrl := k + healthPath
			checkResult = CheckDomain(checkUrl)
			if checkResult {
				UsingDomain = k
				logger.Info("Find available domain", "domain", UsingDomain)
				config.RecordConfigLineToFile(config.ServerDomain, UsingDomain)
				return
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}

	logger.Error("Server Domain not available")
	logger.Error("Please check network environment or download new version Terminal in https://meson.network")
}

func CheckDomain(url string) bool {
	r := req.New()
	r.SetTimeout(time.Duration(8) * time.Second)
	response, err := r.Get(url)
	if err != nil {
		logger.Error("request error", "err", err)
		return false
	}
	responseData := response.Response()
	responseStatusCode := responseData.StatusCode
	if responseStatusCode == 200 {
		return true
	} else {
		return false
	}
}
