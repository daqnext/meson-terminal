package api

import (
	"github.com/daqnext/meson-common/common"
	"github.com/gin-gonic/gin"
)

func init() {
	//you must initialize this
	common.AutoConfigRouter()

	// http://xxxx.com/api/testapi/health
	common.GetMyRouter().GET("/health", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"status": 0,
		})
	})

}
