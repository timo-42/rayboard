package runtime

import (
	"context"
	"fmt"
	"io"

	"github.com/timo-42/rayboard/internal/backend"
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
	group, ctx := newServerGroup(ctx)

	backendServer := backend.NewServer(cfg.BackendAddr)
	frontendServer := frontend.NewServer(cfg.FrontendAddr, cfg.BackendURL)

	group.start("backend", backendServer.ListenAndServe, backendServer.Shutdown)
	group.start("frontend", frontendServer.ListenAndServe, frontendServer.Shutdown)

	fmt.Fprintf(stdout, "backend listening on http://%s\n", cfg.BackendAddr)
	fmt.Fprintf(stdout, "frontend listening on http://%s\n", cfg.FrontendAddr)

	return group.wait(ctx, stderr)
}

func runBackend(ctx context.Context, cfg config.Config, stdout, stderr io.Writer) error {
	group, ctx := newServerGroup(ctx)
	server := backend.NewServer(cfg.BackendAddr)
	group.start("backend", server.ListenAndServe, server.Shutdown)
	fmt.Fprintf(stdout, "backend listening on http://%s\n", cfg.BackendAddr)
	return group.wait(ctx, stderr)
}

func runFrontend(ctx context.Context, cfg config.Config, stdout, stderr io.Writer) error {
	group, ctx := newServerGroup(ctx)
	server := frontend.NewServer(cfg.FrontendAddr, cfg.BackendURL)
	group.start("frontend", server.ListenAndServe, server.Shutdown)
	fmt.Fprintf(stdout, "frontend listening on http://%s\n", cfg.FrontendAddr)
	return group.wait(ctx, stderr)
}
