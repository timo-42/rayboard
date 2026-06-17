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
	BackendURL    string
	Designs       []designOption
	Design        designVariant
	DesignPreview bool
}

type designOption struct {
	Path   string
	Label  string
	Active bool
}

type designVariant struct {
	Path        string
	Label       string
	Name        string
	Description string
	BodyClass   string
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
		if isAppRoute(r.URL.Path) {
			index(w, backendURL, "")
			return
		}
		http.NotFound(w, r)
	})
	return mux
}

func isAppRoute(path string) bool {
	switch path {
	case "/projects", "/profile", "/rbac", "/admin/rbac", "/search", "/automation", "/settings":
		return true
	}
	for _, prefix := range []string{"/projects/", "/issues/"} {
		if len(path) > len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func isDesignVariantPath(path string) bool {
	if len(path) != 2 || path[0] != '/' {
		return false
	}
	return path[1] >= '1' && path[1] <= '5'
}

func designOptions(selectedPath string) []designOption {
	designs := make([]designOption, 0, 5)
	for _, variant := range designVariants() {
		path := variant.Path
		designs = append(designs, designOption{
			Path:   path,
			Label:  variant.Label,
			Active: path == selectedPath,
		})
	}
	return designs
}

func selectedDesign(path string) designVariant {
	for _, variant := range designVariants() {
		if variant.Path == path {
			return variant
		}
	}
	return designVariant{
		Path:        "/",
		Label:       "Dashboard",
		Name:        "Dashboard",
		Description: "Operational overview for Rayboard projects, tickets, notifications, and automation.",
		BodyClass:   "production-dashboard",
	}
}

func designVariants() []designVariant {
	return []designVariant{
		{Path: "/1", Label: "1", Name: "Operations", Description: "Dense admin shell for repeated project and ticket work.", BodyClass: "design-operations"},
		{Path: "/2", Label: "2", Name: "Planning", Description: "Backlog-forward layout with stronger planning and prioritization cues.", BodyClass: "design-planning"},
		{Path: "/3", Label: "3", Name: "Automation", Description: "Automation-oriented workspace for hooks, scripts, and run feedback.", BodyClass: "design-automation"},
		{Path: "/4", Label: "4", Name: "Triage", Description: "High-contrast queue view for fast support and QA scanning.", BodyClass: "design-triage"},
		{Path: "/5", Label: "5", Name: "Executive", Description: "Quiet overview style for roadmap, risk, and delivery status reviews.", BodyClass: "design-executive"},
	}
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

func index(w http.ResponseWriter, backendURL string, selectedPath string) {
	tpl, err := template.ParseFS(assets, "templates/index.html")
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tpl.Execute(w, indexData{
		BackendURL:    backendURL,
		Designs:       designOptions(selectedPath),
		Design:        selectedDesign(selectedPath),
		DesignPreview: isDesignVariantPath(selectedPath),
	})
}
