package gvar

import (
	"path"
	"runtime"
)

var RootPath string //Project root path

func init() {
	_, file, _, _ := runtime.Caller(0)
	dir := path.Dir(file)
	RootPath = path.Join(dir, "/..")
}
