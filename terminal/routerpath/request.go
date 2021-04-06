package routerpath

import (
	"github.com/daqnext/meson-common/common/ginrouter"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/config"
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"github.com/daqnext/meson-terminal/terminal/manager/terminallogger"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"strings"
)

const DefaultGin = "default"
const CheckStartGin = "checkStart"

func init() {
	if config.GetString("ginMode") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if logger.BaseLogger == nil {
		terminallogger.InitLogger()
	}
	gin.DefaultWriter = logger.BaseLogger.Out

	//outer request gin
	defaultGin := ginrouter.New(DefaultGin)
	//send panic to server
	defaultGin.GinInstance.Use(panichandler.Recover)
	defaultGin.EnableDefaultCors()
	//http://bindname.coldcdn.com/xxxxxxxxx
	defaultGin.GinInstance.Any("/*action", requestHandler)
	//http://bindname-terminaltag.shoppynext.com/xxxxxxxxx
	defaultGin.GinInstance.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".bin"})))

	//inner check gin
	checkStartGin := ginrouter.New(CheckStartGin)
	//send panic to server
	checkStartGin.GinInstance.Use(panichandler.Recover)
}

var HandlerMap = map[string]func(ctx *gin.Context){
	"POST /api/v1/file/save":  saveNewFileHandler,
	"POST /api/v1/file/pause": pauseHandler,
	"GET /api/testapi/test":   testHandler,
	"GET /api/testapi/health": healthHandler,
}

func requestHandler(ctx *gin.Context) {
	hostName := strings.Split(ctx.Request.Host, ".")[0]
	hostInfo := strings.Split(hostName, "-")
	bindName := hostInfo[0]
	path := ctx.Request.URL.String()

	method := ctx.Request.Method
	// not GET or HEAD
	//if bindName!="0" && (method != "GET" && method != "HEAD") {
	//	serverUrl := global.ServerDomain + "/api/cdn/" + bindName + path
	//	ctx.Redirect(302, serverUrl)
	//	return
	//}

	//if request is a query
	//queryPos := strings.Index(path, "?")
	//if bindName!="0" && queryPos != -1 {
	//	serverUrl := global.ServerDomain + "/api/cdn/" + bindName + path
	//	ctx.Redirect(302, serverUrl)
	//	return
	//}

	//browser file request
	// https://bindName-tagxxxxxx.shoppynext.com:19091/filepath/filename
	if bindName != "0" {
		//isRequestCachedFiles
		requestCachedFilesHandler(ctx, bindName, path)
		return
	}

	//if speedTester request file
	// https://0-tagxxxxxx.shoppynext.com:19091/api/static/files/standardfile/100.bin
	if strings.Contains(path, "/api/static/files/") {
		path := ctx.Request.URL.Path
		requestFile := strings.Replace(path, "/api/static/", "", 1)
		ctx.File("./" + requestFile)
		return
	}

	//apiRequest form server
	// POST https://0-tagxxxxxx.shoppynext.com:19091/api/v1/file/save  timestamp+sign
	// POST https://0-tagxxxxxx.shoppynext.com:19091/api/v1/file/pause   timestamp+sign
	// GET https://0-tagxxxxxx.shoppynext.com:19091/api/testapi/test
	// GET https://0-tagxxxxxx.shoppynext.com:19091/api/testapi/health
	hitKey := method + " " + path
	handler, exist := HandlerMap[hitKey]
	if exist {
		handler(ctx)
		return
	}

	ctx.Status(404)
}
