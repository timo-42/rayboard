package app

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/timo-42/rayboard/internal/backend"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/store"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestMainUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Main(context.Background(), []string{"missing"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected usage output on stderr")
	}
}

func TestDemoSeedRequiresFreshReset(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Main(context.Background(), []string{
		"demo", "seed",
		"--backend-url", "http://127.0.0.1:8081",
		"--admin-user", "admin",
		"--admin-password", "secret",
	}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
}

func TestDemoSeedPopulatesBackend(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openAppTestDB(t, ctx)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	hooks := tracker.NewHookService(db.SQL, authorizer)
	trackerService := tracker.NewService(db.SQL, authorizer, tracker.WithHookService(hooks))
	server := httptest.NewServer(backend.NewHandler(
		backend.WithAuthService(auth.NewService(db.SQL)),
		backend.WithAuthorizer(authorizer),
		backend.WithTrackerService(trackerService),
		backend.WithTicketHookService(hooks),
	))
	t.Cleanup(server.Close)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Main(ctx, []string{
		"demo", "seed",
		"--backend-url", server.URL,
		"--admin-user", bootstrap.Username,
		"--admin-password", bootstrap.Password,
		"--fresh-reset",
	}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "demo user: role=lead") ||
		!strings.Contains(output, "demo project:") ||
		!strings.Contains(output, "demo ticket:") ||
		!strings.Contains(output, "demo workflow:") ||
		!strings.Contains(output, "demo board:") ||
		!strings.Contains(output, "demo sprint:") ||
		!strings.Contains(output, "demo ticket hook:") {
		t.Fatalf("unexpected demo output: %s", output)
	}
}

func openAppTestDB(t *testing.T, ctx context.Context) (*store.DB, auth.BootstrapAdminResult) {
	t.Helper()

	db, err := store.Open(ctx, t.TempDir()+"/rayboard.sqlite")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	bootstrap, err := auth.BootstrapAdmin(ctx, db.SQL)
	if err != nil {
		t.Fatalf("bootstrap admin: %v", err)
	}
	return db, bootstrap
}
