package backend

import (
	"context"
	"net/http"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/httpjson"
)

type Server struct {
	http *http.Server
}

type options struct {
	auth *auth.Service
}

type Option func(*options)

func WithAuthService(service *auth.Service) Option {
	return func(options *options) {
		options.auth = service
	}
}

func NewServer(addr string, opts ...Option) *Server {
	return &Server{
		http: &http.Server{
			Addr:    addr,
			Handler: NewHandler(opts...),
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func NewHandler(opts ...Option) http.Handler {
	options := options{}
	for _, option := range opts {
		option(&options)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health)
	if options.auth != nil {
		registerAuthRoutes(mux, options.auth)
	}
	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpjson.Write(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "backend",
	})
}
