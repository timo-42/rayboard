package backend

import (
	"context"
	"net/http"

	"github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/cronjobs"
	"github.com/timo-42/rayboard/internal/backend/httpapi"
	"github.com/timo-42/rayboard/internal/backend/notifications"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/tracker"
	"github.com/timo-42/rayboard/internal/backend/webhooks"
)

type Server struct {
	http *http.Server
}

type options struct {
	auth          *auth.Service
	audit         *audit.Store
	authorizer    authz.Evaluator
	tracker       *tracker.Service
	attachments   *attachments.Service
	comments      *comments.Service
	createPages   *tracker.CreatePageService
	cron          *cronjobs.Service
	notifications *notifications.Service
	openrouter    *openrouter.Service
	search        *search.Service
	ticketHooks   *tracker.HookService
	webhooks      *webhooks.Service
}

type Option func(*options)

func WithAuthService(service *auth.Service) Option {
	return func(options *options) {
		options.auth = service
	}
}

func WithAuditStore(store *audit.Store) Option {
	return func(options *options) {
		options.audit = store
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

func WithCommentService(service *comments.Service) Option {
	return func(options *options) {
		options.comments = service
	}
}

func WithCreatePageService(service *tracker.CreatePageService) Option {
	return func(options *options) {
		options.createPages = service
	}
}

func WithCronService(service *cronjobs.Service) Option {
	return func(options *options) {
		options.cron = service
	}
}

func WithNotificationService(service *notifications.Service) Option {
	return func(options *options) {
		options.notifications = service
	}
}

func WithOpenRouterService(service *openrouter.Service) Option {
	return func(options *options) {
		options.openrouter = service
	}
}

func WithSearchService(service *search.Service) Option {
	return func(options *options) {
		options.search = service
	}
}

func WithTicketHookService(service *tracker.HookService) Option {
	return func(options *options) {
		options.ticketHooks = service
	}
}

func WithWebhookService(service *webhooks.Service) Option {
	return func(options *options) {
		options.webhooks = service
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

	return httpapi.NewHandler(httpapi.Options{
		Auth:          options.auth,
		Audit:         options.audit,
		Authorizer:    options.authorizer,
		Tracker:       options.tracker,
		Attachments:   options.attachments,
		Comments:      options.comments,
		CreatePages:   options.createPages,
		Cron:          options.cron,
		Notifications: options.notifications,
		OpenRouter:    options.openrouter,
		Search:        options.search,
		TicketHooks:   options.ticketHooks,
		Webhooks:      options.webhooks,
	})
}
