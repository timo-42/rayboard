package engines_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/engines"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestLuaEngineTestReturnsOutputAndLogs(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))

	run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
		Surface: "ticket_hook_before",
		Context: map[string]any{"ticket_id": "ticket-1", "dry_run": false},
		Input:   map[string]any{"title": "Example"},
		Engine: engines.EngineSpec{
			Type: "lua",
			Script: `
rayboard.log("checking " .. input.title)
return { ok = true, title = input.title, surface = context.surface, ticket_id = context.ticket_id, dry_run = context.dry_run }
`,
		},
	})
	if err != nil {
		t.Fatalf("test lua engine: %v", err)
	}
	if run.Status != automation.StatusSucceeded {
		t.Fatalf("expected succeeded run, got %#v", run)
	}
	output, _ := run.Output["output"].(map[string]any)
	if output["ok"] != true || output["title"] != "Example" || output["surface"] != "ticket_hook_before" || output["ticket_id"] != "ticket-1" || output["dry_run"] != true {
		t.Fatalf("unexpected output: %#v", run.Output)
	}
	logs, _ := run.Output["logs"].([]any)
	if !slices.Equal(anyStrings(logs), []string{"checking Example"}) {
		t.Fatalf("unexpected logs: %#v", run.Output)
	}
	if encoded := run.Input; encoded["engine"] != "lua" {
		t.Fatalf("expected run input to store only engine type, got %#v", encoded)
	}
	inputEnvelope, _ := run.Input["input"].(map[string]any)
	contextEnvelope, _ := inputEnvelope["context"].(map[string]any)
	if inputEnvelope["dry_run"] != true || contextEnvelope["ticket_id"] != "ticket-1" || contextEnvelope["dry_run"] != true {
		t.Fatalf("expected normalized dry-run context in run input, got %#v", run.Input)
	}
}

func TestLuaEngineTestDefaultsToScratchSurface(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))

	run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
		Engine: engines.EngineSpec{
			Type:   "lua",
			Script: `return { surface = context.surface, dry_run = context.dry_run }`,
		},
	})
	if err != nil {
		t.Fatalf("test scratch lua engine: %v", err)
	}
	output, _ := run.Output["output"].(map[string]any)
	if output["surface"] != "scratch" || output["dry_run"] != true {
		t.Fatalf("expected scratch dry-run output, got %#v", run.Output)
	}
	inputEnvelope, _ := run.Input["input"].(map[string]any)
	contextEnvelope, _ := inputEnvelope["context"].(map[string]any)
	if inputEnvelope["dry_run"] != true || contextEnvelope["surface"] != "scratch" {
		t.Fatalf("expected scratch run input context, got %#v", run.Input)
	}
}

func TestLuaEngineTestRecordsFailureAsRunStatus(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))

	run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
		Engine: engines.EngineSpec{Type: "lua", Script: `error("boom")`},
	})
	if err == nil {
		t.Fatal("expected runtime error")
	}
	if run.ID == "" || run.Status != automation.StatusFailed || run.Error == "" {
		t.Fatalf("expected failed run with error, got run=%#v err=%v", run, err)
	}
}

func TestEngineTestRequiresAutomationPermission(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-member")
	evaluator := authz.NewInMemoryEvaluator()
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL))

	_, err := service.Test(ctx, principal("user-member"), engines.TestInput{
		Engine: engines.EngineSpec{Type: "lua", Script: `return { ok = true }`},
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func openTestDB(t *testing.T, ctx context.Context) *store.DB {
	t.Helper()

	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "rayboard.sqlite"))
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
	return db
}

func seedUser(t *testing.T, ctx context.Context, db *sql.DB, userID string) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, username, display_name)
		VALUES (?, ?, ?)
	`, userID, userID, userID)
	if err != nil {
		t.Fatalf("seed user %s: %v", userID, err)
	}
}

func principal(userID string) authz.Principal {
	return authz.Principal{UserID: userID, ActorUserID: userID, AuthKind: authz.AuthKindSession}
}

func fixedNow() time.Time {
	return time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
}

func anyStrings(values []any) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		text, _ := value.(string)
		out = append(out, text)
	}
	return out
}
