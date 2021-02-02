module github.com/daqnext/meson-terminal

go 1.15

require (
	github.com/daqnext/meson-common v1.0.9
	github.com/fvbock/endless v0.0.0-20170109170031-447134032cb6
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-contrib/gzip v0.0.3
	github.com/gin-gonic/gin v1.6.3
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil v2.20.5+incompatible
	github.com/sirupsen/logrus v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca
)

replace github.com/daqnext/meson-common => /Users/zhangzhenbo/workspace/go/project/meson-common
