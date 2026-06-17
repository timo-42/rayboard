package tracker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestCreatePageLifecycleAndSubmit(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)
	seedUser(t, ctx, db.SQL, "user-viewer")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	trackerService := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	pageService := tracker.NewCreatePageService(db.SQL, trackerService, evaluator)
	admin := principal("user-admin")
	project, err := trackerService.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "FORM", Name: "Forms"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	evaluator.BindRole(authz.UserBinding("user-viewer", authz.RoleProjectViewer, authz.ProjectScope(project.ID)))

	page, err := pageService.Create(ctx, admin, tracker.CreateCreatePageInput{
		ProjectID:    project.ID,
		Name:         "Bug Intake",
		Slug:         "Bug-Intake",
		Description:  "Structured bug form",
		Enabled:      true,
		TargetType:   "Bug",
		TargetStatus: "todo",
		FieldLayout:  []map[string]any{{"name": "title", "type": "text", "required": true}},
		Defaults: map[string]any{
			"priority": "High",
			"labels":   []any{"intake"},
		},
		OwnerUserID: "user-admin",
	})
	if err != nil {
		t.Fatalf("create page: %v", err)
	}
	if page.ID == "" || page.Slug != "bug-intake" || page.TargetType != "bug" || !page.Enabled {
		t.Fatalf("unexpected page: %#v", page)
	}

	listed, err := pageService.List(ctx, admin, tracker.ListCreatePagesInput{ProjectID: project.ID})
	if err != nil {
		t.Fatalf("list pages: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != page.ID {
		t.Fatalf("unexpected listed pages: %#v", listed)
	}

	resolved, err := pageService.Resolve(ctx, principal("user-member"), project.ID, "bug-intake")
	if err != nil {
		t.Fatalf("resolve page: %v", err)
	}
	if resolved.ID != page.ID {
		t.Fatalf("unexpected resolved page: %#v", resolved)
	}

	submitted, err := pageService.Submit(ctx, principal("user-member"), project.ID, "bug-intake", tracker.SubmitCreatePageInput{
		Ticket: tracker.CreateTicketInput{
			Title:       "Login form fails",
			Description: "Submitted through create page",
		},
	})
	if err != nil {
		t.Fatalf("submit page: %v", err)
	}
	if submitted.ProjectID != project.ID || submitted.Type != "bug" || submitted.Status != "todo" || submitted.Priority != "high" || submitted.ReporterID != "user-member" {
		t.Fatalf("unexpected submitted ticket: %#v", submitted)
	}
	if !slicesEqual(submitted.Labels, []string{"intake"}) {
		t.Fatalf("expected default labels, got %#v", submitted.Labels)
	}

	renamed := "Bug Intake Updated"
	enabled := false
	updated, err := pageService.Update(ctx, admin, page.ID, tracker.UpdateCreatePageInput{Name: &renamed, Enabled: &enabled})
	if err != nil {
		t.Fatalf("update page: %v", err)
	}
	if updated.Name != renamed || updated.Enabled {
		t.Fatalf("unexpected updated page: %#v", updated)
	}
	if _, err := pageService.Submit(ctx, principal("user-member"), project.ID, "bug-intake", tracker.SubmitCreatePageInput{Ticket: tracker.CreateTicketInput{Title: "Disabled"}}); !errors.Is(err, tracker.ErrNotFound) {
		t.Fatalf("expected disabled page not found on submit, got %v", err)
	}
	if err := pageService.Delete(ctx, admin, page.ID); err != nil {
		t.Fatalf("delete page: %v", err)
	}
	if _, err := pageService.Get(ctx, admin, page.ID); !errors.Is(err, tracker.ErrNotFound) {
		t.Fatalf("expected deleted page not found, got %v", err)
	}
}

