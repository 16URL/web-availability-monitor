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
			Handler:      corsMiddleware(mux),
		}, service: service,
	}
	s.registerRoutes(mux)
	return s
}
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "http://localhost:4321" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) Run() error {
	return s.ListenAndServe()
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}
