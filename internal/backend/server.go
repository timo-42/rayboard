package backend

import (
	"context"
	"net/http"

	"github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpjson"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type Server struct {
	http *http.Server
}

type options struct {
	auth        *auth.Service
	authorizer  authz.Evaluator
	tracker     *tracker.Service
	attachments *attachments.Service
}

type Option func(*options)

func WithAuthService(service *auth.Service) Option {
	return func(options *options) {
		options.auth = service
	}
}

func WithAuthorizer(authorizer authz.Evaluator) Option {
	return func(options *options) {
		options.authorizer = authorizer
	}
}

func WithTrackerService(service *tracker.Service) Option {
	return func(options *options) {
		options.tracker = service
	}
}

func WithAttachmentService(service *attachments.Service) Option {
	return func(options *options) {
		options.attachments = service
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
		registerAuthRoutes(mux, options.auth, options.authorizer)
	}
	if options.auth != nil && options.tracker != nil {
		registerTrackerRoutes(mux, options.auth, options.tracker)
	}
	if options.auth != nil && options.attachments != nil {
		registerAttachmentRoutes(mux, options.auth, options.attachments)
	}
	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpjson.Write(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "backend",
	})
}
