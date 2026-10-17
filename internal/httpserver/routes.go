//internal/httpserver/httpserver.go

package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"webaviable/storage"
)

func (s *HTTPServer) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	mux.HandleFunc("GET /urls", func(w http.ResponseWriter, r *http.Request) {
		urls, err := s.service.GetAllURLs()
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(urls); err != nil {
			return
		}

	})

	mux.HandleFunc("GET /urls/{alias}", func(w http.ResponseWriter, r *http.Request) {
		alias := r.PathValue("alias")

		url, err := s.service.GetURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				http.Error(w, "404", http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(url); err != nil {
			return
		}
	})

	mux.HandleFunc("GET /urls/{alias}/checks", func(w http.ResponseWriter, r *http.Request) {
		alias := r.PathValue("alias")

		checks, err := s.service.GetChecks(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				http.Error(w, "404", http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(checks); err != nil {
			return
		}
	})

	mux.HandleFunc("GET /urls/{alias}/check", func(w http.ResponseWriter, r *http.Request) {
		alias := r.PathValue("alias")

		checks, err := s.service.CheckURL(alias, r.Context())
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				http.Error(w, "404", http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(checks); err != nil {
			return
		}
	})

	mux.HandleFunc("POST /urls", func(w http.ResponseWriter, r *http.Request) {
		type NewURL struct {
			Alias string `json:"alias"`
			URL   string `json:"url"`
		}
		var url NewURL
		err := json.NewDecoder(r.Body).Decode(&url)
		if err != nil {
			http.Error(w, "can't read JSON:"+err.Error(), http.StatusBadRequest)
			return
		}
		if _, err := s.service.AddURL(url.Alias, url.URL); err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				http.Error(w, "that url already have", http.StatusConflict)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status": "created"}`))
	})

	mux.HandleFunc("DELETE /urls/{alias}", func(w http.ResponseWriter, r *http.Request) {
		alias := r.PathValue("alias")

		err := s.service.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				http.Error(w, "404", http.StatusNotFound)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
