package api

import (
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-common/common/logger"
	"github.com/daqnext/meson-terminal/terminal/manager/account"
	"github.com/gin-gonic/gin"
	"time"
)

func init() {
	//you must initialize this
	common.AutoConfigRouter()

	// http://xxxx.com/api/testapi/test
	common.GetMyRouter().GET("/test", func(context *gin.Context) {
		logger.Debug("Get test Request form Server")
		if account.ServerRequestTest != nil {
			account.ServerRequestTest <- true
		}
		context.JSON(200, gin.H{
			"status": 0,
			"time":   time.Now().Format("2006-01-02 15:04:05.000"),
		})
	})

	common.GetMyRouter().GET("/health", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"status": 0,
		})
	})

	//common.GetMyRouter().GET("/savefile", func(context *gin.Context) {
	//
	//	//localFilePath := global.FileDirPath + "/" + "testdir" + "/" + "assets/img/homebrew-256x256.png"
	//	//url := "https://brew.sh/assets/img/homebrew-256x256.png"
	//	//err := downloadtaskmgr.DownLoadFile(url, localFilePath)
	//	//if err != nil {
	//	//	logger.Error("download file url="+url+"error", "err", err)
	//	//}
	//
	//	context.JSON(200, gin.H{
	//		"status": 0,
	//	})
	//})

	//common.GetMyRouter().GET("/devtest", func(ctx *gin.Context) {
	//	//for k, v := range ctx.Request.Header {
	//	//	fmt.Println(k, v)
	//	//}
	//
	//	ctx.JSON(200, gin.H{
	//		"status": 0,
	//	})
	//})
}
