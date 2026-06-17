package cronjobs

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
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

	openRouterService := openrouter.NewService(db.SQL)
	disabledProvider, err := openRouterService.CreateProvider(ctx, openrouter.CreateProviderInput{
		Name:         "Disabled",
		DefaultModel: "openai/gpt-4.1-mini",
		APIKey:       "sk-or-secret",
		Enabled:      false,
	})
	if err != nil {
		t.Fatalf("create disabled provider: %v", err)
	}
	aiService := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-1", authz.RoleGlobalAdmin, authz.GlobalScope()))),
		automation.NewRunStore(db.SQL),
		WithOpenRouterService(openRouterService),
	)
	if _, err := aiService.Create(ctx, principal, CreateInput{
		Name:     "Disabled AI",
		Schedule: "* * * * *",
		Timezone: "UTC",
		Engine: EngineSpec{
			Type:       EngineAI,
			Prompt:     "Return JSON",
			ProviderID: disabledProvider.ID,
		},
	}); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected disabled provider validation, got %v", err)
	}
}

func TestCronJobAIManualRun(t *testing.T) {
	ctx := context.Background()
	db := openCronTestDB(t, ctx)
	seedCronUser(t, ctx, db, "user-1", false)

	var receivedAuth string
	var receivedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			t.Fatalf("decode OpenRouter request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "gen_cron",
			"choices": [{"message": {"role": "assistant", "content": "{\"summary\":\"done\",\"tickets\":3}"}}],
			"usage": {"prompt_tokens": 7, "completion_tokens": 4}
		}`))
	}))
	defer server.Close()

	openRouterService := openrouter.NewService(db.SQL, openrouter.WithBaseURL(server.URL))
	provider, err := openRouterService.CreateProvider(ctx, openrouter.CreateProviderInput{
		Name:                  "Default",
		DefaultModel:          "openai/gpt-4.1-mini",
		APIKey:                "sk-or-secret",
		DefaultTimeoutSeconds: 15,
		MaxOutputTokens:       321,
		Enabled:               true,
	})
	if err != nil {
		t.Fatalf("create OpenRouter provider: %v", err)
	}

	service := NewService(
		db.SQL,
		authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-1", authz.RoleGlobalAdmin, authz.GlobalScope()))),
		automation.NewRunStore(db.SQL),
		WithOpenRouterService(openRouterService),
	)
	principal := authz.Principal{UserID: "user-1", AuthKind: authz.AuthKindSession}

	job, err := service.Create(ctx, principal, CreateInput{
		Name:     "AI triage",
		Schedule: "0 9 * * *",
		Timezone: "UTC",
		Engine: EngineSpec{
			Type:       EngineAI,
			Prompt:     "Summarize the current backlog as JSON.",
			ProviderID: provider.ID,
		},
	})
	if err != nil {
		t.Fatalf("create AI cron job: %v", err)
	}

	run, err := service.RunNow(ctx, principal, job.ID)
	if err != nil {
		t.Fatalf("run AI cron job: %v", err)
	}
	if receivedAuth != "Bearer sk-or-secret" {
		t.Fatalf("unexpected OpenRouter auth header: %q", receivedAuth)
	}
	if receivedBody["model"] != "openai/gpt-4.1-mini" || receivedBody["max_tokens"] != float64(321) {
		t.Fatalf("unexpected OpenRouter request body: %#v", receivedBody)
	}
	if run.Status != automation.StatusSucceeded {
		t.Fatalf("unexpected AI run: %#v", run)
	}
	envelope, ok := run.Output["output"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected AI output: %#v", run.Output)
	}
	output, ok := envelope["output"].(map[string]any)
	if !ok || output["summary"] != "done" || output["tickets"] != float64(3) {
		t.Fatalf("unexpected AI output: %#v", run.Output)
	}
	if envelope["model"] != "openai/gpt-4.1-mini" || envelope["provider_id"] != provider.ID {
		t.Fatalf("expected model/provider metadata, got %#v", run.Output)
	}
	encoded, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("marshal run: %v", err)
	}
	if strings.Contains(string(encoded), "sk-or-secret") || strings.Contains(string(encoded), "Summarize the current backlog") {
		t.Fatalf("run history leaked secret or prompt: %s", string(encoded))
	}
}

func TestCronJobAIActionsActAsOwner(t *testing.T) {
	ctx := context.Background()
	db := openCronTestDB(t, ctx)
	seedCronUser(t, ctx, db, "owner", false)
	seedCronUser(t, ctx, db, "admin", false)
	seedCronProject(t, ctx, db, "project-1")

	authorizer := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
		authz.UserBinding("owner", authz.RoleProjectMember, authz.ProjectScope("project-1")),
	))
	trackerService := tracker.NewService(db.SQL, authorizer)
	searchService := search.NewService(db.SQL, authorizer)
	commentService := comments.NewService(db.SQL, authorizer)
	existing, err := trackerService.CreateTicket(ctx, authz.Principal{UserID: "owner", AuthKind: authz.AuthKindSession}, tracker.CreateTicketInput{
		ProjectID:   "project-1",
		Title:       "Existing ticket",
		Description: "Before AI cron",
	})
	if err != nil {
		t.Fatalf("seed existing ticket: %v", err)
	}

	var receivedPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode OpenRouter request: %v", err)
		}
		if len(body.Messages) > 1 {
			receivedPrompt = body.Messages[1].Content
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "gen_cron_actions",
			"choices": []map[string]any{{
				"message": map[string]any{
					"role": "assistant",
					"content": `{
						"summary": "planned",
						"actions": [
							{"type":"create_ticket","input":{"project_id":"project-1","title":"AI-created cron ticket","description":"Created by AI cron","labels":["ai","cron"]}},
							{"type":"update_ticket","input":{"ticket_id":"` + existing.ID + `","priority":"High","labels":["cron-updated"]}},
							{"type":"comment","input":{"ticket_id":"` + existing.ID + `","body":"AI cron comment"}},
							{"type":"search","input":{"project_id":"project-1","filter":"labels == \"cron-updated\"","limit":10}}
						]
					}`,
				},
			}},
			"usage": map[string]any{"prompt_tokens": 13, "completion_tokens": 9},
		})
	}))
	defer server.Close()

	openRouterService := openrouter.NewService(db.SQL, openrouter.WithBaseURL(server.URL))
	provider, err := openRouterService.CreateProvider(ctx, openrouter.CreateProviderInput{
		Name:         "Default",
		DefaultModel: "openai/gpt-4.1-mini",
		APIKey:       "sk-or-secret",
		Enabled:      true,
	})
	if err != nil {
		t.Fatalf("create OpenRouter provider: %v", err)
	}
	service := NewService(
		db.SQL,
		authorizer,
		automation.NewRunStore(db.SQL),
		WithTrackerService(trackerService),
		WithSearchService(searchService),
		WithCommentService(commentService),
		WithOpenRouterService(openRouterService),
	)
	principal := authz.Principal{UserID: "admin", AuthKind: authz.AuthKindSession}

	job, err := service.Create(ctx, principal, CreateInput{
		OwnerUserID: "owner",
		ProjectID:   "project-1",
		Name:        "AI action triage",
		Schedule:    "0 9 * * *",
		Timezone:    "UTC",
		Engine: EngineSpec{
			Type:       EngineAI,
			Prompt:     "Plan and execute cron triage actions.",
			ProviderID: provider.ID,
		},
	})
	if err != nil {
		t.Fatalf("create AI cron job: %v", err)
	}

	run, err := service.RunNow(ctx, principal, job.ID)
	if err != nil {
		t.Fatalf("run AI cron job: %v", err)
	}
	if run.Status != automation.StatusSucceeded {
		t.Fatalf("unexpected AI run: %#v", run)
	}
	if !strings.Contains(receivedPrompt, "Plan and execute cron triage actions.") || !strings.Contains(receivedPrompt, `"surface":"cron"`) || !strings.Contains(receivedPrompt, "Allowed action types") {
		t.Fatalf("unexpected AI prompt: %s", receivedPrompt)
	}
	if strings.Contains(receivedPrompt, "sk-or-secret") {
		t.Fatalf("AI prompt leaked OpenRouter secret: %s", receivedPrompt)
	}
	if countRows(t, ctx, db, "tickets") != 2 {
		t.Fatalf("expected existing plus AI-created ticket")
	}
	if countRows(t, ctx, db, "ticket_comments") != 1 {
		t.Fatalf("expected AI-created comment")
	}
	var reporterID string
	if err := db.SQL.QueryRowContext(ctx, "SELECT reporter_id FROM tickets WHERE title = ?", "AI-created cron ticket").Scan(&reporterID); err != nil {
		t.Fatalf("query AI-created ticket reporter: %v", err)
	}
	if reporterID != "owner" {
		t.Fatalf("expected AI action to run as owner, got reporter %q", reporterID)
	}
	envelope, ok := run.Output["output"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected AI output: %#v", run.Output)
	}
	output, ok := envelope["output"].(map[string]any)
	if !ok || output["summary"] != "planned" {
		t.Fatalf("unexpected AI output body: %#v", run.Output)
	}
	results, ok := output["action_results"].([]any)
	if !ok || len(results) != 4 {
		t.Fatalf("unexpected AI action results: %#v", run.Output)
	}
	encoded, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("marshal run: %v", err)
	}
	if strings.Contains(string(encoded), "sk-or-secret") || strings.Contains(string(encoded), "Plan and execute cron triage actions") {
		t.Fatalf("run history leaked secret or prompt: %s", string(encoded))
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
