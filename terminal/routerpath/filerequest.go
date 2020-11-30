package routerpath

import (
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func FileRequestApi(addr string, crtFileName string, keyFileName string) {

	cdnGin := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization")
	//cdnGin.Use(cors.New(corsConfig))

	// http://bindname-terminaltag.shoppynext.com/xxxxxxxxx
	cdnGin.Any("/*action", func(ctx *gin.Context) {

		hostName := strings.Split(ctx.Request.Host, ".")[0]
		hostInfo := strings.Split(hostName, "-")
		bindName := hostInfo[0]
		fileName := ctx.Request.URL.String()

		if bindName == "0" {
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

		simpleHostProxy := httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = "127.0.0.1:" + global.ApiPort
				req.URL.Path = "/api/static/files/" + bindName + fileName
				req.Host = ctx.Request.Host
			},
		}
		simpleHostProxy.ServeHTTP(ctx.Writer, ctx.Request)
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
