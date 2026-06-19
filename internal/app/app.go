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
	"github.com/timo-42/rayboard/internal/docscheck"
	"github.com/timo-42/rayboard/internal/releasecheck"
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
	case "verify":
		return runVerify(args[1:], stdout, stderr)
	case "--help", "help":
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
	configureLongFlagUsage(flags, stderr, fmt.Sprintf("usage: rayboard %s [flags]", mode))
	if err := rejectSingleDashFlags(flags, args); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if flagHelpRequested(args) {
		configureLongFlagUsage(flags, stdout, fmt.Sprintf("usage: rayboard %s [flags]", mode))
		flags.Usage()
		return 0
	}

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

func runVerify(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: rayboard verify <docs|release>")
		return 2
	}
	switch args[0] {
	case "docs":
		report := docscheck.Check()
		if len(report.Errors) != 0 {
			for _, err := range report.Errors {
				fmt.Fprintf(stderr, "docs check: %s\n", err)
			}
			return 1
		}
		fmt.Fprintf(stdout, "docs check passed: files=%d local_links=%d\n", report.FilesChecked, report.LinksChecked)
		return 0
	case "release":
		report := releasecheck.Check(".")
		if len(report.Errors) != 0 {
			for _, err := range report.Errors {
				fmt.Fprintf(stderr, "release check: %s\n", err)
			}
			return 1
		}
		fmt.Fprintf(stdout, "release check passed: docs_files=%d docs_links=%d repo_files=%d\n", report.DocsFilesChecked, report.DocsLinksChecked, report.FilesChecked)
		return 0
	case "-h", "--help", "help":
		fmt.Fprintln(stdout, "usage: rayboard verify <docs|release>")
		return 0
	default:
		fmt.Fprintf(stderr, "unknown verify command %q\n\n", args[0])
		fmt.Fprintln(stderr, "usage: rayboard verify <docs|release>")
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `usage: rayboard <command> [flags]

commands:
  combined   run frontend and backend in one process
  backend    run only the backend API
  frontend   run only the frontend server
  demo seed  populate a running backend with demo data
  verify docs
             check embedded documentation links and required references
  verify release
             check docs plus release workflow and cross-build artifact wiring`)
}
