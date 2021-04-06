package requestutil

import (
	"errors"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/imroc/req"
	"strconv"
	"time"
)

func SendPostRequest(url string, param req.Param, body interface{}, timeoutSecond int, v *resp.RespBody) error {
	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": "Basic " + accountmgr.Token,
	}
	r := req.New()
	r.SetTimeout(time.Duration(timeoutSecond) * time.Second)
	response, err := r.Post(url, param, authHeader, req.BodyJSON(body))
	if err != nil {
		logger.Error("request error", "err", err)
		return err
	}
	err = response.ToJSON(v)
	if err != nil {
		return err
	}
	if v.Status != 0 {
		return errors.New("post request error. \npath:" + url + "\nerror code:" + strconv.Itoa(v.Status) + "\nmsg:" + v.Msg)
	}
	return nil
}

func SendGetRequest() {

}
