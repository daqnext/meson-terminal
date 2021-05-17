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
var Using = config.Using

var others = map[string]int{}
var isChecking = false

var FixRegionD string

func init() {
	for i := 10; i < 50; i = i + 10 {
		k := utils.GetStringHash(strconv.Itoa(i))
		k = k[3:18]
		k = reverseString(k)
		k = "http://" + k + ".com"
		//fmt.Println(k)

		others[k] = 1
	}
	others[Using] = 1

	//fmt.Println(backupDomain)
}
func reverseString(s string) string {
	runes := []rune(s)

	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}

	return string(runes)
}

func CheckAvailable() {

	if isChecking {
		return
	}
	logger.Debug("Initializing...")
	isChecking = true
	defer func() {
		isChecking = false
	}()

	usingUrl := Using + healthPath
	checkResult := CheckOthers(usingUrl)
	if checkResult {
		GetFixRegion()
		return
	}

	logger.Debug("start to check others")
	others[Using] = 1
	time.Sleep(5 * time.Second)
	for i := 0; i < 2; i++ {
		for k := range others {
			checkUrl := k + healthPath
			checkResult = CheckOthers(checkUrl)
			if checkResult {
				Using = k
				logger.Debug("Find another", "another", Using)
				config.RecordConfigLineToFile(config.Server, Using)
				GetFixRegion()
				return
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}

	logger.Error("Please check network environment or download new version Terminal in https://meson.network")
}

func CheckOthers(url string) bool {
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
	url := Using + global.GetFixRegionServerUrl
	response, err := r.Get(url)
	if err != nil {
		logger.Error("request error", "err", err)
		FixRegionD = Using
		return
	}
	responseData := response.Response()
	responseStatusCode := responseData.StatusCode
	if responseStatusCode != 200 {
		FixRegionD = Using
		return
	}

	var respBody resp.RespBody
	err = response.ToJSON(&respBody)
	if err != nil {
		logger.Error("ToJSON error", "err", err)
		FixRegionD = Using
		return
	}

	switch respBody.Status {
	case 0:
		v := respBody.Data
		value, ok := v.(string)
		if ok == false {
			FixRegionD = Using
			return
		}
		if value == "" {
			FixRegionD = Using
			return
		}
		FixRegionD = "http://" + value
		return
	default:
		FixRegionD = Using
		return
	}
}
