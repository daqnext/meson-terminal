package routerpath

import (
	"github.com/daqnext/meson-terminal/terminal/manager/panichandler"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"strings"
)

func RequestServer() *gin.Engine {
	cdnGin := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization")
	cdnGin.Use(cors.New(corsConfig))

	//send panic to server
	cdnGin.Use(panichandler.Recover)

	//http://bindname-terminaltag.shoppynext.com/xxxxxxxxx
	cdnGin.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".bin"})))
	cdnGin.Any("/*action", requestHandler)

	return cdnGin
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
