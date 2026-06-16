package automation

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestRunLifecycle(t *testing.T) {
	ctx := context.Background()
	db := openAutomationTestDB(t, ctx)
	seedAutomationProject(t, ctx, db)
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	runStore := NewRunStore(db.SQL, WithNow(func() time.Time { return now }))

	started, err := runStore.Start(ctx, StartInput{
		TriggerType: "cron",
		TriggerRef:  "job-1",
		ProjectID:   "project-1",
		Engine:      "lua",
		ActorUserID: "user-1",
		Input:       map[string]any{"ticket_count": float64(3)},
		Limits:      map[string]any{"timeout_ms": float64(1000)},
	})
	if err != nil {
		t.Fatalf("start run: %v", err)
	}
	if started.ID == "" || started.Status != StatusRunning || started.StartedAt == nil {
		t.Fatalf("unexpected started run: %#v", started)
	}
	if started.Input["engine"] != "lua" || started.Input["actor_user_id"] != "user-1" {
		t.Fatalf("unexpected input envelope: %#v", started.Input)
	}

	now = now.Add(2 * time.Second)
	finished, err := runStore.Finish(ctx, started.ID, FinishInput{
		Status: StatusSucceeded,
		Output: map[string]any{
			"actions": []any{"commented"},
		},
		Logs: []string{"processed"},
	})
	if err != nil {
		t.Fatalf("finish run: %v", err)
	}
	if finished.Status != StatusSucceeded || finished.FinishedAt == nil || finished.Error != "" {
		t.Fatalf("unexpected finished run: %#v", finished)
	}
	if _, ok := finished.Output["output"].(map[string]any); !ok {
		t.Fatalf("expected output envelope, got %#v", finished.Output)
	}

	got, err := runStore.Get(ctx, started.ID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if got.ID != started.ID || got.TriggerRef != "job-1" {
		t.Fatalf("unexpected fetched run: %#v", got)
	}

	listed, err := runStore.List(ctx, ListInput{TriggerType: "cron", ProjectID: "project-1"})
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != started.ID {
		t.Fatalf("unexpected listed runs: %#v", listed)
	}
}

func TestRunStoreValidationAndNotFound(t *testing.T) {
	ctx := context.Background()
	db := openAutomationTestDB(t, ctx)
	runStore := NewRunStore(db.SQL)

	if _, err := runStore.Start(ctx, StartInput{}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected start validation error, got %v", err)
	}
	if _, err := runStore.Finish(ctx, "missing", FinishInput{Status: StatusSucceeded}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing finish not found, got %v", err)
	}
	started, err := runStore.Start(ctx, StartInput{TriggerType: "ticket_hook"})
	if err != nil {
		t.Fatalf("start run: %v", err)
	}
	if _, err := runStore.Finish(ctx, started.ID, FinishInput{Status: StatusRunning}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid final status validation error, got %v", err)
	}
	if _, err := runStore.List(ctx, ListInput{Limit: 201}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected list validation error, got %v", err)
	}
}

func openAutomationTestDB(t *testing.T, ctx context.Context) *store.DB {
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

func seedAutomationProject(t *testing.T, ctx context.Context, db *store.DB) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES ('project-1', 'AUTO', 'Automation')
	`); err != nil {
		t.Fatalf("seed automation project: %v", err)
	}
}
