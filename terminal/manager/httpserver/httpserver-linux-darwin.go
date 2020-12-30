// +build linux darwin

package httpserver

import (
	"github.com/fvbock/endless"
	"net/http"
)

func StartHttpServer(httpAddr string, handler http.Handler) error {
	//return endless.ListenAndServe(httpAddr,handler)
	srvHttp := &http.Server{
		Addr:    httpAddr,
		Handler: handler,
	}
	return srvHttp.ListenAndServe()
}

func StartHttpsServer(httpsAddr string, certFile string, keyFile string, handler http.Handler) error {
	return endless.ListenAndServeTLS(httpsAddr, certFile, keyFile, handler)
}
