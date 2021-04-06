package global

const FileDirPath = "./files"
const SpaceHolderDir = "./spaceholder"
const LDBPath = "./ldb"
const LDBFile = "./ldb/index"

var ReportDownloadFinishUrl = "/api/v1/s/terminal/downloadfinish"
var ReportDownloadFailedUrl = "/api/v1/s/terminal/downloadfailed"
var SendHeartBeatUrl = "/api/v1/s/terminal/heartbeat"
var TerminalLoginUrl = "/api/v1/s/serverreg/terminallogin"
var RequestFileExpirationTimeUrl = "/api/v1/s/terminal/expirationtime"
var RequestToDeleteFilesUrl = "/api/v1/s/terminal/deletefiles"
var RequestCheckVersion = "/api/v1/common/terminalversion"
var PanicReportUrl = "/api/v1/common/panicreport"

var ApiPort = ""

var PauseMoment = int64(0)
var TerminalIsRunning = false
