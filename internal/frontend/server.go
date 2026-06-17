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

type indexData struct {
	BackendURL string
	Designs    []designOption
}

type designOption struct {
	Path   string
	Label  string
	Active bool
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
	mux.Handle("PUT /api/", backendProxy(backendURL))
	mux.Handle("PATCH /api/", backendProxy(backendURL))
	mux.Handle("DELETE /api/", backendProxy(backendURL))
	mux.Handle("GET /docs", docsHandler())
	mux.Handle("GET /docs/", docsHandler())
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
		if r.URL.Path == "/" {
			index(w, backendURL, "")
			return
		}
		if isDesignVariantPath(r.URL.Path) {
			index(w, backendURL, r.URL.Path)
			return
		}
		http.NotFound(w, r)
	})
	return mux
}

func isDesignVariantPath(path string) bool {
	if len(path) != 2 || path[0] != '/' {
		return false
	}
	return path[1] >= '1' && path[1] <= '5'
}

func designOptions(selectedPath string) []designOption {
	designs := make([]designOption, 0, 5)
	for _, label := range []string{"1", "2", "3", "4", "5"} {
		path := "/" + label
		designs = append(designs, designOption{
			Path:   path,
			Label:  label,
			Active: path == selectedPath,
		})
	}
	return designs
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

func index(w http.ResponseWriter, backendURL string, selectedDesign string) {
	tpl, err := template.ParseFS(assets, "templates/index.html")
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tpl.Execute(w, indexData{
		BackendURL: backendURL,
		Designs:    designOptions(selectedDesign),
	})
}
