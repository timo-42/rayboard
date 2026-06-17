package tracker_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/events"
	"github.com/timo-42/rayboard/internal/backend/store"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func TestCreateProjectCreatesOwnerBindingAndListsReadableProjects(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-lead")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	principal := principal("user-admin")

	project, err := service.CreateProject(ctx, principal, tracker.CreateProjectInput{
		Key:         "core",
		Name:        "Core Tracking",
		Description: "Planning and ticket workflow",
		LeadUserID:  "user-lead",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if project.Key != "CORE" || project.LeadUserID != "user-lead" || project.CreatedBy != "user-admin" {
		t.Fatalf("unexpected project: %#v", project)
	}

	var bindings int
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM role_bindings
		WHERE role_id = ? AND subject_type = 'user' AND subject_id = ?
		  AND resource_type = 'project' AND resource_id = ?
	`, string(authz.RoleProjectOwner), "user-lead", project.ID).Scan(&bindings); err != nil {
		t.Fatalf("query owner binding: %v", err)
	}
	if bindings != 1 {
		t.Fatalf("expected one owner binding, got %d", bindings)
	}

	got, err := service.GetProject(ctx, principal, project.ID)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if got.ID != project.ID {
		t.Fatalf("expected project %s, got %#v", project.ID, got)
	}

	projects, err := service.ListProjects(ctx, principal, tracker.ListProjectsInput{})
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(projects) != 1 || projects[0].ID != project.ID {
		t.Fatalf("unexpected projects: %#v", projects)
	}
}

func TestProjectValidationConflictAndNotFound(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	principal := principal("user-admin")

	_, err := service.CreateProject(ctx, principal, tracker.CreateProjectInput{Key: "bad-key", Name: ""})
	if !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	var validationErr *tracker.ValidationError
	if !errors.As(err, &validationErr) || validationErr.Fields["key"] == "" || validationErr.Fields["name"] == "" {
		t.Fatalf("expected key and name field errors, got %#v", err)
	}

	if _, err := service.CreateProject(ctx, principal, tracker.CreateProjectInput{Key: "CORE", Name: "Core"}); err != nil {
		t.Fatalf("create first project: %v", err)
	}
	if _, err := service.CreateProject(ctx, principal, tracker.CreateProjectInput{Key: "core", Name: "Duplicate"}); !errors.Is(err, tracker.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}

	if _, err := service.GetProject(ctx, principal, "missing-project"); !errors.Is(err, tracker.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestTicketCreateListGetUpdateAndActivity(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedUser(t, ctx, db.SQL, "user-assignee")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "CORE", Name: "Core"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	member := principal("user-member")

	ticket, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{
		ProjectID:  project.ID,
		Title:      "First ticket",
		Priority:   "High",
		Type:       "Bug",
		AssigneeID: "user-assignee",
		Labels:     []string{"Backend", "backend", "API"},
	})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	if ticket.Key != "CORE-1" || ticket.Status != "todo" || ticket.Priority != "high" || ticket.Type != "bug" || ticket.ReporterID != "user-member" {
		t.Fatalf("unexpected ticket: %#v", ticket)
	}
	if !slices.Equal(ticket.Labels, []string{"api", "backend"}) {
		t.Fatalf("unexpected created labels: %#v", ticket.Labels)
	}

	second, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Second ticket"})
	if err != nil {
		t.Fatalf("create second ticket: %v", err)
	}
	if second.Key != "CORE-2" {
		t.Fatalf("expected second key CORE-2, got %s", second.Key)
	}

	activities, err := service.ListTicketActivity(ctx, member, ticket.ID)
	if err != nil {
		t.Fatalf("list initial activity: %v", err)
	}
	if len(activities) != 1 || activities[0].ActivityType != "ticket.created" || activities[0].ActorID != "user-member" {
		t.Fatalf("unexpected initial activity: %#v", activities)
	}
	if activities[0].Data["key"] != "CORE-1" {
		t.Fatalf("expected activity key CORE-1, got %#v", activities[0].Data)
	}

	title := "First ticket updated"
	status := "In_Progress"
	emptyAssignee := ""
	updated, err := service.UpdateTicket(ctx, member, ticket.ID, tracker.UpdateTicketInput{
		Title:      &title,
		Status:     &status,
		AssigneeID: &emptyAssignee,
	})
	if err != nil {
		t.Fatalf("update ticket: %v", err)
	}
	if updated.Title != title || updated.Status != "in_progress" || updated.AssigneeID != "" {
		t.Fatalf("unexpected updated ticket: %#v", updated)
	}

	got, err := service.GetTicket(ctx, member, ticket.ID)
	if err != nil {
		t.Fatalf("get ticket: %v", err)
	}
	if got.Title != title || got.Status != "in_progress" {
		t.Fatalf("unexpected fetched ticket: %#v", got)
	}
	if !slices.Equal(got.Labels, []string{"api", "backend"}) {
		t.Fatalf("expected fetched labels to be preserved, got %#v", got.Labels)
	}

	listed, err := service.ListTickets(ctx, member, tracker.ListTicketsInput{ProjectID: project.ID, Status: "in_progress", Label: "Backend"})
	if err != nil {
		t.Fatalf("list tickets: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != ticket.ID || !slices.Equal(listed[0].Labels, []string{"api", "backend"}) {
		t.Fatalf("unexpected tickets: %#v", listed)
	}

	labels := []string{"Docs", "api"}
	updatedLabels, err := service.UpdateTicket(ctx, member, ticket.ID, tracker.UpdateTicketInput{Labels: &labels})
	if err != nil {
		t.Fatalf("update labels: %v", err)
	}
	if !slices.Equal(updatedLabels.Labels, []string{"api", "docs"}) {
		t.Fatalf("unexpected updated labels: %#v", updatedLabels.Labels)
	}
	emptyTitlePreserve := "Labels preserved"
	preserved, err := service.UpdateTicket(ctx, member, ticket.ID, tracker.UpdateTicketInput{Title: &emptyTitlePreserve})
	if err != nil {
		t.Fatalf("update without labels: %v", err)
	}
	if !slices.Equal(preserved.Labels, []string{"api", "docs"}) {
		t.Fatalf("expected omitted labels to preserve existing labels, got %#v", preserved.Labels)
	}
	cleared := []string{}
	clearedLabels, err := service.UpdateTicket(ctx, member, ticket.ID, tracker.UpdateTicketInput{Labels: &cleared})
	if err != nil {
		t.Fatalf("clear labels: %v", err)
	}
	if len(clearedLabels.Labels) != 0 {
		t.Fatalf("expected cleared labels, got %#v", clearedLabels.Labels)
	}

	activities, err = service.ListTicketActivity(ctx, member, ticket.ID)
	if err != nil {
		t.Fatalf("list updated activity: %v", err)
	}
	updatedActivity := ticketActivityWithChanges(activities, "status", "assignee_id")
	if len(activities) != 5 || updatedActivity == nil {
		t.Fatalf("unexpected activity after update: %#v", activities)
	}
	changes, ok := updatedActivity.Data["changes"].(map[string]any)
	if !ok || changes["status"] == nil || changes["assignee_id"] == nil {
		t.Fatalf("expected status and assignee changes, got %#v", updatedActivity.Data)
	}
}

func TestTicketMutationsAppendDomainEvents(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow), tracker.WithEventStore(events.NewStore(db.SQL)))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "CORE", Name: "Core"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	member := principal("user-member")

	ticket, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "First ticket"})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	title := "Updated"
	if _, err := service.UpdateTicket(ctx, member, ticket.ID, tracker.UpdateTicketInput{Title: &title}); err != nil {
		t.Fatalf("update ticket: %v", err)
	}

	rows, err := db.SQL.QueryContext(ctx, `
		SELECT event_type, actor_id, project_id, subject_type, subject_id, processing_status
		FROM domain_events
		WHERE subject_type = 'ticket' AND subject_id = ?
		ORDER BY occurred_at, created_at
	`, ticket.ID)
	if err != nil {
		t.Fatalf("list domain events: %v", err)
	}
	defer rows.Close()

	var eventTypes []string
	for rows.Next() {
		var eventType string
		var actorID string
		var projectID string
		var subjectType string
		var subjectID string
		var status string
		if err := rows.Scan(&eventType, &actorID, &projectID, &subjectType, &subjectID, &status); err != nil {
			t.Fatalf("scan domain event: %v", err)
		}
		if actorID != "user-member" || projectID != project.ID || subjectType != "ticket" || subjectID != ticket.ID || status != "pending" {
			t.Fatalf("unexpected domain event row: %s %s %s %s %s %s", eventType, actorID, projectID, subjectType, subjectID, status)
		}
		eventTypes = append(eventTypes, eventType)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate domain events: %v", err)
	}
	if !slices.Equal(eventTypes, []string{"ticket.created", "ticket.updated"}) {
		t.Fatalf("unexpected event types: %#v", eventTypes)
	}
}

func TestTicketBeforeHooksTransformAndReject(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	hooks := tracker.NewHookService(db.SQL, evaluator)
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow), tracker.WithHookService(hooks))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "HOOK", Name: "Hooks"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	member := principal("user-member")

	if _, err := hooks.Create(ctx, admin, tracker.CreateHookInput{
		ProjectID: project.ID,
		Name:      "create defaults",
		Event:     tracker.HookEventTicketCreate,
		Phase:     tracker.HookPhaseBefore,
		Enabled:   true,
		Position:  10,
		Engine: tracker.HookEngineSpec{
			Type: tracker.HookEngineLua,
			Script: `
ticket.priority = "High"
ticket.type = "Bug"
ticket.labels = {"Hooked", "Lua"}
return { ticket = ticket }
`,
		},
	}); err != nil {
		t.Fatalf("create before hook: %v", err)
	}

	ticket, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Needs defaults"})
	if err != nil {
		t.Fatalf("create ticket with hook: %v", err)
	}
	if ticket.Priority != "high" || ticket.Type != "bug" || !slices.Equal(ticket.Labels, []string{"hooked", "lua"}) {
		t.Fatalf("expected hook-transformed ticket, got %#v", ticket)
	}

	if _, err := hooks.Create(ctx, admin, tracker.CreateHookInput{
		ProjectID: project.ID,
		Name:      "reject create",
		Event:     tracker.HookEventTicketCreate,
		Phase:     tracker.HookPhaseBefore,
		Enabled:   true,
		Position:  20,
		Engine: tracker.HookEngineSpec{
			Type:   tracker.HookEngineLua,
			Script: `return { reject = { message = "blocked by policy" } }`,
		},
	}); err != nil {
		t.Fatalf("create reject hook: %v", err)
	}
	ticketsBeforeReject := countTrackerRows(t, ctx, db.SQL, "tickets")
	activityBeforeReject := countTrackerRows(t, ctx, db.SQL, "ticket_activity")
	eventsBeforeReject := countTrackerRows(t, ctx, db.SQL, "domain_events")

	if _, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Blocked"}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected hook validation error, got %v", err)
	}
	if countTrackerRows(t, ctx, db.SQL, "tickets") != ticketsBeforeReject {
		t.Fatalf("expected rejected hook to leave ticket count unchanged")
	}
	if countTrackerRows(t, ctx, db.SQL, "ticket_activity") != activityBeforeReject {
		t.Fatalf("expected rejected hook to leave activity count unchanged")
	}
	if countTrackerRows(t, ctx, db.SQL, "domain_events") != eventsBeforeReject {
		t.Fatalf("expected rejected hook to leave domain event count unchanged")
	}
}

func TestTicketBeforeUpdateHookTransformsAndAfterHookDoesNotRollback(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	hooks := tracker.NewHookService(db.SQL, evaluator)
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow), tracker.WithHookService(hooks))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "UPD", Name: "Update Hooks"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	member := principal("user-member")
	ticket, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Original", Description: "Keep me"})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	if _, err := hooks.Create(ctx, admin, tracker.CreateHookInput{
		ProjectID: project.ID,
		Name:      "update defaults",
		Event:     tracker.HookEventTicketUpdate,
		Phase:     tracker.HookPhaseBefore,
		Enabled:   true,
		Position:  10,
		Engine: tracker.HookEngineSpec{
			Type: tracker.HookEngineLua,
			Script: `
ticket.priority = "High"
ticket.labels = {"updated"}
return { ticket = ticket }
`,
		},
	}); err != nil {
		t.Fatalf("create before update hook: %v", err)
	}
	afterHook, err := hooks.Create(ctx, admin, tracker.CreateHookInput{
		ProjectID: project.ID,
		Name:      "after fails",
		Event:     tracker.HookEventTicketUpdate,
		Phase:     tracker.HookPhaseAfter,
		Enabled:   true,
		Position:  10,
		Engine: tracker.HookEngineSpec{
			Type: tracker.HookEngineLua,
			Script: `
if rayboard.update_ticket ~= nil or rayboard.create_ticket ~= nil then
  error("mutating helper exposed")
end
error("after failed after commit")
`,
		},
	})
	if err != nil {
		t.Fatalf("create after hook: %v", err)
	}

	title := "Changed"
	updated, err := service.UpdateTicket(ctx, member, ticket.ID, tracker.UpdateTicketInput{Title: &title})
	if err != nil {
		t.Fatalf("update ticket with hooks: %v", err)
	}
	if updated.Title != title || updated.Description != "Keep me" || updated.Priority != "high" || !slices.Equal(updated.Labels, []string{"updated"}) {
		t.Fatalf("expected hook-transformed update, got %#v", updated)
	}
	if countTrackerRows(t, ctx, db.SQL, "ticket_activity") != 2 {
		t.Fatalf("expected committed update activity")
	}
	var lastError string
	if err := db.SQL.QueryRowContext(ctx, "SELECT COALESCE(last_error, '') FROM ticket_hooks WHERE id = ?", afterHook.ID).Scan(&lastError); err != nil {
		t.Fatalf("query after hook error: %v", err)
	}
	if lastError == "" {
		t.Fatalf("expected after hook error to be recorded")
	}
}

func ticketActivityByType(activities []tracker.TicketActivity, activityType string) *tracker.TicketActivity {
	for index := range activities {
		if activities[index].ActivityType == activityType {
			return &activities[index]
		}
	}
	return nil
}

func ticketActivityWithChanges(activities []tracker.TicketActivity, fields ...string) *tracker.TicketActivity {
	for index := range activities {
		if activities[index].ActivityType != "ticket.updated" {
			continue
		}
		changes, ok := activities[index].Data["changes"].(map[string]any)
		if !ok {
			continue
		}
		matches := true
		for _, field := range fields {
			if changes[field] == nil {
				matches = false
				break
			}
		}
		if matches {
			return &activities[index]
		}
	}
	return nil
}

func TestTicketValidationNotFoundAndAuthorization(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-viewer")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "CORE", Name: "Core"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Bad label", Labels: []string{""}}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected empty label validation error, got %v", err)
	}
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Bad label", Labels: []string{"not valid"}}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected invalid label validation error, got %v", err)
	}
	manyLabels := make([]string, 51)
	for index := range manyLabels {
		manyLabels[index] = "label_" + string(rune('a'+index%26))
	}
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Too many labels", Labels: manyLabels}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected too many labels validation error, got %v", err)
	}
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: "missing-project", Title: "Missing"}); !errors.Is(err, tracker.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}

	evaluator.BindRole(authz.UserBinding("user-viewer", authz.RoleProjectViewer, authz.ProjectScope(project.ID)))
	viewer := principal("user-viewer")
	if _, err := service.CreateTicket(ctx, viewer, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Denied"}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestRoadmapEpicsDatesAndProgress(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "CORE", Name: "Core"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	epic, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{
		ProjectID: project.ID,
		Title:     "Roadmap epic",
		Type:      "Epic",
		StartDate: "2026-07-01",
		DueDate:   "2026-07-31",
	})
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}
	if epic.Type != "epic" || epic.StartDate != "2026-07-01" || epic.DueDate != "2026-07-31" {
		t.Fatalf("unexpected epic fields: %#v", epic)
	}

	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{
		ProjectID: project.ID,
		Title:     "Bad date",
		Type:      "epic",
		StartDate: "2026/07/01",
	}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected invalid date validation, got %v", err)
	}
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{
		ProjectID: project.ID,
		Title:     "Bad range",
		Type:      "epic",
		StartDate: "2026-08-01",
		DueDate:   "2026-07-01",
	}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected date range validation, got %v", err)
	}

	childTodo, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{
		ProjectID:      project.ID,
		Title:          "Todo child",
		ParentTicketID: epic.ID,
	})
	if err != nil {
		t.Fatalf("create todo child: %v", err)
	}
	_, err = service.CreateTicket(ctx, admin, tracker.CreateTicketInput{
		ProjectID:      project.ID,
		Title:          "Done child",
		Status:         "done",
		ParentTicketID: epic.ID,
	})
	if err != nil {
		t.Fatalf("create done child: %v", err)
	}
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{
		ProjectID:      project.ID,
		Title:          "Nested child",
		ParentTicketID: childTodo.ID,
	}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected non-epic parent validation, got %v", err)
	}

	roadmap, err := service.ListRoadmap(ctx, admin, project.ID)
	if err != nil {
		t.Fatalf("list roadmap: %v", err)
	}
	if len(roadmap) != 1 || roadmap[0].Epic.ID != epic.ID {
		t.Fatalf("unexpected roadmap: %#v", roadmap)
	}
	if roadmap[0].Progress.Total != 2 || roadmap[0].Progress.Done != 1 || roadmap[0].Progress.ByStatus["todo"] != 1 || roadmap[0].Progress.ByStatus["done"] != 1 {
		t.Fatalf("unexpected roadmap progress: %#v", roadmap[0].Progress)
	}

	newDueDate := "2026-08-15"
	updated, err := service.UpdateTicket(ctx, admin, epic.ID, tracker.UpdateTicketInput{DueDate: &newDueDate})
	if err != nil {
		t.Fatalf("update epic due date: %v", err)
	}
	if updated.DueDate != newDueDate {
		t.Fatalf("expected updated due date, got %#v", updated)
	}
}

func TestSprintLifecycleAndTicketAssignment(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "CORE", Name: "Core"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	ticket, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Sprint ticket"})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	sprint, err := service.CreateSprint(ctx, admin, tracker.CreateSprintInput{
		ProjectID: project.ID,
		Name:      "Sprint 1",
		Goal:      "Ship sprint support",
		StartDate: "2026-06-17",
		EndDate:   "2026-06-30",
	})
	if err != nil {
		t.Fatalf("create sprint: %v", err)
	}
	if sprint.ID == "" || sprint.State != tracker.SprintStatePlanned {
		t.Fatalf("unexpected sprint: %#v", sprint)
	}

	listed, err := service.ListSprints(ctx, admin, project.ID, "")
	if err != nil {
		t.Fatalf("list sprints: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != sprint.ID {
		t.Fatalf("unexpected sprint list: %#v", listed)
	}

	assigned, err := service.SetTicketSprint(ctx, admin, ticket.ID, sprint.ID)
	if err != nil {
		t.Fatalf("assign ticket sprint: %v", err)
	}
	if assigned.SprintID != sprint.ID {
		t.Fatalf("expected sprint assignment, got %#v", assigned)
	}
	tickets, err := service.ListTickets(ctx, admin, tracker.ListTicketsInput{ProjectID: project.ID, SprintID: sprint.ID})
	if err != nil {
		t.Fatalf("list sprint tickets: %v", err)
	}
	if len(tickets) != 1 || tickets[0].ID != ticket.ID {
		t.Fatalf("unexpected sprint tickets: %#v", tickets)
	}

	started, err := service.StartSprint(ctx, admin, sprint.ID)
	if err != nil {
		t.Fatalf("start sprint: %v", err)
	}
	if started.State != tracker.SprintStateActive || started.StartedAt == nil {
		t.Fatalf("unexpected started sprint: %#v", started)
	}
	other, err := service.CreateSprint(ctx, admin, tracker.CreateSprintInput{ProjectID: project.ID, Name: "Sprint 2"})
	if err != nil {
		t.Fatalf("create other sprint: %v", err)
	}
	if _, err := service.StartSprint(ctx, admin, other.ID); !errors.Is(err, tracker.ErrConflict) {
		t.Fatalf("expected active sprint conflict, got %v", err)
	}

	completed, err := service.CompleteSprint(ctx, admin, sprint.ID)
	if err != nil {
		t.Fatalf("complete sprint: %v", err)
	}
	if completed.State != tracker.SprintStateCompleted || completed.CompletedAt == nil {
		t.Fatalf("unexpected completed sprint: %#v", completed)
	}
	removed, err := service.SetTicketSprint(ctx, admin, ticket.ID, "")
	if err != nil {
		t.Fatalf("remove ticket sprint: %v", err)
	}
	if removed.SprintID != "" {
		t.Fatalf("expected sprint removal, got %#v", removed)
	}
}

func TestBacklogListAndReorder(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "CORE", Name: "Core"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	first, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "First"})
	if err != nil {
		t.Fatalf("create first ticket: %v", err)
	}
	second, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Second"})
	if err != nil {
		t.Fatalf("create second ticket: %v", err)
	}

	backlog, err := service.ReorderBacklog(ctx, admin, project.ID, tracker.ReorderBacklogInput{TicketIDs: []string{second.ID, first.ID}})
	if err != nil {
		t.Fatalf("reorder backlog: %v", err)
	}
	if len(backlog) != 2 || backlog[0].ID != second.ID || backlog[0].Rank != "000001" || backlog[1].ID != first.ID || backlog[1].Rank != "000002" {
		t.Fatalf("unexpected reordered backlog: %#v", backlog)
	}

	listed, err := service.ListBacklog(ctx, admin, project.ID)
	if err != nil {
		t.Fatalf("list backlog: %v", err)
	}
	if len(listed) != 2 || listed[0].ID != second.ID || listed[1].ID != first.ID {
		t.Fatalf("unexpected listed backlog: %#v", listed)
	}

	if _, err := service.ReorderBacklog(ctx, admin, project.ID, tracker.ReorderBacklogInput{TicketIDs: []string{first.ID, first.ID}}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected duplicate validation error, got %v", err)
	}

	otherProject, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "OPS", Name: "Ops"})
	if err != nil {
		t.Fatalf("create other project: %v", err)
	}
	otherTicket, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: otherProject.ID, Title: "Other"})
	if err != nil {
		t.Fatalf("create other ticket: %v", err)
	}
	if _, err := service.ReorderBacklog(ctx, admin, project.ID, tracker.ReorderBacklogInput{TicketIDs: []string{otherTicket.ID}}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected cross-project validation error, got %v", err)
	}
}

func TestProjectStatusesAndBoards(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-viewer")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "CORE", Name: "Core"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	statuses, err := service.ListProjectStatuses(ctx, admin, project.ID)
	if err != nil {
		t.Fatalf("list statuses: %v", err)
	}
	if got := statusSlugs(statuses); len(got) != 3 || got[0] != "todo" || got[1] != "in_progress" || got[2] != "done" {
		t.Fatalf("unexpected default statuses: %#v", statuses)
	}

	boards, err := service.ListBoards(ctx, admin, project.ID)
	if err != nil {
		t.Fatalf("list boards: %v", err)
	}
	if len(boards) != 1 || boards[0].Name != "Default Board" || len(boards[0].Columns) != 3 {
		t.Fatalf("unexpected default board: %#v", boards)
	}

	blocked, err := service.ReplaceProjectStatuses(ctx, admin, project.ID, tracker.ReplaceProjectStatusesInput{Statuses: []tracker.ProjectStatusInput{
		{Slug: "todo", Name: "Todo"},
		{Slug: "blocked", Name: "Blocked"},
		{Slug: "done", Name: "Done"},
	}})
	if err != nil {
		t.Fatalf("replace statuses: %v", err)
	}
	if got := statusSlugs(blocked); len(got) != 3 || got[1] != "blocked" {
		t.Fatalf("unexpected replaced statuses: %#v", blocked)
	}

	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Blocked ticket", Status: "blocked"}); err != nil {
		t.Fatalf("create blocked ticket: %v", err)
	}
	if _, err := service.ReplaceProjectStatuses(ctx, admin, project.ID, tracker.ReplaceProjectStatusesInput{Statuses: []tracker.ProjectStatusInput{
		{Slug: "todo", Name: "Todo"},
		{Slug: "done", Name: "Done"},
	}}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected validation for removing used status, got %v", err)
	}

	board, err := service.CreateBoard(ctx, admin, tracker.CreateBoardInput{
		ProjectID:   project.ID,
		Name:        "Triage",
		Description: "Triage board",
		StatusSlugs: []string{"todo", "blocked", "done"},
	})
	if err != nil {
		t.Fatalf("create board: %v", err)
	}
	if board.ID == "" || len(board.Columns) != 3 || board.Columns[1].StatusSlug != "blocked" {
		t.Fatalf("unexpected board: %#v", board)
	}

	boardTickets, err := service.ListBoardTickets(ctx, admin, board.ID)
	if err != nil {
		t.Fatalf("list board tickets: %v", err)
	}
	if len(boardTickets.Columns) != 3 || len(boardTickets.Columns[1].Tickets) != 1 || boardTickets.Columns[1].Tickets[0].Status != "blocked" {
		t.Fatalf("unexpected board tickets: %#v", boardTickets)
	}

	name := "Triage Updated"
	columns := []string{"blocked", "done"}
	updated, err := service.UpdateBoard(ctx, admin, board.ID, tracker.UpdateBoardInput{Name: &name, StatusSlugs: &columns})
	if err != nil {
		t.Fatalf("update board: %v", err)
	}
	if updated.Name != name || len(updated.Columns) != 2 || updated.Columns[0].StatusSlug != "blocked" {
		t.Fatalf("unexpected updated board: %#v", updated)
	}

	evaluator.BindRole(authz.UserBinding("user-viewer", authz.RoleProjectViewer, authz.ProjectScope(project.ID)))
	if _, err := service.ReplaceProjectStatuses(ctx, principal("user-viewer"), project.ID, tracker.ReplaceProjectStatusesInput{Statuses: []tracker.ProjectStatusInput{{Slug: "todo", Name: "Todo"}}}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden for viewer status management, got %v", err)
	}

	if err := service.DeleteBoard(ctx, admin, board.ID); err != nil {
		t.Fatalf("delete board: %v", err)
	}
	if _, err := service.GetBoard(ctx, admin, board.ID); !errors.Is(err, tracker.ErrNotFound) {
		t.Fatalf("expected deleted board not found, got %v", err)
	}
}

func statusSlugs(statuses []tracker.ProjectStatus) []string {
	slugs := make([]string, len(statuses))
	for index, status := range statuses {
		slugs[index] = status.Slug
	}
	return slugs
}

func TestComponentsVersionsAndTicketAssignment(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-owner")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "CORE", Name: "Core"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	component, err := service.CreateComponent(ctx, admin, tracker.CreateComponentInput{
		ProjectID:         project.ID,
		Name:              "API",
		Description:       "Backend API",
		OwnerUserID:       "user-owner",
		DefaultAssigneeID: "user-admin",
	})
	if err != nil {
		t.Fatalf("create component: %v", err)
	}
	if component.ID == "" || component.OwnerUserID != "user-owner" {
		t.Fatalf("unexpected component: %#v", component)
	}

	version, err := service.CreateVersion(ctx, admin, tracker.CreateVersionInput{
		ProjectID:   project.ID,
		Name:        "1.0",
		Description: "First release",
		TargetDate:  "2026-07-01",
	})
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	if version.ID == "" || version.Status != tracker.VersionStatusPlanned {
		t.Fatalf("unexpected version: %#v", version)
	}

	ticket, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{
		ProjectID:   project.ID,
		Title:       "Component ticket",
		ComponentID: component.ID,
		VersionID:   version.ID,
	})
	if err != nil {
		t.Fatalf("create ticket with component/version: %v", err)
	}
	if ticket.ComponentID != component.ID || ticket.VersionID != version.ID {
		t.Fatalf("unexpected ticket component/version: %#v", ticket)
	}

	components, err := service.ListComponents(ctx, admin, project.ID)
	if err != nil {
		t.Fatalf("list components: %v", err)
	}
	if len(components) != 1 || components[0].ID != component.ID {
		t.Fatalf("unexpected components: %#v", components)
	}
	versions, err := service.ListVersions(ctx, admin, project.ID, tracker.VersionStatusPlanned)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 1 || versions[0].ID != version.ID {
		t.Fatalf("unexpected versions: %#v", versions)
	}

	name := "API Platform"
	updatedComponent, err := service.UpdateComponent(ctx, admin, component.ID, tracker.UpdateComponentInput{Name: &name})
	if err != nil {
		t.Fatalf("update component: %v", err)
	}
	if updatedComponent.Name != name {
		t.Fatalf("unexpected updated component: %#v", updatedComponent)
	}

	status := tracker.VersionStatusReleased
	releaseDate := "2026-07-03"
	updatedVersion, err := service.UpdateVersion(ctx, admin, version.ID, tracker.UpdateVersionInput{
		Status:      &status,
		ReleaseDate: &releaseDate,
	})
	if err != nil {
		t.Fatalf("update version: %v", err)
	}
	if updatedVersion.Status != tracker.VersionStatusReleased || updatedVersion.ReleaseDate != releaseDate {
		t.Fatalf("unexpected updated version: %#v", updatedVersion)
	}

	otherProject, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "OPS", Name: "Ops"})
	if err != nil {
		t.Fatalf("create other project: %v", err)
	}
	otherComponent, err := service.CreateComponent(ctx, admin, tracker.CreateComponentInput{ProjectID: otherProject.ID, Name: "Ops"})
	if err != nil {
		t.Fatalf("create other component: %v", err)
	}
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{
		ProjectID:   project.ID,
		Title:       "Wrong component",
		ComponentID: otherComponent.ID,
	}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected cross-project component validation, got %v", err)
	}
}

func TestCustomFieldsAndTicketValues(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-owner")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "CUST", Name: "Custom Fields"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	severity, err := service.CreateCustomField(ctx, admin, tracker.CreateCustomFieldInput{
		ProjectID: project.ID,
		Key:       "severity",
		Name:      "Severity",
		FieldType: tracker.CustomFieldTypeSingleSelect,
		Required:  true,
		Options:   []string{"Low", "High"},
	})
	if err != nil {
		t.Fatalf("create custom field: %v", err)
	}
	if severity.ID == "" || len(severity.Options) != 2 {
		t.Fatalf("unexpected severity field: %#v", severity)
	}

	_, err = service.CreateCustomField(ctx, admin, tracker.CreateCustomFieldInput{
		ProjectID: project.ID,
		Key:       "estimate",
		Name:      "Estimate",
		FieldType: tracker.CustomFieldTypeNumber,
	})
	if err != nil {
		t.Fatalf("create number custom field: %v", err)
	}

	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{
		ProjectID: project.ID,
		Title:     "Missing required field",
	}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected missing required field validation, got %v", err)
	}

	ticket, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{
		ProjectID: project.ID,
		Title:     "Custom field ticket",
		CustomFields: map[string]any{
			"severity": "High",
			"estimate": float64(3),
		},
	})
	if err != nil {
		t.Fatalf("create ticket with custom fields: %v", err)
	}
	if ticket.CustomFields["severity"] != "High" || ticket.CustomFields["estimate"] != float64(3) {
		t.Fatalf("unexpected ticket custom fields: %#v", ticket.CustomFields)
	}

	fetched, err := service.GetTicket(ctx, admin, ticket.ID)
	if err != nil {
		t.Fatalf("get ticket: %v", err)
	}
	if fetched.CustomFields["severity"] != "High" {
		t.Fatalf("expected persisted custom field, got %#v", fetched.CustomFields)
	}

	low := map[string]any{"severity": "Low"}
	updated, err := service.UpdateTicket(ctx, admin, ticket.ID, tracker.UpdateTicketInput{CustomFields: &low})
	if err != nil {
		t.Fatalf("update ticket custom fields: %v", err)
	}
	if updated.CustomFields["severity"] != "Low" {
		t.Fatalf("expected updated severity, got %#v", updated.CustomFields)
	}

	fields, err := service.ListCustomFields(ctx, admin, project.ID)
	if err != nil {
		t.Fatalf("list custom fields: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("expected two custom fields, got %#v", fields)
	}

	if err := service.DeleteCustomField(ctx, admin, severity.ID); err != nil {
		t.Fatalf("delete custom field: %v", err)
	}
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

func seedRole(t *testing.T, ctx context.Context, db *sql.DB, role authz.RoleName) {
	t.Helper()

	_, err := db.ExecContext(ctx, `
		INSERT INTO roles (id, name, description)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO NOTHING
	`, string(role), string(role), "Built-in test role")
	if err != nil {
		t.Fatalf("seed role %s: %v", role, err)
	}
}

func countTrackerRows(t *testing.T, ctx context.Context, db *sql.DB, table string) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
		t.Fatalf("count rows in %s: %v", table, err)
	}
	return count
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
