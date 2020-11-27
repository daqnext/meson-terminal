package api

import (
	"github.com/daqnext/meson-common/common"
	"github.com/daqnext/meson-terminal/terminal/manager/filemgr"
	"github.com/daqnext/meson-terminal/terminal/manager/global"
	"net/http"
)

func init() {
	//you must initialize this
	common.AutoConfigRouter()

	//access time
	//common.GetMyRouter().Use(filemgr.AccessTime())

	//open gzip
	//common.GetMyRouter().Use(gzip.Gzip(gzip.DefaultCompression))

	common.GetMyRouter().Use(filemgr.PreHandler())
	// http://xxxx.com/api/static/files
	common.GetMyRouter().StaticFS("/files", http.Dir(global.FileDirPath))
}
