//internal/httpserver/httpserver.go

package httpserver

import (
	"context"
	"net/http"
	"time"
	"webaviable/internal/service"
)

type HTTPServer struct {
	*http.Server
	service *service.Service
}

func NewHTTPServer(addr string, readTimeOut time.Duration, idleTimeOut time.Duration, writeTimeout time.Duration, service *service.Service) *HTTPServer {
	mux := http.NewServeMux()
	s := &HTTPServer{
		Server: &http.Server{
			Addr:         addr,
			ReadTimeout:  readTimeOut,
			IdleTimeout:  idleTimeOut,
			WriteTimeout: writeTimeout,
			Handler:      mux,
		}, service: service,
	}
	s.registerRoutes(mux)
	return s
}

func (s *HTTPServer) Run() error {
	return s.ListenAndServe()
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}
