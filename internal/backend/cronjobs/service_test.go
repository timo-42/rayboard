package cronjobs

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestCronJobLifecycleAndManualRun(t *testing.T) {
	ctx := context.Background()
	db := openCronTestDB(t, ctx)
	seedCronUser(t, ctx, db, "user-1", false)

	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-1", authz.RoleGlobalAdmin, authz.GlobalScope()))),
		automation.NewRunStore(db.SQL, automation.WithNow(func() time.Time { return now })),
		WithNow(func() time.Time { return now }),
	)
	principal := authz.Principal{UserID: "user-1", AuthKind: authz.AuthKindSession}

	job, err := service.Create(ctx, principal, CreateInput{
		Name:      "Daily triage",
		Schedule:  "0 9 * * *",
		Timezone:  "UTC",
		Engine:    EngineLua,
		LuaSource: `rayboard.log("triage started")`,
	})
	if err != nil {
		t.Fatalf("create cron job: %v", err)
	}
	if job.ID == "" || job.OwnerUserID != "user-1" || job.NextRunAt == nil {
		t.Fatalf("unexpected created job: %#v", job)
	}

	got, err := service.Get(ctx, principal, job.ID)
	if err != nil {
		t.Fatalf("get cron job: %v", err)
	}
	if got.Name != "Daily triage" {
		t.Fatalf("unexpected fetched job: %#v", got)
	}

	enabled := true
	name := "Morning triage"
	got, err = service.Update(ctx, principal, job.ID, UpdateInput{
		Name:    &name,
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("update cron job: %v", err)
	}
	if got.Name != name || !got.Enabled {
		t.Fatalf("unexpected updated job: %#v", got)
	}

	now = now.Add(time.Minute)
	run, err := service.RunNow(ctx, principal, job.ID)
	if err != nil {
		t.Fatalf("run cron job: %v", err)
	}
	if run.Status != automation.StatusSucceeded || run.TriggerRef != job.ID {
		t.Fatalf("unexpected run: %#v", run)
	}
	logs, ok := run.Output["logs"].([]any)
	if !ok || len(logs) != 1 || logs[0] != "triage started" {
		t.Fatalf("unexpected run logs: %#v", run.Output)
	}

	runs, err := service.ListRuns(ctx, principal, job.ID, 0, 0)
	if err != nil {
		t.Fatalf("list cron runs: %v", err)
	}
	if len(runs) != 1 || runs[0].ID != run.ID {
		t.Fatalf("unexpected listed runs: %#v", runs)
	}

	if err := service.Delete(ctx, principal, job.ID); err != nil {
		t.Fatalf("delete cron job: %v", err)
	}
	if _, err := service.Get(ctx, principal, job.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestCronJobValidationAndDisabledOwner(t *testing.T) {
	ctx := context.Background()
	db := openCronTestDB(t, ctx)
	seedCronUser(t, ctx, db, "user-1", false)
	seedCronUser(t, ctx, db, "disabled-user", true)

	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-1", authz.RoleGlobalAdmin, authz.GlobalScope()))),
		automation.NewRunStore(db.SQL),
	)
	principal := authz.Principal{UserID: "user-1", AuthKind: authz.AuthKindSession}

	if _, err := service.Create(ctx, principal, CreateInput{
		OwnerUserID: "disabled-user",
		Name:        "Disabled",
		Schedule:    "* * * * *",
		Timezone:    "UTC",
		Engine:      EngineLua,
		LuaSource:   `rayboard.log("nope")`,
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected disabled owner validation, got %v", err)
	}

	if _, err := service.Create(ctx, principal, CreateInput{
		Name:      "AI",
		Schedule:  "* * * * *",
		Timezone:  "UTC",
		Engine:    EngineAI,
		AIPrompt:  "Return JSON",
		LuaSource: "",
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected AI validation, got %v", err)
	}
}

func openCronTestDB(t *testing.T, ctx context.Context) *store.DB {
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

func seedCronUser(t *testing.T, ctx context.Context, db *store.DB, id string, disabled bool) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO users (id, username, display_name, is_disabled)
		VALUES (?, ?, ?, ?)
	`, id, id, id, disabled); err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}
