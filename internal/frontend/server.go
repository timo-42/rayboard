package frontend

import (
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
)

//go:embed templates static
var assets embed.FS

type Server struct {
	http *http.Server
}

func NewServer(addr string, backendURL string) *Server {
	return &Server{
		http: &http.Server{
			Addr:    addr,
			Handler: NewHandler(backendURL),
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func NewHandler(backendURL string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /api/", backendProxy(backendURL))
	mux.Handle("POST /api/", backendProxy(backendURL))
	mux.Handle("PATCH /api/", backendProxy(backendURL))
	mux.Handle("DELETE /api/", backendProxy(backendURL))
	mux.Handle("GET /static/", http.FileServerFS(assets))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":      "ok",
			"service":     "frontend",
			"backend_url": backendURL,
		})
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		index(w, backendURL)
	})
	return mux
}

func backendProxy(backendURL string) http.Handler {
	target, err := url.Parse(backendURL)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "backend url is invalid", http.StatusBadGateway)
		})
	}
	return httputil.NewSingleHostReverseProxy(target)
}

func index(w http.ResponseWriter, backendURL string) {
	tpl, err := template.ParseFS(assets, "templates/index.html")
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tpl.Execute(w, map[string]string{
		"BackendURL": backendURL,
	})
}
