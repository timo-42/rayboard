package app

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/timo-42/rayboard/internal/backend"
	"github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/cronjobs"
	"github.com/timo-42/rayboard/internal/backend/notifications"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/store"
	"github.com/timo-42/rayboard/internal/backend/tracker"
	"github.com/timo-42/rayboard/internal/backend/webhooks"
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

func TestVerifyDocs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Main(context.Background(), []string{"verify", "docs"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "docs check passed:") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestVerifyUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Main(context.Background(), []string{"verify", "missing"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown verify command") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
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
	server := newDemoTestServer(t, db, "")
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
		!strings.Contains(output, "demo user: role=manager") ||
		!strings.Contains(output, "demo group: role=product") ||
		!strings.Contains(output, "demo group: role=automation") ||
		!strings.Contains(output, "demo project:") ||
		!strings.Contains(output, "demo ticket:") ||
		!strings.Contains(output, "demo workflow:") ||
		!strings.Contains(output, "demo board:") ||
		!strings.Contains(output, "demo sprint:") ||
		!strings.Contains(output, "demo ticket hook:") ||
		!strings.Contains(output, "demo ticket create page:") ||
		!strings.Contains(output, "demo ticket create page submission:") ||
		!strings.Contains(output, "demo comment:") ||
		!strings.Contains(output, "demo attachment:") ||
		!strings.Contains(output, "demo saved view:") ||
		!strings.Contains(output, "demo search:") ||
		!strings.Contains(output, "demo cron job:") ||
		!strings.Contains(output, "demo incoming webhook:") ||
		!strings.Contains(output, "demo outgoing webhook:") ||
		!strings.Contains(output, "demo notification destination:") ||
		!strings.Contains(output, "demo notification policy:") ||
		!strings.Contains(output, "demo notification hook:") ||
		!strings.Contains(output, "demo AI examples: skipped") {
		t.Fatalf("unexpected demo output: %s", output)
	}
	if !strings.Contains(output, "token=wh_") {
		t.Fatalf("expected incoming webhook token in demo output: %s", output)
	}
	var createPageCount int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM ticket_create_pages").Scan(&createPageCount); err != nil {
		t.Fatalf("count ticket create pages: %v", err)
	}
	if createPageCount != 1 {
		t.Fatalf("expected one demo ticket create page, got %d", createPageCount)
	}
	var intakeSubmissionCount int
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM ticket_labels
		WHERE label = 'intake-submission'
	`).Scan(&intakeSubmissionCount); err != nil {
		t.Fatalf("count intake submission labels: %v", err)
	}
	if intakeSubmissionCount != 1 {
		t.Fatalf("expected one intake submission ticket label, got %d", intakeSubmissionCount)
	}
	for table, want := range map[string]int{
		"users":                     9,
		"groups":                    6,
		"group_memberships":         11,
		"ticket_comments":           1,
		"ticket_attachments":        1,
		"saved_views":               1,
		"cron_jobs":                 1,
		"automation_runs":           0,
		"ticket_create_pages":       1,
		"webhooks":                  2,
		"notification_destinations": 1,
		"notification_policies":     1,
		"notification_hooks":        1,
	} {
		var got int
		if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&got); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if got != want {
			t.Fatalf("expected %d rows in %s, got %d", want, table, got)
		}
	}
	var projectBindingCount int
	if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM role_bindings WHERE resource_type = 'project'").Scan(&projectBindingCount); err != nil {
		t.Fatalf("count project role bindings: %v", err)
	}
	if projectBindingCount != 7 {
		t.Fatalf("expected seven project role bindings, got %d", projectBindingCount)
	}
}

func TestDemoSeedAddsAIExamplesWhenOpenRouterConfigured(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openAppTestDB(t, ctx)
	server := newDemoTestServer(t, db, "sk-or-demo")
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
	for _, marker := range []string{
		"demo AI cron job:",
		"demo AI ticket hook:",
		"demo AI webhook:",
		"demo AI outgoing webhook:",
		"demo AI notification hook:",
		"demo AI examples: provider=",
	} {
		if !strings.Contains(output, marker) {
			t.Fatalf("expected %q in demo output: %s", marker, output)
		}
	}
	for table, want := range map[string]int{
		"openrouter_providers": 1,
		"cron_jobs":            2,
		"ticket_hooks":         2,
		"webhooks":             4,
		"notification_hooks":   2,
	} {
		var got int
		if err := db.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&got); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if got != want {
			t.Fatalf("expected %d rows in %s, got %d", want, table, got)
		}
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

func newDemoTestServer(t *testing.T, db *store.DB, openRouterAPIKey string) *httptest.Server {
	t.Helper()

	ctx := context.Background()
	authorizer := authz.NewSQLEvaluator(db.SQL)
	openRouterService := openrouter.NewService(db.SQL)
	if openRouterAPIKey != "" {
		if _, err := openRouterService.CreateProvider(ctx, openrouter.CreateProviderInput{
			Name:                  "Demo AI",
			DefaultModel:          "openai/gpt-4o-mini",
			APIKey:                openRouterAPIKey,
			AllowedModels:         []string{"openai/gpt-4o-mini"},
			DefaultTimeoutSeconds: 30,
			MaxOutputTokens:       256,
			Enabled:               true,
		}); err != nil {
			t.Fatalf("create OpenRouter provider: %v", err)
		}
	}
	hooks := tracker.NewHookService(db.SQL, authorizer, tracker.WithHookOpenRouterService(openRouterService))
	trackerService := tracker.NewService(db.SQL, authorizer, tracker.WithHookService(hooks))
	createPages := tracker.NewCreatePageService(db.SQL, trackerService, authorizer)
	commentService := comments.NewService(db.SQL, authorizer)
	attachmentService := attachments.NewService(db.SQL, authorizer)
	searchService := search.NewService(db.SQL, authorizer)
	runStore := automation.NewRunStore(db.SQL)
	webhookService := webhooks.NewService(
		db.SQL,
		authorizer,
		webhooks.WithRunStore(runStore),
		webhooks.WithTrackerService(trackerService),
		webhooks.WithSearchService(searchService),
		webhooks.WithCommentService(commentService),
		webhooks.WithOpenRouterService(openRouterService),
	)
	notificationService := notifications.NewService(db.SQL, notifications.WithOpenRouterService(openRouterService))
	cronService := cronjobs.NewService(
		db.SQL,
		authorizer,
		runStore,
		cronjobs.WithTrackerService(trackerService),
		cronjobs.WithSearchService(searchService),
		cronjobs.WithCommentService(commentService),
		cronjobs.WithOpenRouterService(openRouterService),
	)
	return httptest.NewServer(backend.NewHandler(
		backend.WithAuthService(auth.NewService(db.SQL)),
		backend.WithAuditStore(audit.NewStore(db.SQL)),
		backend.WithAuthorizer(authorizer),
		backend.WithTrackerService(trackerService),
		backend.WithTicketHookService(hooks),
		backend.WithCreatePageService(createPages),
		backend.WithCommentService(commentService),
		backend.WithAttachmentService(attachmentService),
		backend.WithSearchService(searchService),
		backend.WithCronService(cronService),
		backend.WithWebhookService(webhookService),
		backend.WithNotificationService(notificationService),
		backend.WithOpenRouterService(openRouterService),
	))
}
