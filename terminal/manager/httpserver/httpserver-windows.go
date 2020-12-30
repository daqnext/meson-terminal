// +build windows

package httpserver

import (
	"net/http"
)

func StartHttpServer(httpAddr string, handler http.Handler) error {
	srvHttp := &http.Server{
		Addr:    httpAddr,
		Handler: handler,
	}
	return srvHttp.ListenAndServe()
}

func StartHttpsServer(httpsAddr string, certFile string, keyFile string, handler http.Handler) error {
	srvHttps := &http.Server{
		Addr:    httpsAddr,
		Handler: handler,
	}
	return srvHttps.ListenAndServeTLS(certFile, keyFile)
}
