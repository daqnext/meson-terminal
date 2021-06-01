package account

import (
	"bytes"
	"encoding/json"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"io/ioutil"
	"net/http"
	"strconv"
)

var Token string
var ServerRequestTest = make(chan bool, 1)

func TerminalLogin(url string, token string) {
	if Token != "" && len(Token) == 24 {
		return
	}

	postData := make(map[string]string)
	postData["token"] = token
	bytesData, _ := json.Marshal(postData)

	res, err := http.Post(
		url,
		"application/json;charset=utf-8",
		bytes.NewBuffer(bytesData),
	)
	if err != nil {
		logger.Fatal("Login failed Fatal error ", "err", err.Error())
		return
	}

	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Fatal("Login failed Fatal error ", "err", err.Error())
	}
	//logger.Debug("response form server", "response string", string(content))
	var respBody resp.RespBody
	if err := json.Unmarshal(content, &respBody); err != nil {
		logger.Error("response from terminal unmarshal error", "err", err)
		logger.Fatal("login failed", "err", err)
		return
	}

	switch respBody.Status {
	case 0:
		Token = respBody.Data.(string)
		logger.Debug("login success! ", "token", Token)
		logger.Info("login success!")
		accountmgr.Token = Token
		space := strconv.Itoa(config.UsingSpaceLimit)
		//config.RecordUserInputConfigToFile(Token, config.UsingPort, space)
		newConfigs := map[string]string{
			config.Token:      Token,
			config.Port:       config.UsingPort,
			config.SpaceLimit: space,
		}
		err := config.RecordConfigToFile(newConfigs)
		if err != nil {
			logger.Error("RecordConfigToFile error", "err", err)
		}
	default:
		logger.Fatal("Token error,please login the website to get token")
	}
}
