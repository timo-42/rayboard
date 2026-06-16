package tracker_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
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
	})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	if ticket.Key != "CORE-1" || ticket.Status != "todo" || ticket.Priority != "high" || ticket.Type != "bug" || ticket.ReporterID != "user-member" {
		t.Fatalf("unexpected ticket: %#v", ticket)
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

	listed, err := service.ListTickets(ctx, member, tracker.ListTicketsInput{ProjectID: project.ID, Status: "in_progress"})
	if err != nil {
		t.Fatalf("list tickets: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != ticket.ID {
		t.Fatalf("unexpected tickets: %#v", listed)
	}

	activities, err = service.ListTicketActivity(ctx, member, ticket.ID)
	if err != nil {
		t.Fatalf("list updated activity: %v", err)
	}
	updatedActivity := ticketActivityByType(activities, "ticket.updated")
	if len(activities) != 2 || updatedActivity == nil {
		t.Fatalf("unexpected activity after update: %#v", activities)
	}
	changes, ok := updatedActivity.Data["changes"].(map[string]any)
	if !ok || changes["status"] == nil || changes["assignee_id"] == nil {
		t.Fatalf("expected status and assignee changes, got %#v", updatedActivity.Data)
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
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: "missing-project", Title: "Missing"}); !errors.Is(err, tracker.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}

	evaluator.BindRole(authz.UserBinding("user-viewer", authz.RoleProjectViewer, authz.ProjectScope(project.ID)))
	viewer := principal("user-viewer")
	if _, err := service.CreateTicket(ctx, viewer, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Denied"}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
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
