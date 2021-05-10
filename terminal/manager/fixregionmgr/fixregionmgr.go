package fixregionmgr

import (
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/imroc/req"
	"strconv"
	"time"
)

var healthPath = "/api/testapi/health"
var UsingDomain = config.UsingServerDomain

var backupDomain = map[string]int{}
var isCheckingAvailableDomain = false

var FixRegionD string

func init() {
	for i := 10; i < 50; i = i + 10 {
		k := utils.GetStringHash(strconv.Itoa(i))
		k = k[3:18]
		k = reverseString(k)
		k = "http://" + k + ".com"
		//fmt.Println(k)

		backupDomain[k] = 1
	}
	backupDomain[UsingDomain] = 1

	//fmt.Println(backupDomain)
}
func reverseString(s string) string {
	runes := []rune(s)

	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}

	return string(runes)
}

func CheckAvailableDomain() {

	if isCheckingAvailableDomain {
		return
	}
	logger.Debug("checking available domain")
	isCheckingAvailableDomain = true
	defer func() {
		isCheckingAvailableDomain = false
	}()

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
				GetFixRegion()
				return
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}

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
	if responseStatusCode != 200 {
		return false
	}

	var respBody resp.RespBody
	err = response.ToJSON(&respBody)
	if err != nil {
		logger.Error("ToJSON error", "err", err)
		return false
	}

	switch respBody.Status {
	case 0:
		return true
	default:
		return false
	}

}

func GetFixRegion() {
	r := req.New()
	r.SetTimeout(time.Duration(8) * time.Second)
	url := UsingDomain + global.GetFixRegionServerUrl
	response, err := r.Get(url)
	if err != nil {
		logger.Error("request error", "err", err)
		FixRegionD = UsingDomain
		return
	}
	responseData := response.Response()
	responseStatusCode := responseData.StatusCode
	if responseStatusCode != 200 {
		FixRegionD = UsingDomain
		return
	}

	var respBody resp.RespBody
	err = response.ToJSON(&respBody)
	if err != nil {
		logger.Error("ToJSON error", "err", err)
		FixRegionD = UsingDomain
		return
	}

	switch respBody.Status {
	case 0:
		v := respBody.Data
		domain, ok := v.(string)
		if ok == false {
			FixRegionD = UsingDomain
			return
		}
		FixRegionD = domain
	default:
		FixRegionD = UsingDomain
		return
	}
}
