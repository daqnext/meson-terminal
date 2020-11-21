package global

import (
	"github.com/daqnext/meson-terminal/terminal/manager/config"
)

const FileDirPath = "./files"
const LDBPath = "./ldb"
const LDBFile = "./ldb/index"

var ServerDomain = config.UsingServerDomain
var ReportDownloadFinishUrl = ServerDomain + "/api/v1/s/terminal/downloadfinish"
var ReportDownloadFailedUrl = ServerDomain + "/api/v1/s/terminal/downloadfailed"
var SendHeartBeatUrl = ServerDomain + "/api/v1/s/terminal/heartbeat"
var SLoginUrl = ServerDomain + "/api/v1/s/serverreg/slogin"
var RequestFileExpirationTimeUrl = ServerDomain + "/api/v1/s/terminal/expirationtime"
var RequestToDeleteFilsUrl = ServerDomain + "/api/v1/s/terminal/deletefiles"
