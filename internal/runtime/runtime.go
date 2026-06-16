package runtime

import (
	"context"
	"fmt"
	"io"

	"github.com/timo-42/rayboard/internal/backend"
	"github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/store"
	"github.com/timo-42/rayboard/internal/backend/tracker"
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
	authService := auth.NewService(db.SQL)
	trackerService := tracker.NewService(db.SQL, authorizer)
	attachmentService := attachments.NewService(db.SQL, authorizer)
	commentService := comments.NewService(db.SQL, authorizer)
	backendServer := backend.NewServer(
		cfg.BackendAddr,
		backend.WithAuthService(authService),
		backend.WithAuthorizer(authorizer),
		backend.WithTrackerService(trackerService),
		backend.WithAttachmentService(attachmentService),
		backend.WithCommentService(commentService),
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
	authService := auth.NewService(db.SQL)
	trackerService := tracker.NewService(db.SQL, authorizer)
	attachmentService := attachments.NewService(db.SQL, authorizer)
	commentService := comments.NewService(db.SQL, authorizer)
	server := backend.NewServer(
		cfg.BackendAddr,
		backend.WithAuthService(authService),
		backend.WithAuthorizer(authorizer),
		backend.WithTrackerService(trackerService),
		backend.WithAttachmentService(attachmentService),
		backend.WithCommentService(commentService),
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

func runFrontend(ctx context.Context, cfg config.Config, stdout, stderr io.Writer) error {
	group, ctx := newServerGroup(ctx)
	server := frontend.NewServer(cfg.FrontendAddr, cfg.BackendURL)
	group.start("frontend", server.ListenAndServe, server.Shutdown)
	fmt.Fprintf(stdout, "frontend listening on http://%s\n", cfg.FrontendAddr)
	return group.wait(ctx, stderr)
}
