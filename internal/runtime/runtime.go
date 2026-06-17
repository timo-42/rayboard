package runtime

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/timo-42/rayboard/internal/backend"
	"github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/cronjobs"
	"github.com/timo-42/rayboard/internal/backend/events"
	"github.com/timo-42/rayboard/internal/backend/notifications"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/store"
	"github.com/timo-42/rayboard/internal/backend/tracker"
	"github.com/timo-42/rayboard/internal/backend/webhooks"
	"github.com/timo-42/rayboard/internal/config"
	"github.com/timo-42/rayboard/internal/frontend"
)

type Mode string

const (
	ModeCombined Mode = "combined"
	ModeBackend  Mode = "backend"
	ModeFrontend Mode = "frontend"
)

func Run(ctx context.Context, mode Mode, cfg config.Config, stdout, stderr io.Writer) error {
	switch mode {
	case ModeCombined:
		return runCombined(ctx, cfg, stdout, stderr)
	case ModeBackend:
		return runBackend(ctx, cfg, stdout, stderr)
	case ModeFrontend:
		return runFrontend(ctx, cfg, stdout, stderr)
	default:
		return fmt.Errorf("unsupported runtime mode %q", mode)
	}
}

func runCombined(ctx context.Context, cfg config.Config, stdout, stderr io.Writer) error {
	db, err := openAndMigrate(ctx, cfg, stdout)
	if err != nil {
		return err
	}
	defer db.Close()

	group, ctx := newServerGroup(ctx)

	authorizer := authz.NewSQLEvaluator(db.SQL)
	eventBus := events.NewBus()
	eventStore := events.NewStore(db.SQL)
	auditStore := audit.NewStore(db.SQL)
	authService := auth.NewService(db.SQL)
	trackerService := tracker.NewService(db.SQL, authorizer, tracker.WithEventBus(eventBus), tracker.WithEventStore(eventStore))
	attachmentService := attachments.NewService(db.SQL, authorizer, attachments.WithEventBus(eventBus), attachments.WithEventStore(eventStore))
	commentService := comments.NewService(db.SQL, authorizer, comments.WithEventBus(eventBus), comments.WithEventStore(eventStore))
	searchService := search.NewService(db.SQL, authorizer)
	openRouterService := openrouter.NewService(db.SQL)
	runStore := automation.NewRunStore(db.SQL)
	webhookService := webhooks.NewService(
		db.SQL,
		authorizer,
		webhooks.WithRunStore(runStore),
		webhooks.WithTrackerService(trackerService),
		webhooks.WithSearchService(searchService),
		webhooks.WithCommentService(commentService),
	)
	notificationService := notifications.NewService(db.SQL, notifications.WithEventStore(eventStore))
	group.startWorker("notifications", func() error {
		return runNotificationProcessor(ctx, notificationService, stderr)
	})
	cronService := cronjobs.NewService(
		db.SQL,
		authorizer,
		runStore,
		cronjobs.WithTrackerService(trackerService),
		cronjobs.WithSearchService(searchService),
		cronjobs.WithCommentService(commentService),
	)
	if err := cronService.StartScheduler(ctx); err != nil {
		return err
	}
	group.addShutdown(cronService.Shutdown)
	backendServer := backend.NewServer(
		cfg.BackendAddr,
		backend.WithAuthService(authService),
		backend.WithAuditStore(auditStore),
		backend.WithAuthorizer(authorizer),
		backend.WithTrackerService(trackerService),
		backend.WithAttachmentService(attachmentService),
		backend.WithCommentService(commentService),
		backend.WithCronService(cronService),
		backend.WithNotificationService(notificationService),
		backend.WithOpenRouterService(openRouterService),
		backend.WithSearchService(searchService),
		backend.WithWebhookService(webhookService),
	)
	frontendServer := frontend.NewServer(cfg.FrontendAddr, cfg.BackendURL)

	group.start("backend", backendServer.ListenAndServe, backendServer.Shutdown)
	group.start("frontend", frontendServer.ListenAndServe, frontendServer.Shutdown)

	fmt.Fprintf(stdout, "backend listening on http://%s\n", cfg.BackendAddr)
	fmt.Fprintf(stdout, "frontend listening on http://%s\n", cfg.FrontendAddr)

	return group.wait(ctx, stderr)
}

