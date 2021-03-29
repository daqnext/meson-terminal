package panichandler

import (
	"errors"
	"fmt"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/enum/machinetype"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
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
		logger.Debug("turn to error fail")
		err = errors.New("turn to error fail")
	}

	str := string(debug.Stack())
	fmt.Println("===error===")
	fmt.Println(err)
	fmt.Println(str)
	fmt.Println("===error===")

	//machineStat:=statemgr.GetMachineState()
	report := &commonmsg.PanicReportMsg{
		MachineType: machinetype.Terminal,
		TimeStamp:   time.Now().Unix(),
		Error:       err.Error(),
		Stack:       string(debug.Stack()),
	}
	//report.OS=machineStat.OS
	//report.CPU = machineStat.CPU
	//report.Port = machineStat.Port
	//report.CdnDiskTotal = machineStat.CdnDiskTotal
	//report.CdnDiskAvailable = machineStat.CdnDiskAvailable
	//report.MacAddr = machineStat.MacAddr
	//report.MemTotal = machineStat.MemTotal
	//report.MemAvailable = machineStat.MemAvailable
	//report.DiskTotal = machineStat.DiskTotal
	//report.DiskAvailable = machineStat.DiskAvailable
	//report.Version = machineStat.Version
	//report.CpuUsage = machineStat.CpuUsage

	_, err = httputils.Request("POST", global.PanicReportUrl, report, header)
	if err != nil {
		logger.Error("report panic to server error", "err", err)
	}
}
