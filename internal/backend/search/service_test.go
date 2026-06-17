package search_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestSavedViewCRUDListAndRBAC(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedUser(t, ctx, db.SQL, "user-outsider")
	seedProject(t, ctx, db.SQL, "project-core", "CORE")
	seedProject(t, ctx, db.SQL, "project-ops", "OPS")

	seedBuiltInRole(t, ctx, db.SQL, authz.RoleGlobalAdmin)
	seedBuiltInRole(t, ctx, db.SQL, authz.RoleProjectMember)
	seedBuiltInRole(t, ctx, db.SQL, authz.RoleProjectViewer)
	seedRoleBinding(t, ctx, db.SQL, "binding-admin", "user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())
	seedRoleBinding(t, ctx, db.SQL, "binding-member-core", "user-member", authz.RoleProjectMember, authz.ProjectScope("project-core"))
	seedRoleBinding(t, ctx, db.SQL, "binding-outsider-ops", "user-outsider", authz.RoleProjectViewer, authz.ProjectScope("project-ops"))

	evaluator := authz.NewSQLEvaluator(db.SQL)
	service := search.NewService(db.SQL, evaluator, search.WithNow(fixedNow))
	admin := principal("user-admin")
	member := principal("user-member")
	outsider := principal("user-outsider")

	personal, err := service.CreateSavedView(ctx, member, search.CreateSavedViewInput{
		ScopeType: search.SavedViewScopeUser,
		ProjectID: "project-core",
		Name:      "My Bugs",
		Query: search.SavedViewQuery{
			Filter: `assignee_id == currentUser() && status != "Done" && start_date == "2026-06-16" && due_date != "2026-06-30"`,
			Text:   "login",
		},
		Sort:    []search.SortSpec{{Field: "due_date", Direction: "desc"}, {Field: "start_date", Direction: "asc"}},
		Columns: []string{"key", "title", "status", "start_date", "due_date"},
	})
	if err != nil {
		t.Fatalf("create personal saved view: %v", err)
	}
	if personal.OwnerUserID != "user-member" || personal.ScopeType != search.SavedViewScopeUser {
		t.Fatalf("unexpected personal view: %#v", personal)
	}
	if len(personal.Sort) != 2 || personal.Sort[0].Field != "due_date" || personal.Sort[1].Field != "start_date" {
		t.Fatalf("expected roadmap date sort fields, got %#v", personal.Sort)
	}
	if personal.DisplayMode != search.SavedViewDisplayList || personal.Pinned {
		t.Fatalf("unexpected personal view display metadata: %#v", personal)
	}

	if _, err := service.GetSavedView(ctx, outsider, personal.ID); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected outsider forbidden for personal view, got %v", err)
	}

	name := "My Open Bugs"
	columns := []string{"KEY", "updated_at", "START_DATE", "due_date", "labels", "key"}
	updated, err := service.UpdateSavedView(ctx, member, personal.ID, search.UpdateSavedViewInput{
		Name:    &name,
		Columns: &columns,
	})
	if err != nil {
		t.Fatalf("update personal saved view: %v", err)
	}
	if updated.Name != name || len(updated.Columns) != 5 || updated.Columns[0] != "key" || updated.Columns[1] != "updated_at" || updated.Columns[2] != "start_date" || updated.Columns[3] != "due_date" || updated.Columns[4] != "labels" {
		t.Fatalf("expected normalized columns, got %#v", updated.Columns)
	}

	_, err = service.CreateSavedView(ctx, member, search.CreateSavedViewInput{
		ScopeType: search.SavedViewScopeProject,
		ProjectID: "project-core",
		Name:      "Shared Backlog",
		Query:     search.SavedViewQuery{Filter: `status != "done"`},
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected member project view create forbidden, got %v", err)
	}

	projectView, err := service.CreateSavedView(ctx, admin, search.CreateSavedViewInput{
		ScopeType:   search.SavedViewScopeProject,
		ProjectID:   "project-core",
		Name:        "Shared Backlog",
		Query:       search.SavedViewQuery{Filter: `status != "done"`},
		DisplayMode: search.SavedViewDisplayBoard,
		GroupBy:     "status",
		Pinned:      true,
	})
	if err != nil {
		t.Fatalf("create project saved view: %v", err)
	}
	if projectView.DisplayMode != search.SavedViewDisplayBoard || projectView.GroupBy != "status" || !projectView.Pinned {
		t.Fatalf("unexpected project view metadata: %#v", projectView)
	}
	globalView, err := service.CreateSavedView(ctx, admin, search.CreateSavedViewInput{
		ScopeType: search.SavedViewScopeGlobal,
		Name:      "Recently Updated",
		Sort:      []search.SortSpec{{Field: "updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("create global saved view: %v", err)
	}
	if _, err := service.CreateSavedView(ctx, admin, search.CreateSavedViewInput{
		ScopeType: search.SavedViewScopeGlobal,
		Name:      "Bad Pinned Global",
		Pinned:    true,
	}); !errors.Is(err, search.ErrValidation) {
		t.Fatalf("expected pinned global validation, got %v", err)
	}

	memberViews, err := service.ListSavedViews(ctx, member, search.ListSavedViewsInput{ProjectID: "project-core"})
	if err != nil {
		t.Fatalf("list member views: %v", err)
	}
	assertViewIDs(t, memberViews, personal.ID, projectView.ID, globalView.ID)
	pinnedViews, err := service.ListSavedViews(ctx, member, search.ListSavedViewsInput{ProjectID: "project-core", Pinned: true})
	if err != nil {
		t.Fatalf("list pinned views: %v", err)
	}
	assertViewIDs(t, pinnedViews, projectView.ID)

	outsiderViews, err := service.ListSavedViews(ctx, outsider, search.ListSavedViewsInput{ProjectID: "project-ops"})
	if err != nil {
		t.Fatalf("list outsider views: %v", err)
	}
	assertViewIDs(t, outsiderViews, globalView.ID)

	if err := service.DeleteSavedView(ctx, outsider, personal.ID); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected outsider delete forbidden, got %v", err)
	}
	if err := service.DeleteSavedView(ctx, member, personal.ID); err != nil {
		t.Fatalf("delete personal saved view: %v", err)
	}
	if _, err := service.GetSavedView(ctx, member, personal.ID); !errors.Is(err, search.ErrNotFound) {
		t.Fatalf("expected deleted view not found, got %v", err)
	}
}

func TestSearchTicketsFTSRefreshAndRBAC(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedProject(t, ctx, db.SQL, "project-core", "CORE")
	seedProject(t, ctx, db.SQL, "project-ops", "OPS")
	seedTicket(t, ctx, db.SQL, testTicket{
		ID:          "ticket-core-1",
		ProjectID:   "project-core",
		Key:         "CORE-1",
		Title:       "Login panic",
		Description: "OAuth callback stack trace",
		Status:      "todo",
		AssigneeID:  "user-member",
		Labels:      []string{"backend", "auth"},
		StartDate:   "2026-06-15",
		DueDate:     "2026-06-16",
		UpdatedAt:   fixedNow().Add(-3 * time.Hour),
	})
	seedTicket(t, ctx, db.SQL, testTicket{
		ID:        "ticket-core-2",
		ProjectID: "project-core",
		Key:       "CORE-2",
		Title:     "Billing issue",
		Status:    "done",
		Labels:    []string{"docs"},
		StartDate: "2026-06-10",
		DueDate:   "2026-06-15",
		UpdatedAt: fixedNow().Add(-2 * time.Hour),
	})
	seedTicket(t, ctx, db.SQL, testTicket{
		ID:          "ticket-ops-1",
		ProjectID:   "project-ops",
		Key:         "OPS-1",
		Title:       "Login panic in private ops",
		Description: "Should not be visible to core member",
		Status:      "todo",
		Labels:      []string{"backend"},
		UpdatedAt:   fixedNow().Add(-1 * time.Hour),
	})
	seedCustomField(t, ctx, db.SQL, "field-severity", "project-core", "severity", "single_select")
	seedCustomField(t, ctx, db.SQL, "field-impact", "project-core", "impact", "number")
	seedCustomFieldValue(t, ctx, db.SQL, "ticket-core-1", "field-severity", `"critical"`)
	seedCustomFieldValue(t, ctx, db.SQL, "ticket-core-1", "field-impact", `8`)
	seedCustomFieldValue(t, ctx, db.SQL, "ticket-core-2", "field-severity", `"low"`)
	seedCustomFieldValue(t, ctx, db.SQL, "ticket-core-2", "field-impact", `3`)
	seedComment(t, ctx, db.SQL, "comment-core-2", "ticket-core-2", "panic also appears in a comment")
	seedAttachment(t, ctx, db.SQL, "attachment-core-2", "ticket-core-2", "runbook-panic-notes.txt", "text/plain")
	seedAttachment(t, ctx, db.SQL, "attachment-ops-1", "ticket-ops-1", "secret-ops-runbook.pdf", "application/pdf")

	var ftsRows int
	if err := db.SQL.QueryRowContext(ctx, `SELECT COUNT(*) FROM ticket_fts`).Scan(&ftsRows); err != nil {
		t.Fatalf("count initial ticket fts: %v", err)
	}
	if ftsRows != 0 {
		t.Fatalf("expected empty ticket fts before service refresh, got %d rows", ftsRows)
	}

	seedBuiltInRole(t, ctx, db.SQL, authz.RoleGlobalAdmin)
	seedBuiltInRole(t, ctx, db.SQL, authz.RoleProjectMember)
	seedRoleBinding(t, ctx, db.SQL, "binding-admin", "user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())
	seedRoleBinding(t, ctx, db.SQL, "binding-member-core", "user-member", authz.RoleProjectMember, authz.ProjectScope("project-core"))

	evaluator := authz.NewSQLEvaluator(db.SQL)
	service := search.NewService(db.SQL, evaluator, search.WithNow(fixedNow))
	member := principal("user-member")

	result, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Text: "panic",
		Sort: []search.SortSpec{{Field: "key", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("search panic: %v", err)
	}
	assertTicketIDs(t, result.Tickets, "ticket-core-1", "ticket-core-2")
	if !slices.Equal(result.Tickets[0].Labels, []string{"auth", "backend"}) || !slices.Equal(result.Tickets[1].Labels, []string{"docs"}) {
		t.Fatalf("unexpected search labels: %#v", result.Tickets)
	}

	attachmentOnly, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Text: "runbook",
		Sort: []search.SortSpec{{Field: "key", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("search attachment metadata: %v", err)
	}
	assertTicketIDs(t, attachmentOnly.Tickets, "ticket-core-2")

	attachmentContentType, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Text: "plain",
		Sort: []search.SortSpec{{Field: "key", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("search attachment content type: %v", err)
	}
	assertTicketIDs(t, attachmentContentType.Tickets, "ticket-core-2")

	hiddenAttachment, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{Text: "secret"})
	if err != nil {
		t.Fatalf("search hidden attachment metadata: %v", err)
	}
	if len(hiddenAttachment.Tickets) != 0 {
		t.Fatalf("expected RBAC to hide ops attachment match, got %#v", hiddenAttachment.Tickets)
	}

	filtered, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Filter: `project == "core" && status != "Done"`,
		Sort:   []search.SortSpec{{Field: "key", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("search filter: %v", err)
	}
	assertTicketIDs(t, filtered.Tickets, "ticket-core-1")

	backendLabel, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Filter: `labels == "Backend"`,
		Sort:   []search.SortSpec{{Field: "key", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("search label: %v", err)
	}
	assertTicketIDs(t, backendLabel.Tickets, "ticket-core-1")
	if !slices.Equal(backendLabel.Tickets[0].Labels, []string{"auth", "backend"}) {
		t.Fatalf("unexpected backend label result labels: %#v", backendLabel.Tickets[0].Labels)
	}

	labelMembership, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Filter: `"auth" in labels`,
		Sort:   []search.SortSpec{{Field: "key", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("search label membership: %v", err)
	}
	assertTicketIDs(t, labelMembership.Tickets, "ticket-core-1")

	notBackendLabel, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Filter: `labels != "backend"`,
		Sort:   []search.SortSpec{{Field: "key", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("search not label: %v", err)
	}
	assertTicketIDs(t, notBackendLabel.Tickets, "ticket-core-2")

	complexCEL, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Filter: `(status == "todo" || key == "CORE-2") && due_date <= today() && title.contains("i")`,
		Sort:   []search.SortSpec{{Field: "key", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("search complex CEL: %v", err)
	}
	assertTicketIDs(t, complexCEL.Tickets, "ticket-core-1", "ticket-core-2")

	inList, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Filter: `key in ["CORE-1", "OPS-1"]`,
		Sort:   []search.SortSpec{{Field: "key", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("search in list: %v", err)
	}
	assertTicketIDs(t, inList.Tickets, "ticket-core-1")

	customFiltered, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Filter: `custom.severity == "critical" && custom.impact >= 8`,
		Sort:   []search.SortSpec{{Field: "key", Direction: "asc"}},
	})
	if err != nil {
		t.Fatalf("search custom fields: %v", err)
	}
	assertTicketIDs(t, customFiltered.Tickets, "ticket-core-1")

	if _, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{Filter: `size(labels) > 0`}); !errors.Is(err, search.ErrValidation) {
		t.Fatalf("expected unsupported function validation, got %v", err)
	}

	if _, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{ProjectID: "project-ops", Text: "panic"}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected project search forbidden, got %v", err)
	}

	_, err = db.SQL.ExecContext(ctx, `
		UPDATE tickets
		SET title = 'Login issue', description = '', updated_at = ?
		WHERE id = 'ticket-core-1'
	`, formatTime(fixedNow()))
	if err != nil {
		t.Fatalf("update ticket text: %v", err)
	}
	_, err = db.SQL.ExecContext(ctx, `
		UPDATE ticket_comments
		SET deleted_at = ?
		WHERE id = 'comment-core-2'
	`, formatTime(fixedNow()))
	if err != nil {
		t.Fatalf("delete comment: %v", err)
	}
	_, err = db.SQL.ExecContext(ctx, `
		UPDATE ticket_attachments
		SET deleted_at = ?
		WHERE id = 'attachment-core-2'
	`, formatTime(fixedNow()))
	if err != nil {
		t.Fatalf("delete attachment: %v", err)
	}

	refreshed, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{Text: "panic"})
	if err != nil {
		t.Fatalf("search after refresh: %v", err)
	}
	if len(refreshed.Tickets) != 0 {
		t.Fatalf("expected refresh to remove stale core matches, got %#v", refreshed.Tickets)
	}
	refreshedAttachment, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{Text: "runbook"})
	if err != nil {
		t.Fatalf("search attachment after refresh: %v", err)
	}
	if len(refreshedAttachment.Tickets) != 0 {
		t.Fatalf("expected refresh to remove deleted attachment match, got %#v", refreshedAttachment.Tickets)
	}
}

func TestSearchValidationAndPagination(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-member")
	seedProject(t, ctx, db.SQL, "project-core", "CORE")
	for i, ticket := range []testTicket{
		{ID: "ticket-1", ProjectID: "project-core", Key: "CORE-1", Title: "First", Status: "todo", AssigneeID: "user-member"},
		{ID: "ticket-2", ProjectID: "project-core", Key: "CORE-2", Title: "Second", Status: "todo", AssigneeID: "user-member"},
		{ID: "ticket-3", ProjectID: "project-core", Key: "CORE-3", Title: "Third", Status: "todo", AssigneeID: "user-member"},
	} {
		ticket.UpdatedAt = fixedNow().Add(time.Duration(i) * time.Minute)
		seedTicket(t, ctx, db.SQL, ticket)
	}

	seedBuiltInRole(t, ctx, db.SQL, authz.RoleProjectMember)
	seedRoleBinding(t, ctx, db.SQL, "binding-member-core", "user-member", authz.RoleProjectMember, authz.ProjectScope("project-core"))

	evaluator := authz.NewSQLEvaluator(db.SQL)
	service := search.NewService(db.SQL, evaluator, search.WithNow(fixedNow))
	member := principal("user-member")

	firstPage, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Filter: `assignee_id == currentUser()`,
		Sort:   []search.SortSpec{{Field: "key", Direction: "asc"}},
		Limit:  2,
	})
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	assertTicketIDs(t, firstPage.Tickets, "ticket-1", "ticket-2")
	if firstPage.NextCursor == "" {
		t.Fatal("expected next cursor")
	}

	secondPage, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{
		Filter: `assignee_id == currentUser()`,
		Sort:   []search.SortSpec{{Field: "key", Direction: "asc"}},
		Limit:  2,
		Cursor: firstPage.NextCursor,
	})
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	assertTicketIDs(t, secondPage.Tickets, "ticket-3")
	if secondPage.NextCursor != "" {
		t.Fatalf("expected no second cursor, got %q", secondPage.NextCursor)
	}

	if _, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{Filter: `summary == "x"`}); !errors.Is(err, search.ErrValidation) {
		t.Fatalf("expected invalid filter validation, got %v", err)
	}
	if _, err := service.SearchTickets(ctx, member, search.SearchTicketsInput{Text: "!!!"}); !errors.Is(err, search.ErrValidation) {
		t.Fatalf("expected invalid text validation, got %v", err)
	}
}

type testTicket struct {
	ID          string
	ProjectID   string
	Key         string
	Title       string
	Description string
	Status      string
	AssigneeID  string
	Labels      []string
	StartDate   string
	DueDate     string
	UpdatedAt   time.Time
}

func openMigratedDB(t *testing.T, ctx context.Context) *store.DB {
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

func seedProject(t *testing.T, ctx context.Context, db *sql.DB, projectID string, key string) {
	t.Helper()

	now := formatTime(fixedNow())
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (id, key, name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, projectID, key, key+" Project", now, now)
	if err != nil {
		t.Fatalf("seed project %s: %v", projectID, err)
	}
}

func seedBuiltInRole(t *testing.T, ctx context.Context, db *sql.DB, roleName authz.RoleName) {
	t.Helper()

	role, ok := authz.BuiltInRole(roleName)
	if !ok {
		t.Fatalf("unknown built-in role %s", roleName)
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO roles (id, name, description)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO NOTHING
	`, string(role.Name), string(role.Name), "Built-in test role")
	if err != nil {
		t.Fatalf("seed role %s: %v", roleName, err)
	}
	for _, permission := range role.Permissions {
		_, err := db.ExecContext(ctx, `
			INSERT INTO role_permissions (role_id, permission)
			VALUES (?, ?)
			ON CONFLICT(role_id, permission) DO NOTHING
		`, string(role.Name), string(permission))
		if err != nil {
			t.Fatalf("seed role permission %s/%s: %v", roleName, permission, err)
		}
	}
}

func seedRoleBinding(t *testing.T, ctx context.Context, db *sql.DB, id string, userID string, roleName authz.RoleName, scope authz.Scope) {
	t.Helper()

	var resourceType any
	var resourceID any
	switch scope.Kind {
	case authz.ScopeKindGlobal:
		resourceType = nil
		resourceID = nil
	case authz.ScopeKindProject:
		resourceType = string(authz.ScopeKindProject)
		resourceID = scope.ProjectID
	default:
		t.Fatalf("unsupported scope %#v", scope)
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO role_bindings (
			id, role_id, subject_type, subject_id, resource_type, resource_id, created_at
		)
		VALUES (?, ?, 'user', ?, ?, ?, ?)
	`, id, string(roleName), userID, resourceType, resourceID, formatTime(fixedNow()))
	if err != nil {
		t.Fatalf("seed role binding %s: %v", id, err)
	}
}

func seedTicket(t *testing.T, ctx context.Context, db *sql.DB, ticket testTicket) {
	t.Helper()

	if ticket.Status == "" {
		ticket.Status = "todo"
	}
	if ticket.UpdatedAt.IsZero() {
		ticket.UpdatedAt = fixedNow()
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO tickets (
			id, project_id, key, title, description, status, assignee_id,
			start_date, due_date, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ticket.ID, ticket.ProjectID, ticket.Key, ticket.Title, ticket.Description, ticket.Status, nullableString(ticket.AssigneeID), nullableString(ticket.StartDate), nullableString(ticket.DueDate), formatTime(fixedNow()), formatTime(ticket.UpdatedAt))
	if err != nil {
		t.Fatalf("seed ticket %s: %v", ticket.ID, err)
	}
	for _, label := range ticket.Labels {
		_, err := db.ExecContext(ctx, `
			INSERT INTO ticket_labels (ticket_id, label, created_at)
			VALUES (?, ?, ?)
		`, ticket.ID, label, formatTime(fixedNow()))
		if err != nil {
			t.Fatalf("seed ticket label %s/%s: %v", ticket.ID, label, err)
		}
	}
}

func seedCustomField(t *testing.T, ctx context.Context, db *sql.DB, fieldID string, projectID string, key string, fieldType string) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO custom_field_definitions (
			id, project_id, key, name, field_type, required, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, 0, ?, ?)
	`, fieldID, projectID, key, key, fieldType, formatTime(fixedNow()), formatTime(fixedNow()))
	if err != nil {
		t.Fatalf("seed custom field %s: %v", fieldID, err)
	}
}

func seedCustomFieldValue(t *testing.T, ctx context.Context, db *sql.DB, ticketID string, fieldID string, valueJSON string) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO ticket_custom_field_values (ticket_id, field_id, value_json, updated_at)
		VALUES (?, ?, ?, ?)
	`, ticketID, fieldID, valueJSON, formatTime(fixedNow()))
	if err != nil {
		t.Fatalf("seed custom field value %s/%s: %v", ticketID, fieldID, err)
	}
}

func seedComment(t *testing.T, ctx context.Context, db *sql.DB, commentID string, ticketID string, body string) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO ticket_comments (id, ticket_id, body, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, commentID, ticketID, body, formatTime(fixedNow()), formatTime(fixedNow()))
	if err != nil {
		t.Fatalf("seed comment %s: %v", commentID, err)
	}
}

func seedAttachment(t *testing.T, ctx context.Context, db *sql.DB, attachmentID string, ticketID string, fileName string, contentType string) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO ticket_attachments (
			id, ticket_id, file_name, content_type, size_bytes, data, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, attachmentID, ticketID, fileName, contentType, 4, []byte("test"), formatTime(fixedNow()))
	if err != nil {
		t.Fatalf("seed attachment %s: %v", attachmentID, err)
	}
}

func assertViewIDs(t *testing.T, views []search.SavedView, want ...string) {
	t.Helper()

	got := make(map[string]struct{}, len(views))
	for _, view := range views {
		got[view.ID] = struct{}{}
	}
	for _, id := range want {
		if _, ok := got[id]; !ok {
			t.Fatalf("expected view %s in %#v", id, views)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d views, got %#v", len(want), views)
	}
}

func assertTicketIDs(t *testing.T, tickets []search.Ticket, want ...string) {
	t.Helper()

	if len(tickets) != len(want) {
		t.Fatalf("expected tickets %v, got %#v", want, tickets)
	}
	for i, ticket := range tickets {
		if ticket.ID != want[i] {
			t.Fatalf("expected ticket %d to be %s, got %#v", i, want[i], tickets)
		}
	}
}

func principal(userID string) authz.Principal {
	return authz.Principal{
		UserID:      userID,
		ActorUserID: userID,
		AuthKind:    authz.AuthKindSession,
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
