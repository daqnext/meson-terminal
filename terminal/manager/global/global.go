package global

import (
	"github.com/daqnext/meson-terminal/terminal/manager/config"
)

const FileDirPath = "./files"
const SpaceHolderDir = "./spaceholder"
const LDBPath = "./ldb"
const LDBFile = "./ldb/index"

var ServerDomain = config.UsingServerDomain
var ReportDownloadFinishUrl = ServerDomain + "/api/v1/s/terminal/downloadfinish"
var ReportDownloadFailedUrl = ServerDomain + "/api/v1/s/terminal/downloadfailed"
var SendHeartBeatUrl = ServerDomain + "/api/v1/s/terminal/heartbeat"
var SLoginUrl = ServerDomain + "/api/v1/s/serverreg/slogin"
var TerminalLoginUrl = ServerDomain + "/api/v1/s/serverreg/terminallogin"
var RequestFileExpirationTimeUrl = ServerDomain + "/api/v1/s/terminal/expirationtime"
var RequestToDeleteFilsUrl = ServerDomain + "/api/v1/s/terminal/deletefiles"
var RequestNotExistFileUrl = ServerDomain + "/api/v1/terminalfindfile"
var RequestCheckVersion = ServerDomain + "/api/v1/common/terminalversion"

var FilePort = ""
var ApiPort = ""

var PauseTransfer = false
var TerminalIsRunning = false
