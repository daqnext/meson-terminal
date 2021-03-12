package routerpath

import (
	"bufio"
	"fmt"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-common/common/utils"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"github.com/daqnext/meson-terminal/terminal/manager/ldb"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func FileRequestServer() *gin.Engine {
	cdnGin := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization")
	cdnGin.Use(cors.New(corsConfig))

	//http://bindname-terminaltag.shoppynext.com/xxxxxxxxx
	cdnGin.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".bin"})))
	cdnGin.Any("/*action", requestHandler)

	return cdnGin
}

func requestHandler(ctx *gin.Context) {
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
	storagePath := global.FileDirPath + "/" + bindName + filePath
	exist := utils.Exists(storagePath)
	if exist {
		//fileName := path.Base(filePath)
		filePath = strings.Replace(filePath, "-redirecter456gt", "", 1)
		//set access time
		go ldb.SetAccessTimeStamp(bindName+filePath, time.Now().Unix())
		transferCacheFileFS(ctx, storagePath)
		return
	}

	//if not exist
	//redirect to server
	serverUrl := global.ServerDomain + "/api/cdn/" + bindName + filePath
	//todo: modify cdn user path
	//serverUrl := fmt.Sprintf("https://%s.coldcdn.com%s",bindName,filePath)
	ctx.Redirect(302, serverUrl)
	return
}

func transferCacheFileFS(ctx *gin.Context, filePath string) {
	ServeFile(ctx.Writer, ctx.Request, filePath)
}

func transferCacheFile(ctx *gin.Context, filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		ctx.Status(404)
		return
	}
	fstate, _ := f.Stat()
	size := fstate.Size()
	buf := bufio.NewReader(f)
	type_b, _ := buf.Peek(512)
	ctx.Writer.Header().Set("Content-type", http.DetectContentType(type_b)) //set type
	ctx.Writer.Header().Set("Content-Length", strconv.FormatInt(size, 10))  //set size

	buf_b := make([]byte, 1024*1024)
	for {
		for global.PauseTransfer == true {
			fmt.Println("pausing") //only for dev
			time.Sleep(time.Millisecond * 100)
		}
		fmt.Println("transferfile") //only for dev
		n, err := buf.Read(buf_b)
		ctx.Writer.Write(buf_b[:n])
		//time.Sleep(time.Millisecond*100) //only for dev
		if err == io.EOF || n == 0 {
			break
		}
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
