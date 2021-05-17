package global

import (
	"github.com/daqnext/meson-common/common/runpath"
	"path/filepath"
)

var FileDirPath = filepath.Join(runpath.RunPath, "./files")
var SpaceHolderDir = filepath.Join(runpath.RunPath, "./spaceholder")

var ReportDownloadFinishUrl = "/api/v1/s/terminal/downloadfinish"
var ReportDownloadFailedUrl = "/api/v1/s/terminal/downloadfailed"
var ReportDownloadStartUrl = "/api/v1/s/terminal/downloadstart"
var ReportDownloadProcessUrl = "/api/v1/s/terminal/downloadprocess"

var SendHeartBeatUrl = "/api/v1/s/terminal/heartbeat"
var TerminalLoginUrl = "/api/v1/s/serverreg/terminallogin"
var RequestFileExpirationTimeUrl = "/api/v1/s/terminal/expirationtime"
var RequestToDeleteFilesUrl = "/api/v1/s/terminal/deletefiles"
var RequestCheckVersion = "/api/v1/common/terminalversion"
var PanicReportUrl = "/api/v1/common/panicreport"
var UploadFileRequestLog = "/api/v1/s/terminal/uploadlog"

var GetFixRegionServerUrl = "/api/v1/common/fixregion"

var HealthCheckPort = ""

var PauseMoment = int64(0)
var TerminalIsRunning = false
