package requestmgr

import "github.com/gin-gonic/gin"

var HandlerMap = map[string][]gin.HandlerFunc{}

type Method string

const GET Method = "GET"
const POST Method = "POST"

func BindHandler(ctx *gin.Context, method string, path string, handler ...gin.HandlerFunc) {
	hitKey := method + "^" + path
	_, exist := HandlerMap[hitKey]
	if exist == false {
		HandlerMap[hitKey] = []gin.HandlerFunc{}
	}
	HandlerMap[hitKey] = append(HandlerMap[hitKey], handler...)
}
