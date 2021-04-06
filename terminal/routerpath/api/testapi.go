package api

import (
	"github.com/daqnext/meson-common/common/ginrouter"
	"github.com/daqnext/meson-terminal/terminal/routerpath"
	"github.com/gin-gonic/gin"
)

func init() {
	//you must initialize this
	checkStartGin := ginrouter.GetGinInstance(routerpath.CheckStartGin)
	checkStartGin.AutoConfigRouter()

	// http://xxxx.com/api/testapi/health
	checkStartGin.GetMyRouter().GET("/health", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"status": 0,
		})
	})
}
