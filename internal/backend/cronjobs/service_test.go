package cronjobs

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/store"
	"github.com/timo-42/rayboard/internal/backend/tracker"
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
		Name:     "Daily triage",
		Schedule: "0 9 * * *",
		Timezone: "UTC",
		Engine: EngineSpec{
			Type:   EngineLua,
			Script: `rayboard.log("triage started")`,
		},
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
		Engine: EngineSpec{
			Type:   EngineLua,
			Script: `rayboard.log("nope")`,
		},
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected disabled owner validation, got %v", err)
	}

	if _, err := service.Create(ctx, principal, CreateInput{
		Name:     "AI",
		Schedule: "* * * * *",
		Timezone: "UTC",
		Engine: EngineSpec{
			Type:       EngineAI,
			Prompt:     "Return JSON",
			ProviderID: "openrouter-default",
		},
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected AI validation, got %v", err)
	}
}

func TestCronJobLuaRayboardHelpers(t *testing.T) {
	ctx := context.Background()
	db := openCronTestDB(t, ctx)
	seedCronUser(t, ctx, db, "user-1", false)
	seedCronProject(t, ctx, db, "project-1")

	authorizer := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-1", authz.RoleGlobalAdmin, authz.GlobalScope())))
	trackerService := tracker.NewService(db.SQL, authorizer)
	searchService := search.NewService(db.SQL, authorizer)
	commentService := comments.NewService(db.SQL, authorizer)
	service := NewService(
		db.SQL,
		authorizer,
		automation.NewRunStore(db.SQL),
		WithTrackerService(trackerService),
		WithSearchService(searchService),
		WithCommentService(commentService),
	)
	principal := authz.Principal{UserID: "user-1", AuthKind: authz.AuthKindSession}

	job, err := service.Create(ctx, principal, CreateInput{
		Name:     "Lua helpers",
		Schedule: "* * * * *",
		Engine: EngineSpec{
			Type: EngineLua,
			Script: `
local ticket, err = rayboard.create_ticket({
  project_id = "project-1",
  title = "Lua-created ticket",
  description = "Created from a cron script",
  labels = {"Backend", "Lua"}
})
if err then error(err.message) end

local comment, comment_err = rayboard.comment({
  ticket_id = ticket.id,
  body = "Lua helper comment"
})
if comment_err then error(comment_err.message) end

local updated, update_err = rayboard.update_ticket({
  ticket_id = ticket.id,
  priority = "High",
  labels = {"automation", "Backend"}
})
if update_err then error(update_err.message) end

local fetched, get_err = rayboard.get_ticket({ ticket_id = updated.id })
if get_err then error(get_err.message) end

local results, search_err = rayboard.search({
  project_id = "project-1",
  filter = 'labels == "automation"',
  limit = 10
})
if search_err then error(search_err.message) end

rayboard.log(fetched.key .. ":" .. tostring(#fetched.labels) .. ":" .. fetched.labels[1] .. ":" .. tostring(#results.items))
`,
		},
	})
	if err != nil {
		t.Fatalf("create cron job: %v", err)
	}

	run, err := service.RunNow(ctx, principal, job.ID)
	if err != nil {
		t.Fatalf("run cron job: %v", err)
	}
	if run.Status != automation.StatusSucceeded {
		t.Fatalf("unexpected run status: %#v", run)
	}
	if countRows(t, ctx, db, "tickets") != 1 {
		t.Fatalf("expected one ticket")
	}
	if countRows(t, ctx, db, "ticket_comments") != 1 {
		t.Fatalf("expected one comment")
	}
	logs, ok := run.Output["logs"].([]any)
	if !ok || len(logs) != 1 || logs[0] != "AUTO-1:2:automation:1" {
		t.Fatalf("unexpected helper logs: %#v", run.Output)
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

func seedCronProject(t *testing.T, ctx context.Context, db *store.DB, id string) {
	t.Helper()

	if _, err := db.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, key, name)
		VALUES (?, 'AUTO', 'Automation')
	`, id); err != nil {
		t.Fatalf("seed project %s: %v", id, err)
	}
}

func countRows(t *testing.T, ctx context.Context, db *store.DB, table string) int {
	t.Helper()

	var count int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return count
}