func runBackend(ctx context.Context, cfg config.Config, stdout, stderr io.Writer) error {
	db, err := openAndMigrate(ctx, cfg, stdout)
	if err != nil {
		return err
	}
	defer db.Close()

	group, ctx := newServerGroup(ctx)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	eventBus := events.NewBus()
	eventStore := events.NewStore(db.SQL)
	auditStore := audit.NewStore(db.SQL)
	authService := auth.NewService(db.SQL)
	trackerService := tracker.NewService(db.SQL, authorizer, tracker.WithEventBus(eventBus), tracker.WithEventStore(eventStore))
	attachmentService := attachments.NewService(db.SQL, authorizer, attachments.WithEventBus(eventBus), attachments.WithEventStore(eventStore))
	commentService := comments.NewService(db.SQL, authorizer, comments.WithEventBus(eventBus), comments.WithEventStore(eventStore))
	searchService := search.NewService(db.SQL, authorizer)
	openRouterService := openrouter.NewService(db.SQL)
	runStore := automation.NewRunStore(db.SQL)
	webhookService := webhooks.NewService(
		db.SQL,
		authorizer,
		webhooks.WithRunStore(runStore),
		webhooks.WithTrackerService(trackerService),
		webhooks.WithSearchService(searchService),
		webhooks.WithCommentService(commentService),
	)
	notificationService := notifications.NewService(db.SQL, notifications.WithEventStore(eventStore))
	group.startWorker("notifications", func() error {
		return runNotificationProcessor(ctx, notificationService, stderr)
	})
	cronService := cronjobs.NewService(
		db.SQL,
		authorizer,
		runStore,
		cronjobs.WithTrackerService(trackerService),
		cronjobs.WithSearchService(searchService),
		cronjobs.WithCommentService(commentService),
	)
	if err := cronService.StartScheduler(ctx); err != nil {
		return err
	}
	group.addShutdown(cronService.Shutdown)
	server := backend.NewServer(
		cfg.BackendAddr,
		backend.WithAuthService(authService),
		backend.WithAuditStore(auditStore),
		backend.WithAuthorizer(authorizer),
		backend.WithTrackerService(trackerService),
		backend.WithAttachmentService(attachmentService),
		backend.WithCommentService(commentService),
		backend.WithCronService(cronService),
		backend.WithNotificationService(notificationService),
		backend.WithOpenRouterService(openRouterService),
		backend.WithSearchService(searchService),
		backend.WithWebhookService(webhookService),
	)
	group.start("backend", server.ListenAndServe, server.Shutdown)
	fmt.Fprintf(stdout, "backend listening on http://%s\n", cfg.BackendAddr)
	return group.wait(ctx, stderr)
}

func openAndMigrate(ctx context.Context, cfg config.Config, stdout io.Writer) (*store.DB, error) {
	db, err := store.Open(ctx, cfg.DBPath)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	admin, err := auth.BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	fmt.Fprintf(stdout, "database ready at %s\n", cfg.DBPath)
	fmt.Fprintf(stdout, "POC admin credentials: username=%s password=%s\n", admin.Username, admin.Password)
	return db, nil
}

func runNotificationProcessor(ctx context.Context, service *notifications.Service, stderr io.Writer) error {
	if service == nil {
		return nil
	}
	process := func() {
		if _, err := service.ProcessPendingDomainEvents(ctx, 100); err != nil {
			fmt.Fprintf(stderr, "notification processor error: %v\n", err)
		}
		if _, err := service.ProcessPendingDeliveries(ctx, notifications.ProcessDeliveriesInput{Limit: 100}); err != nil {
			fmt.Fprintf(stderr, "notification delivery processor error: %v\n", err)
		}
	}
	process()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			process()
		}
	}
}

func runFrontend(ctx context.Context, cfg config.Config, stdout, stderr io.Writer) error {
	group, ctx := newServerGroup(ctx)
	server := frontend.NewServer(cfg.FrontendAddr, cfg.BackendURL)
	group.start("frontend", server.ListenAndServe, server.Shutdown)
	fmt.Fprintf(stdout, "frontend listening on http://%s\n", cfg.FrontendAddr)
	return group.wait(ctx, stderr)
}
