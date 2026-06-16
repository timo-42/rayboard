package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os/signal"
	"syscall"

	"github.com/timo-42/rayboard/internal/config"
	"github.com/timo-42/rayboard/internal/runtime"
)

// Main is split from cmd/rayboard so command behavior can be tested.
func Main(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch args[0] {
	case "combined":
		return runRuntime(ctx, runtime.ModeCombined, args[1:], stdout, stderr)
	case "backend":
		return runRuntime(ctx, runtime.ModeBackend, args[1:], stdout, stderr)
	case "frontend":
		return runRuntime(ctx, runtime.ModeFrontend, args[1:], stdout, stderr)
	case "demo":
		return runDemo(ctx, args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runRuntime(ctx context.Context, mode runtime.Mode, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet(string(mode), flag.ContinueOnError)
	flags.SetOutput(stderr)

	cfg := config.FromEnv()
	cfg.BindRuntimeFlags(flags)

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if err := runtime.Run(ctx, mode, cfg, stdout, stderr); err != nil {
		if errors.Is(err, context.Canceled) {
			return 0
		}
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	return 0
}

func runDemo(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return runDemoSeed(ctx, args, stdout, stderr)
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `usage: rayboard <command> [flags]

commands:
  combined   run frontend and backend in one process
  backend    run only the backend API
  frontend   run only the frontend server
  demo seed  populate a running backend with demo data`)
}
