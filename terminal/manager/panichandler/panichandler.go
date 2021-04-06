package panichandler

import (
	"errors"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/enum/machinetype"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/domainmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/gin-gonic/gin"
	"runtime/debug"
	"time"
)

func Recover(c *gin.Context) {
	defer func() {
		CatchPanicStack()
		c.Abort()
	}()
	//加载完 defer recover，继续后续接口调用
	c.Next()
}

func CatchPanicStack() {

	errs := recover()
	if errs == nil {
		return
	}

	logger.Debug("Catch Error")
	// report to server
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}
	err, ok := errs.(error)
	if ok == false {
		err = errors.New("turn to error fail")
	}

	//str := string(debug.Stack())
	//logger.Debug("===error===")
	//logger.Debug("Catch error", "err", err)
	//logger.Debug(str)
	//logger.Debug("===error===")

	report := &commonmsg.PanicReportMsg{
		MachineType: machinetype.Terminal,
		TimeStamp:   time.Now().Unix(),
		Error:       err.Error(),
		Stack:       string(debug.Stack()),
	}

	_, err = httputils.Request("POST", domainmgr.UsingDomain+global.PanicReportUrl, report, header)
	if err != nil {
		logger.Error("report panic to server error", "err", err)
	}
}
