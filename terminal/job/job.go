package job

import (
	"fmt"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/statemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/terminallogger"
	"github.com/daqnext/meson-terminal/terminal/manager/versionmgr"
	"github.com/robfig/cron/v3"
	"math/rand"
	"time"
)

func StartPreJob() {

}

func StartLoopJob() {
	statemgr.LoopJob()
}

func StartScheduleJob() {
	c := cron.New(cron.WithSeconds())
	rand.Seed(time.Now().Unix())

	//heartbeat
	randSecond := rand.Intn(30)
	schedule := fmt.Sprintf("%d,%d * * * * *", randSecond, randSecond+30)
	jobId, err := c.AddFunc(schedule, statemgr.SendStateToServer)
	//c.AddJob(schedule,cron.NewChain(cron.Recover(cron.DefaultLogger)).Then(&statemgr.StateJob{}))
	if err != nil {
		logger.Error("ScheduleJob-"+"SendStateToServer"+" start error", "err", err)
	} else {
		logger.Debug("ScheduleJob-"+"SendStateToServer"+" start", "ID", jobId, "Schedule", schedule)
	}

	//version check
	schedule = fmt.Sprintf("0 %d/30 * * * *", rand.Intn(30))
	jobId, err = c.AddFunc(schedule, versionmgr.CheckVersion)
	if err != nil {
		logger.Error("ScheduleJob-"+"VersionCheck"+" start error", "err", err)
	} else {
		logger.Debug("ScheduleJob-"+"VersionCheck"+" start", "ID", jobId, "Schedule", schedule)
	}

	//sync folder size
	jobId, err = c.AddFunc("0 0 * * * *", filemgr.SyncCdnDirSize)
	if err != nil {
		logger.Error("ScheduleJob-"+"SyncCdnDirSize"+" start error", "err", err)
	} else {
		logger.Debug("ScheduleJob-"+"SyncCdnDirSize"+" start", "ID", jobId, "Schedule", "0 0 * * * *")
	}

	//scan expiration files  every 6 hours
	schedule = fmt.Sprintf("%d 0 0,6,12,18 * * *", rand.Intn(60))
	//schedule = fmt.Sprintf("%d * * * * *", rand.Intn(60))
	jobId, err = c.AddFunc(schedule, filemgr.ScanExpirationFiles)
	if err != nil {
		logger.Error("ScheduleJob-"+"ScanExpirationFiles"+" start error", "err", err)
	} else {
		logger.Debug("ScheduleJob-"+"ScanExpirationFiles"+" start", "ID", jobId, "Schedule", schedule)
	}

	//delete empty folder 1time/hour
	schedule = fmt.Sprintf("%d 0 * * * *", rand.Intn(60))
	jobId, err = c.AddFunc(schedule, filemgr.DeleteEmptyFolder)
	if err != nil {
		logger.Error("ScheduleJob-"+"DeleteEmptyFolder"+" start error", "err", err)
	} else {
		logger.Debug("ScheduleJob-"+"DeleteEmptyFolder"+" start", "ID", jobId, "Schedule", schedule)
	}

	//delete logger file 1time/day
	schedule = fmt.Sprintf("0 0 %d * * *", rand.Intn(24))
	jobId, err = c.AddFunc(schedule, terminallogger.DeleteTimeoutLog)
	if err != nil {
		logger.Error("ScheduleJob-"+"DeleteTimeoutLog"+" start error", "err", err)
	} else {
		logger.Debug("ScheduleJob-"+"DeleteTimeoutLog"+" start", "ID", jobId, "Schedule", schedule)
	}

	c.Start()
}
