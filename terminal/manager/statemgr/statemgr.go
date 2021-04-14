package statemgr

import (
	"encoding/json"
	"fmt"
	"github.com/daqnext/meson-common/common/accountmgr"
	"github.com/daqnext/meson-common/common/commonmsg"
	"github.com/daqnext/meson-common/common/httputils"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/resp"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/domainmgr"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"github.com/daqnext/meson-terminal/terminal/manager/versionmgr"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"runtime"
	"time"
)

var State = &commonmsg.TerminalStatesMsg{}
var ConsecutiveFailures = 0

//linux  ls -lact --full-time /etc | tail -1 |awk '{print $6,$7}'
//mac

var cpuUsageArray = []float64{}
var cpuUsageSum = float64(0)

var netBytesRecv uint64 = 0
var netBytesSent uint64 = 0

func LoopJob() {
	CalAverageNetSpeed()
	CalCpuAverageUsage()
}

func CalAverageNetSpeed() {
	go func() {
		for true {
			if n, err := net.IOCounters(false); err == nil && len(n) > 0 {
				sent := n[0].BytesSent
				recv := n[0].BytesRecv
				if netBytesRecv != 0 && netBytesSent != 0 {
					//State.NetInRate = (recv - netBytesRecv) / uint64(s.config.statsReportPeriod.Milliseconds()/1000)
					//State.NetOutRate = (sent - netBytesSent) / uint64(s.config.statsReportPeriod.Milliseconds()/1000)
					NetInRate := (recv - netBytesRecv) / uint64(5)
					NetOutRate := (sent - netBytesSent) / uint64(5)
					State.NetInMbs = float64(NetInRate*8) / float64(1e6)
					State.NetOutMbs = float64(NetOutRate*8) / float64(1e6)
					//fmt.Println(State.NetInMbs,"Mbs")
					//fmt.Println(State.NetOutMbs,"Mbs")
				}
				netBytesRecv = recv
				netBytesSent = sent
			}
			time.Sleep(time.Second * 5)
		}
	}()
}

func CalCpuAverageUsage() {
	go func() {
		for true {
			percent, err := cpu.Percent(time.Second, false)
			if err != nil || len(percent) <= 0 {
				logger.Error("failed to get cup usage", "err", err)
			} else {
				cpuUsageArray = append(cpuUsageArray, percent[0])
				cpuUsageSum += percent[0]
				if len(cpuUsageArray) > 10 {
					cpuUsageSum -= cpuUsageArray[0]
					cpuUsageArray = cpuUsageArray[1:]
				}
			}
			if cpuUsageSum > 0 && len(cpuUsageArray) > 0 {
				State.CpuUsage = cpuUsageSum / float64(len(cpuUsageArray))
				//logger.Debug("CpuUsage","value",State.CpuUsage,"sum",cpuUsageSum,"array",cpuUsageArray)
			}
			time.Sleep(time.Second * 5)
		}
	}()
}

func GetMachineState() *commonmsg.TerminalStatesMsg {
	if State.OS == "" {
		if h, err := host.Info(); err == nil {
			State.OS = fmt.Sprintf("%v:%v(%v):%v", h.OS, h.Platform, h.PlatformFamily, h.PlatformVersion)
		}
	}

	if State.MachineSetupTime == "" {
		State.MachineSetupTime = GetMachineSetupTime()
	}

	if State.CPU == "" {
		if c, err := cpu.Info(); err == nil {
			State.CPU = c[0].ModelName
		}
	}

	//percent, err := cpu.Percent(time.Second, false)
	//if err != nil {
	//	logger.Error("failed to get cup usage", "err", err)
	//} else {
	//	State.CpuUsage = percent[0]
	//}

	if v, err := mem.VirtualMemory(); err == nil {
		State.MemTotal = v.Total
		State.MemAvailable = v.Available
	}

	if d, err := disk.Usage("/"); err == nil {
		State.DiskTotal = d.Total
		State.DiskAvailable = d.Free
	}

	State.CdnDiskTotal = uint64(filemgr.CdnSpaceLimit)
	State.CdnDiskAvailable = State.CdnDiskTotal - uint64(filemgr.CdnSpaceUsed)

	if State.MacAddr == "" {
		if macAddr, err := utils.GetMainMacAddress(); err != nil {
			logger.Error("failed to get mac address", "err", err)
		} else {
			State.MacAddr = macAddr
		}
	}

	if State.Port == "" {
		State.Port = config.UsingPort
	}

	State.Version = versionmgr.Version

	return State
}

func GetMachineSetupTime() string {
	switch runtime.GOOS {
	case "linux":
		result, err := utils.RunCommand("ls", "-lact --full-time /etc | tail -1 |awk '{print $6,$7}'")
		if err != nil {
			logger.Debug("aws ec2 describe-addresses err", "err", err)
			return ""
		}
		return result
	case "windows":
		return "windows unknown"
	case "darwin":
		return "darwin unknown"
	}
	return "unknown"
}

func SendStateFail() {
	ConsecutiveFailures++
	if ConsecutiveFailures >= 6 {
		domainmgr.CheckAvailableDomain()
	}
}

func SendStateToServer() {
	defer panichandler.CatchPanicStack()

	machineState := GetMachineState()
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accountmgr.Token,
	}
	//提交请求
	content, err := httputils.Request("POST", domainmgr.UsingDomain+global.SendHeartBeatUrl, machineState, header)
	if err != nil {
		logger.Error("send terminalState to server error", "err", err)
		SendStateFail()
		return
	}
	//logger.Debug("response form server", "response string", string(content))
	var respBody resp.RespBody
	if err := json.Unmarshal(content, &respBody); err != nil {
		logger.Error("response from terminal unmarshal error", "err", err)
		SendStateFail()
		return
	}

	switch respBody.Status {
	case 0:
		ConsecutiveFailures = 0
		//logger.Debug("send State success")
	case 101: //auth error
		logger.Error("auth error,please restart terminal with correct username and password")
	case 106: //low version
		logger.Error("Your version need upgrade. Please download new version from meson.network ")
		versionmgr.CheckVersion()
	default:
		logger.Error("server error")
	}
}
