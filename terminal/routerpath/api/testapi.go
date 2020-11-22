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
}
