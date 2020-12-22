package routerpath

import (
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"
)

func FileRequestApi(addr string, crtFileName string, keyFileName string) {
	cdnGin := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization")
	cdnGin.Use(cors.New(corsConfig))

	//http://bindname-terminaltag.shoppynext.com/xxxxxxxxx
	cdnGin.Any("/*action", func(ctx *gin.Context) {
		hostName := strings.Split(ctx.Request.Host, ".")[0]
		hostInfo := strings.Split(hostName, "-")
		bindName := hostInfo[0]
		fileName := ctx.Request.URL.String()

		isFileRequest := strings.Contains(fileName, "/api/static/files/")

		if bindName == "0" && !isFileRequest {
			targetUrl := "http://127.0.0.1:" + global.ApiPort
			Forward(targetUrl, ctx)
			return
		}

		method := ctx.Request.Method
		// not GET or HEAD
		if method != "GET" && method != "HEAD" {
			serverUrl := global.ServerDomain + "/api/cdn/" + bindName + fileName
			ctx.Redirect(302, serverUrl)
			return
		}

		//if request is a query
		queryPos := strings.Index(fileName, "?")
		if queryPos != -1 {
			serverUrl := global.ServerDomain + "/api/cdn/" + bindName + fileName
			ctx.Redirect(302, serverUrl)
			return
		}

		//isFileRequest
		if bindName == "0" && isFileRequest {
			path := ctx.Request.URL.Path
			requestFile := strings.Replace(path, "/api/static/", "", 1)
			ctx.File("./" + requestFile)
			return
		}

		//isRequestCachedFiles
		filePath := ctx.Request.URL.String()
		storagePath := global.FileDirPath + "/" + bindName + "/" + filePath
		exist := utils.Exists(storagePath)
		if exist {
			fileName := path.Base(filePath)
			fileName = strings.Replace(fileName, "-redirecter456gt", "", 1)
			//ctx.Writer.Header().Add("Content-Disposition", "attachment; filename="+fileName)
			//set access time
			go ldb.SetAccessTimeStamp("/"+bindName+"/"+filePath, time.Now().Unix())
			ctx.File(storagePath)
			return
		}

		//if not exist
		//redirect to server
		serverUrl := global.ServerDomain + "/api/cdn/" + bindName + "/" + filePath
		ctx.Redirect(302, serverUrl)
		return
	})

	err := cdnGin.RunTLS(addr, "./"+crtFileName, "./"+keyFileName)
	if err != nil {
		logger.Error("server start error", "err", err)
	}
}

func Forward(targetUrl string, ctx *gin.Context) {
	remote, err := url.Parse(targetUrl)
	if err != nil {
		logger.Error("err", "err", err)
	}
	logger.Debug("remote", "remote", remote)

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}
