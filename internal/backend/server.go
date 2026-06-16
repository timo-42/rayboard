package backend

import (
	"context"
	"net/http"

	"github.com/timo-42/rayboard/internal/backend/httpjson"
)

type Server struct {
	http *http.Server
}

func NewServer(addr string) *Server {
	return &Server{
		http: &http.Server{
			Addr:    addr,
			Handler: NewHandler(),
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health)
	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpjson.Write(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "backend",
	})
}