func TestCreatePagePermissionsAndConflicts(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)
	seedUser(t, ctx, db.SQL, "user-viewer")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	trackerService := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	pageService := tracker.NewCreatePageService(db.SQL, trackerService, evaluator)
	admin := principal("user-admin")
	project, err := trackerService.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "CPG", Name: "Create Pages"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	evaluator.BindRole(authz.UserBinding("user-viewer", authz.RoleProjectViewer, authz.ProjectScope(project.ID)))

	page, err := pageService.Create(ctx, admin, tracker.CreateCreatePageInput{
		ProjectID: project.ID,
		Name:      "Task Intake",
		Slug:      "task-intake",
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("create page: %v", err)
	}
	if _, err := pageService.Create(ctx, admin, tracker.CreateCreatePageInput{ProjectID: project.ID, Name: "Duplicate", Slug: "task-intake", Enabled: true}); !errors.Is(err, tracker.ErrConflict) {
		t.Fatalf("expected duplicate slug conflict, got %v", err)
	}
	if _, err := pageService.Create(ctx, principal("user-member"), tracker.CreateCreatePageInput{ProjectID: project.ID, Name: "Nope", Slug: "nope", Enabled: true}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected member management forbidden, got %v", err)
	}
	if _, err := pageService.Submit(ctx, principal("user-viewer"), project.ID, page.Slug, tracker.SubmitCreatePageInput{Ticket: tracker.CreateTicketInput{Title: "Viewer submit"}}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected viewer submit forbidden from ticket create path, got %v", err)
	}
}

func TestCreatePageLuaSchemaTransform(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	trackerService := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	pageService := tracker.NewCreatePageService(db.SQL, trackerService, evaluator)
	admin := principal("user-admin")
	project, err := trackerService.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "LUA", Name: "Lua Forms"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))

	page, err := pageService.Create(ctx, admin, tracker.CreateCreatePageInput{
		ProjectID: project.ID,
		Name:      "Lua Intake",
		Slug:      "lua-intake",
		Enabled:   true,
		FieldLayout: []map[string]any{
			{"key": "title", "type": "text"},
		},
		Defaults: map[string]any{"priority": "Low"},
		FormLuaScript: `
return {
  description = page.description .. " resolved",
  field_layout = {
    { key = "title", type = "text", required = true },
    { key = "priority", type = "single-select", options = { "High", "Medium" } },
  },
  defaults = { priority = "High", labels = { "lua" } },
}
`,
	})
	if err != nil {
		t.Fatalf("create page: %v", err)
	}

	resolved, err := pageService.Resolve(ctx, principal("user-member"), project.ID, page.Slug)
	if err != nil {
		t.Fatalf("resolve dynamic page: %v", err)
	}
	if len(resolved.FieldLayout) != 2 || resolved.FieldLayout[1]["key"] != "priority" || resolved.Defaults["priority"] != "High" || resolved.Description != "resolved" {
		t.Fatalf("unexpected resolved page: %#v", resolved)
	}

	submitted, err := pageService.Submit(ctx, principal("user-member"), project.ID, page.Slug, tracker.SubmitCreatePageInput{
		Ticket: tracker.CreateTicketInput{Title: "Dynamic default"},
	})
	if err != nil {
		t.Fatalf("submit dynamic page: %v", err)
	}
	if submitted.Priority != "high" || !slicesEqual(submitted.Labels, []string{"lua"}) {
		t.Fatalf("expected dynamic defaults on submitted ticket, got %#v", submitted)
	}
}

func TestCreatePageLuaSchemaRejectsRawHTML(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	trackerService := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	pageService := tracker.NewCreatePageService(db.SQL, trackerService, evaluator)
	admin := principal("user-admin")
	project, err := trackerService.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "BAD", Name: "Bad Forms"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))

	page, err := pageService.Create(ctx, admin, tracker.CreateCreatePageInput{
		ProjectID:     project.ID,
		Name:          "Bad Intake",
		Slug:          "bad-intake",
		Enabled:       true,
		FormLuaScript: `return { field_layout = { { html = "<strong>no</strong>" } } }`,
	})
	if err != nil {
		t.Fatalf("create page: %v", err)
	}
	if _, err := pageService.Resolve(ctx, principal("user-member"), project.ID, page.Slug); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected raw HTML validation failure, got %v", err)
	}
}

func slicesEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
