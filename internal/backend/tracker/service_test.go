package tracker_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/events"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
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

func TestCreateProjectAppendsDomainEvent(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-lead")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow), tracker.WithEventStore(events.NewStore(db.SQL)))
	principal := principal("user-admin")

	if _, err := service.CreateProject(ctx, principal, tracker.CreateProjectInput{Key: "bad-key", Name: ""}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected validation error before project event check, got %v", err)
	}
	if countTrackerRows(t, ctx, db.SQL, "domain_events") != 0 {
		t.Fatalf("validation failure should not append a domain event")
	}

	project, err := service.CreateProject(ctx, principal, tracker.CreateProjectInput{
		Key:        "CORE",
		Name:       "Core Tracking",
		LeadUserID: "user-lead",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	var eventType string
	var actorID string
	var projectID sql.NullString
	var subjectType string
	var subjectID string
	var status string
	var payload string
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT event_type, actor_id, project_id, subject_type, subject_id, processing_status, payload_json
		FROM domain_events
		WHERE subject_type = 'project' AND subject_id = ?
	`, project.ID).Scan(&eventType, &actorID, &projectID, &subjectType, &subjectID, &status, &payload); err != nil {
		t.Fatalf("query project domain event: %v", err)
	}
	if eventType != "project.created" || actorID != "user-admin" || !projectID.Valid || projectID.String != project.ID || subjectType != "project" || subjectID != project.ID || status != "pending" {
		t.Fatalf("unexpected project event row: %s %s %v %s %s %s", eventType, actorID, projectID, subjectType, subjectID, status)
	}
	var payloadMap map[string]any
	if err := json.Unmarshal([]byte(payload), &payloadMap); err != nil {
		t.Fatalf("decode project event payload: %v", err)
	}
	if payloadMap["key"] != "CORE" || payloadMap["name"] != "Core Tracking" {
		t.Fatalf("unexpected project event payload: %#v", payloadMap)
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

	initialStoryPoints := 3.5
	ticket, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{
		ProjectID:   project.ID,
		Title:       "First ticket",
		Priority:    "High",
		Type:        "Bug",
		AssigneeID:  "user-assignee",
		Labels:      []string{"Backend", "backend", "API"},
		StoryPoints: &initialStoryPoints,
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
	if ticket.StoryPoints == nil || *ticket.StoryPoints != 3.5 {
		t.Fatalf("unexpected created story points: %#v", ticket.StoryPoints)
	}

	second, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Second ticket", Labels: []string{"backend", "docs"}})
	if err != nil {
		t.Fatalf("create second ticket: %v", err)
	}
	if second.Key != "CORE-2" {
		t.Fatalf("expected second key CORE-2, got %s", second.Key)
	}
	projectLabels, err := service.ListProjectLabels(ctx, member, project.ID)
	if err != nil {
		t.Fatalf("list project labels: %v", err)
	}
	if !slices.Equal(projectLabels, []tracker.ProjectLabel{
		{ProjectID: project.ID, Label: "api", TicketCount: 1},
		{ProjectID: project.ID, Label: "backend", TicketCount: 2},
		{ProjectID: project.ID, Label: "docs", TicketCount: 1},
	}) {
		t.Fatalf("unexpected project labels: %#v", projectLabels)
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
	updatedStoryPoints := 5.0
	updated, err := service.UpdateTicket(ctx, member, ticket.ID, tracker.UpdateTicketInput{
		Title:          &title,
		Status:         &status,
		AssigneeID:     &emptyAssignee,
		StoryPoints:    &updatedStoryPoints,
		StoryPointsSet: true,
	})
	if err != nil {
		t.Fatalf("update ticket: %v", err)
	}
	if updated.Title != title || updated.Status != "in_progress" || updated.AssigneeID != "" || updated.StoryPoints == nil || *updated.StoryPoints != 5 {
		t.Fatalf("unexpected updated ticket: %#v", updated)
	}

	got, err := service.GetTicket(ctx, member, ticket.ID)
	if err != nil {
		t.Fatalf("get ticket: %v", err)
	}
	if got.Title != title || got.Status != "in_progress" || got.StoryPoints == nil || *got.StoryPoints != 5 {
		t.Fatalf("unexpected fetched ticket: %#v", got)
	}
	clearedPoints, err := service.UpdateTicket(ctx, member, ticket.ID, tracker.UpdateTicketInput{StoryPointsSet: true})
	if err != nil {
		t.Fatalf("clear story points: %v", err)
	}
	if clearedPoints.StoryPoints != nil {
		t.Fatalf("expected cleared story points, got %#v", clearedPoints.StoryPoints)
	}
	negativeStoryPoints := -1.0
	if _, err := service.UpdateTicket(ctx, member, ticket.ID, tracker.UpdateTicketInput{StoryPoints: &negativeStoryPoints, StoryPointsSet: true}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected invalid story points validation, got %v", err)
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
	projectLabels, err = service.ListProjectLabels(ctx, member, project.ID)
	if err != nil {
		t.Fatalf("list updated project labels: %v", err)
	}
	if !slices.Equal(projectLabels, []tracker.ProjectLabel{
		{ProjectID: project.ID, Label: "api", TicketCount: 1},
		{ProjectID: project.ID, Label: "backend", TicketCount: 1},
		{ProjectID: project.ID, Label: "docs", TicketCount: 2},
	}) {
		t.Fatalf("unexpected updated project labels: %#v", projectLabels)
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
	projectLabels, err = service.ListProjectLabels(ctx, member, project.ID)
	if err != nil {
		t.Fatalf("list cleared project labels: %v", err)
	}
	if !slices.Equal(projectLabels, []tracker.ProjectLabel{
		{ProjectID: project.ID, Label: "backend", TicketCount: 1},
		{ProjectID: project.ID, Label: "docs", TicketCount: 1},
	}) {
		t.Fatalf("unexpected cleared project labels: %#v", projectLabels)
	}
	if _, err := db.SQL.ExecContext(ctx, "UPDATE tickets SET deleted_at = ? WHERE id = ?", fixedNow().UTC().Format(time.RFC3339Nano), second.ID); err != nil {
		t.Fatalf("mark second ticket deleted: %v", err)
	}
	projectLabels, err = service.ListProjectLabels(ctx, member, project.ID)
	if err != nil {
		t.Fatalf("list labels after deleted ticket: %v", err)
	}
	if len(projectLabels) != 0 {
		t.Fatalf("expected deleted ticket labels to be excluded, got %#v", projectLabels)
	}
	if _, err := service.ListProjectLabels(ctx, principal("user-outsider"), project.ID); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden project label list, got %v", err)
	}

	catalogLabel, err := service.CreateProjectLabel(ctx, admin, tracker.CreateProjectLabelInput{
		ProjectID:   project.ID,
		Label:       "customer-escalation",
		Description: "Customer-facing escalations",
		Color:       "#FFAA00",
	})
	if err != nil {
		t.Fatalf("create project label: %v", err)
	}
	if catalogLabel.Label != "customer-escalation" || catalogLabel.Description != "Customer-facing escalations" || catalogLabel.Color != "#ffaa00" || catalogLabel.TicketCount != 0 || catalogLabel.CreatedAt.IsZero() {
		t.Fatalf("unexpected catalog label: %#v", catalogLabel)
	}
	if _, err := service.CreateProjectLabel(ctx, admin, tracker.CreateProjectLabelInput{ProjectID: project.ID, Label: "customer-escalation"}); !errors.Is(err, tracker.ErrConflict) {
		t.Fatalf("expected duplicate catalog label conflict, got %v", err)
	}
	if _, err := service.CreateProjectLabel(ctx, admin, tracker.CreateProjectLabelInput{ProjectID: project.ID, Label: "bad-label", Color: "orange"}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected invalid catalog label color validation, got %v", err)
	}
	description := "Escalations owned by support"
	color := "#00bbcc"
	catalogLabel, err = service.UpdateProjectLabel(ctx, admin, project.ID, "customer-escalation", tracker.UpdateProjectLabelInput{Description: &description, Color: &color})
	if err != nil {
		t.Fatalf("update project label: %v", err)
	}
	if catalogLabel.Description != description || catalogLabel.Color != "#00bbcc" || catalogLabel.UpdatedAt.IsZero() {
		t.Fatalf("unexpected updated catalog label: %#v", catalogLabel)
	}
	projectLabels, err = service.ListProjectLabels(ctx, member, project.ID)
	if err != nil {
		t.Fatalf("list project labels with catalog label: %v", err)
	}
	if len(projectLabels) != 1 || projectLabels[0].Label != "customer-escalation" || projectLabels[0].TicketCount != 0 {
		t.Fatalf("expected unused catalog label in list, got %#v", projectLabels)
	}
	if err := service.DeleteProjectLabel(ctx, admin, project.ID, "customer-escalation"); err != nil {
		t.Fatalf("delete project label: %v", err)
	}
	projectLabels, err = service.ListProjectLabels(ctx, member, project.ID)
	if err != nil {
		t.Fatalf("list project labels after catalog delete: %v", err)
	}
	if len(projectLabels) != 0 {
		t.Fatalf("expected catalog label to be removed without ticket labels, got %#v", projectLabels)
	}

	activities, err = service.ListTicketActivity(ctx, member, ticket.ID)
	if err != nil {
		t.Fatalf("list updated activity: %v", err)
	}
	updatedActivity := ticketActivityWithChanges(activities, "status", "assignee_id")
	if len(activities) != 6 || updatedActivity == nil {
		t.Fatalf("unexpected activity after update: %#v", activities)
	}
	changes, ok := updatedActivity.Data["changes"].(map[string]any)
	if !ok || changes["status"] == nil || changes["assignee_id"] == nil {
		t.Fatalf("expected status and assignee changes, got %#v", updatedActivity.Data)
	}
}

func TestTicketDeleteSoftDeletesAndRecordsActivity(t *testing.T) {
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

	ticket, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Delete me", Labels: []string{"cleanup"}})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	if err := service.DeleteTicket(ctx, member, ticket.ID); err != nil {
		t.Fatalf("delete ticket: %v", err)
	}

	if _, err := service.GetTicket(ctx, member, ticket.ID); !errors.Is(err, tracker.ErrNotFound) {
		t.Fatalf("expected deleted ticket to be hidden, got %v", err)
	}
	listed, err := service.ListTickets(ctx, member, tracker.ListTicketsInput{ProjectID: project.ID})
	if err != nil {
		t.Fatalf("list tickets after delete: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected deleted ticket to be omitted, got %#v", listed)
	}
	labels, err := service.ListProjectLabels(ctx, member, project.ID)
	if err != nil {
		t.Fatalf("list labels after delete: %v", err)
	}
	if len(labels) != 0 {
		t.Fatalf("expected deleted ticket labels to be omitted, got %#v", labels)
	}
	if err := service.DeleteTicket(ctx, member, ticket.ID); !errors.Is(err, tracker.ErrNotFound) {
		t.Fatalf("expected repeated delete not found, got %v", err)
	}

	var deletedAt string
	var updatedAt string
	if err := db.SQL.QueryRowContext(ctx, "SELECT deleted_at, updated_at FROM tickets WHERE id = ?", ticket.ID).Scan(&deletedAt, &updatedAt); err != nil {
		t.Fatalf("query deleted ticket: %v", err)
	}
	if deletedAt != fixedNow().UTC().Format(time.RFC3339Nano) || updatedAt != deletedAt {
		t.Fatalf("unexpected delete timestamps deleted_at=%q updated_at=%q", deletedAt, updatedAt)
	}

	var activityType string
	var actorID string
	var payload string
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT activity_type, actor_id, data_json
		FROM ticket_activity
		WHERE ticket_id = ? AND activity_type = 'ticket.deleted'
	`, ticket.ID).Scan(&activityType, &actorID, &payload); err != nil {
		t.Fatalf("query delete activity: %v", err)
	}
	if activityType != "ticket.deleted" || actorID != "user-member" {
		t.Fatalf("unexpected delete activity %s actor %s", activityType, actorID)
	}
	var activityData map[string]any
	if err := json.Unmarshal([]byte(payload), &activityData); err != nil {
		t.Fatalf("decode delete activity: %v", err)
	}
	if activityData["key"] != ticket.Key || activityData["title"] != ticket.Title || activityData["status"] != ticket.Status {
		t.Fatalf("unexpected delete activity data: %#v", activityData)
	}

	var eventType string
	var subjectType string
	var subjectID string
	var projectID sql.NullString
	if err := db.SQL.QueryRowContext(ctx, `
		SELECT event_type, subject_type, subject_id, project_id
		FROM domain_events
		WHERE event_type = 'ticket.deleted' AND subject_id = ?
	`, ticket.ID).Scan(&eventType, &subjectType, &subjectID, &projectID); err != nil {
		t.Fatalf("query delete domain event: %v", err)
	}
	if eventType != "ticket.deleted" || subjectType != "ticket" || subjectID != ticket.ID || !projectID.Valid || projectID.String != project.ID {
		t.Fatalf("unexpected delete domain event %s %s %s %#v", eventType, subjectType, subjectID, projectID)
	}
}

func TestTicketDeleteAuthorizationAndArchivedProject(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
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
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	evaluator.BindRole(authz.UserBinding("user-viewer", authz.RoleProjectViewer, authz.ProjectScope(project.ID)))
	member := principal("user-member")
	viewer := principal("user-viewer")

	ticket, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Protected"})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	if err := service.DeleteTicket(ctx, viewer, ticket.ID); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected viewer delete forbidden, got %v", err)
	}

	if _, err := db.SQL.ExecContext(ctx, "UPDATE projects SET archived_at = ? WHERE id = ?", fixedNow().UTC().Format(time.RFC3339Nano), project.ID); err != nil {
		t.Fatalf("archive project: %v", err)
	}
	if err := service.DeleteTicket(ctx, member, ticket.ID); !errors.Is(err, tracker.ErrConflict) {
		t.Fatalf("expected archived project conflict, got %v", err)
	}
	if _, err := service.GetTicket(ctx, member, ticket.ID); err != nil {
		t.Fatalf("ticket should remain visible after failed delete: %v", err)
	}
}

func TestTicketLinksCreateListDeleteAndValidate(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
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
	otherProject, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "OPS", Name: "Ops"})
	if err != nil {
		t.Fatalf("create other project: %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	evaluator.BindRole(authz.UserBinding("user-viewer", authz.RoleProjectViewer, authz.ProjectScope(project.ID)))
	member := principal("user-member")
	viewer := principal("user-viewer")

	sourcePoints := 2.0
	source, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Implement checkout", StoryPoints: &sourcePoints})
	if err != nil {
		t.Fatalf("create source ticket: %v", err)
	}
	targetPoints := 5.0
	target, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Fix payment gateway", StoryPoints: &targetPoints})
	if err != nil {
		t.Fatalf("create target ticket: %v", err)
	}
	other, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: otherProject.ID, Title: "Other project ticket"})
	if err != nil {
		t.Fatalf("create other ticket: %v", err)
	}

	link, err := service.CreateTicketLink(ctx, member, source.ID, tracker.CreateTicketLinkInput{
		TargetTicketID: target.ID,
		LinkType:       "Blocks",
	})
	if err != nil {
		t.Fatalf("create ticket link: %v", err)
	}
	if link.ID == "" || link.LinkType != "blocks" || link.Source.ID != source.ID || link.Target.ID != target.ID || link.CreatedBy != "user-member" {
		t.Fatalf("unexpected link: %#v", link)
	}

	links, err := service.ListTicketLinks(ctx, member, source.ID)
	if err != nil {
		t.Fatalf("list ticket links: %v", err)
	}
	if len(links) != 1 || links[0].ID != link.ID || links[0].Target.Key != target.Key {
		t.Fatalf("unexpected links: %#v", links)
	}
	if links[0].Source.StoryPoints == nil || *links[0].Source.StoryPoints != 2 || links[0].Target.StoryPoints == nil || *links[0].Target.StoryPoints != 5 {
		t.Fatalf("expected link tickets to include story points, got source=%#v target=%#v", links[0].Source.StoryPoints, links[0].Target.StoryPoints)
	}

	if _, err := service.CreateTicketLink(ctx, member, source.ID, tracker.CreateTicketLinkInput{TargetTicketID: target.ID, LinkType: "blocks"}); !errors.Is(err, tracker.ErrConflict) {
		t.Fatalf("expected duplicate conflict, got %v", err)
	}
	if _, err := service.CreateTicketLink(ctx, member, source.ID, tracker.CreateTicketLinkInput{TargetTicketID: source.ID, LinkType: "blocks"}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected self-link validation, got %v", err)
	}
	if _, err := service.CreateTicketLink(ctx, member, source.ID, tracker.CreateTicketLinkInput{TargetTicketID: target.ID, LinkType: "depends"}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected link type validation, got %v", err)
	}
	if _, err := service.CreateTicketLink(ctx, viewer, source.ID, tracker.CreateTicketLinkInput{TargetTicketID: target.ID, LinkType: "relates_to"}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected source write permission failure, got %v", err)
	}
	if _, err := service.CreateTicketLink(ctx, member, source.ID, tracker.CreateTicketLinkInput{TargetTicketID: other.ID, LinkType: "relates_to"}); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected target read permission failure, got %v", err)
	}

	activities, err := service.ListTicketActivity(ctx, member, source.ID)
	if err != nil {
		t.Fatalf("list link activity: %v", err)
	}
	createdActivity := ticketActivityByType(activities, "ticket.link_created")
	if len(activities) != 2 || createdActivity == nil || createdActivity.Data["target_key"] != target.Key {
		t.Fatalf("unexpected link-created activity: %#v", activities)
	}

	if err := service.DeleteTicketLink(ctx, member, source.ID, link.ID); err != nil {
		t.Fatalf("delete ticket link: %v", err)
	}
	links, err = service.ListTicketLinks(ctx, member, source.ID)
	if err != nil {
		t.Fatalf("list links after delete: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected deleted link to be omitted, got %#v", links)
	}
	activities, err = service.ListTicketActivity(ctx, member, source.ID)
	if err != nil {
		t.Fatalf("list delete activity: %v", err)
	}
	deletedActivity := ticketActivityByType(activities, "ticket.link_deleted")
	if len(activities) != 3 || deletedActivity == nil || deletedActivity.Data["link_id"] != link.ID {
		t.Fatalf("unexpected link-deleted activity: %#v", activities)
	}
	if err := service.DeleteTicketLink(ctx, member, source.ID, link.ID); !errors.Is(err, tracker.ErrNotFound) {
		t.Fatalf("expected deleted link not found, got %v", err)
	}
}

func TestRoadmapDependenciesIncludeEpicAndChildLinks(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
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

	epic, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Checkout epic", Type: "Epic"})
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}
	child, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Checkout child", ParentTicketID: epic.ID})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}
	nonRoadmap, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Standalone task"})
	if err != nil {
		t.Fatalf("create non-roadmap ticket: %v", err)
	}
	link, err := service.CreateTicketLink(ctx, member, epic.ID, tracker.CreateTicketLinkInput{TargetTicketID: child.ID, LinkType: "blocks"})
	if err != nil {
		t.Fatalf("create roadmap link: %v", err)
	}
	if _, err := service.CreateTicketLink(ctx, member, epic.ID, tracker.CreateTicketLinkInput{TargetTicketID: nonRoadmap.ID, LinkType: "relates_to"}); err != nil {
		t.Fatalf("create non-roadmap link: %v", err)
	}

	dependencies, err := service.ListRoadmapDependencies(ctx, member, project.ID)
	if err != nil {
		t.Fatalf("list roadmap dependencies: %v", err)
	}
	if len(dependencies) != 1 || dependencies[0].Link.ID != link.ID || dependencies[0].SourceEpicID != epic.ID || dependencies[0].TargetEpicID != epic.ID {
		t.Fatalf("unexpected roadmap dependencies: %#v", dependencies)
	}
	if dependencies[0].Link.Source.ID != epic.ID || dependencies[0].Link.Target.ID != child.ID {
		t.Fatalf("unexpected roadmap dependency tickets: %#v", dependencies[0].Link)
	}
	if _, err := service.ListRoadmapDependencies(ctx, principal("outsider"), project.ID); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden roadmap dependency list, got %v", err)
	}
}

func TestTicketWatchersListWatchUnwatchAndAuthorize(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
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
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	evaluator.BindRole(authz.UserBinding("user-viewer", authz.RoleProjectViewer, authz.ProjectScope(project.ID)))
	member := principal("user-member")
	viewer := principal("user-viewer")

	ticket, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Watch me"})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	watched, err := service.WatchTicket(ctx, viewer, ticket.ID)
	if err != nil {
		t.Fatalf("watch ticket: %v", err)
	}
	if !watched.Watching || watched.WatcherCount != 1 {
		t.Fatalf("expected viewer watch state on watched ticket, got %#v", watched)
	}
	fetchedByMember, err := service.GetTicket(ctx, member, ticket.ID)
	if err != nil {
		t.Fatalf("get ticket as member: %v", err)
	}
	if fetchedByMember.Watching || fetchedByMember.WatcherCount != 1 {
		t.Fatalf("expected member to see count but not own watch state, got %#v", fetchedByMember)
	}
	listed, err := service.ListTickets(ctx, viewer, tracker.ListTicketsInput{ProjectID: project.ID})
	if err != nil {
		t.Fatalf("list tickets as viewer: %v", err)
	}
	if len(listed) != 1 || !listed[0].Watching || listed[0].WatcherCount != 1 {
		t.Fatalf("expected watcher status in list, got %#v", listed)
	}
	watchers, err := service.ListTicketWatchers(ctx, member, ticket.ID)
	if err != nil {
		t.Fatalf("list watchers: %v", err)
	}
	if len(watchers) != 1 || watchers[0].UserID != "user-viewer" || watchers[0].DisplayName != "user-viewer" {
		t.Fatalf("unexpected watchers: %#v", watchers)
	}
	if _, err := service.WatchTicket(ctx, viewer, ticket.ID); err != nil {
		t.Fatalf("watch ticket idempotently: %v", err)
	}
	watchers, err = service.ListTicketWatchers(ctx, member, ticket.ID)
	if err != nil {
		t.Fatalf("list watchers after duplicate watch: %v", err)
	}
	if len(watchers) != 1 {
		t.Fatalf("expected duplicate watch to be idempotent, got %#v", watchers)
	}
	if _, err := service.WatchTicket(ctx, member, ticket.ID); err != nil {
		t.Fatalf("member watch ticket: %v", err)
	}
	title := "Watched update"
	updated, err := service.UpdateTicket(ctx, member, ticket.ID, tracker.UpdateTicketInput{Title: &title})
	if err != nil {
		t.Fatalf("update watched ticket: %v", err)
	}
	if !updated.Watching || updated.WatcherCount != 2 {
		t.Fatalf("expected update response to preserve member watcher state, got %#v", updated)
	}
	if _, err := service.WatchTicket(ctx, admin, ticket.ID); err != nil {
		t.Fatalf("admin watch ticket: %v", err)
	}
	sprint, err := service.CreateSprint(ctx, admin, tracker.CreateSprintInput{ProjectID: project.ID, Name: "Sprint 1"})
	if err != nil {
		t.Fatalf("create sprint: %v", err)
	}
	assigned, err := service.SetTicketSprint(ctx, admin, ticket.ID, sprint.ID)
	if err != nil {
		t.Fatalf("assign watched ticket sprint: %v", err)
	}
	if !assigned.Watching || assigned.WatcherCount != 3 {
		t.Fatalf("expected sprint response to preserve admin watcher state, got %#v", assigned)
	}
	activities, err := service.ListTicketActivity(ctx, member, ticket.ID)
	if err != nil {
		t.Fatalf("list activity: %v", err)
	}
	if ticketActivityByType(activities, "ticket.watcher_added") == nil {
		t.Fatalf("expected watcher added activity, got %#v", activities)
	}

	unwatched, err := service.UnwatchTicket(ctx, viewer, ticket.ID)
	if err != nil {
		t.Fatalf("unwatch ticket: %v", err)
	}
	if unwatched.Watching || unwatched.WatcherCount != 2 {
		t.Fatalf("expected watch state cleared, got %#v", unwatched)
	}
	if _, err := service.UnwatchTicket(ctx, viewer, ticket.ID); err != nil {
		t.Fatalf("unwatch ticket idempotently: %v", err)
	}
	if _, err := service.WatchTicket(ctx, principal("user-outsider"), ticket.ID); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden watch for outsider, got %v", err)
	}
	if _, err := service.ListTicketWatchers(ctx, principal("user-outsider"), ticket.ID); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden watcher list for outsider, got %v", err)
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

func TestTicketAIHookTransformsAndRejects(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedUser(t, ctx, db.SQL, "user-member")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	var prompts []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode OpenRouter request: %v", err)
		}
		prompt := ""
		if len(body.Messages) > 1 {
			prompt = body.Messages[1].Content
			prompts = append(prompts, prompt)
		}
		content := `{"ticket":{"title":"AI normalized","priority":"High","type":"Bug","labels":["ai","hook"]}}`
		if strings.Contains(prompt, "Reject disallowed tickets.") {
			content = `{"reject":{"message":"blocked by AI policy"}}`
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "gen_hook",
			"choices": []map[string]any{{
				"message": map[string]any{"role": "assistant", "content": content},
			}},
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

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	hooks := tracker.NewHookService(db.SQL, evaluator, tracker.WithHookOpenRouterService(openRouterService))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow), tracker.WithHookService(hooks))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "AIH", Name: "AI Hooks"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	member := principal("user-member")

	if _, err := hooks.Create(ctx, admin, tracker.CreateHookInput{
		ProjectID: project.ID,
		Name:      "ai normalize",
		Event:     tracker.HookEventTicketCreate,
		Phase:     tracker.HookPhaseBefore,
		Enabled:   true,
		Position:  10,
		Engine: tracker.HookEngineSpec{
			Type:       tracker.HookEngineAI,
			Prompt:     "Normalize incoming ticket fields.",
			ProviderID: provider.ID,
		},
	}); err != nil {
		t.Fatalf("create AI transform hook: %v", err)
	}

	ticket, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "needs ai"})
	if err != nil {
		t.Fatalf("create ticket with AI hook: %v", err)
	}
	if ticket.Title != "AI normalized" || ticket.Priority != "high" || ticket.Type != "bug" || !slices.Equal(ticket.Labels, []string{"ai", "hook"}) {
		t.Fatalf("expected AI-transformed ticket, got %#v", ticket)
	}
	if len(prompts) != 1 || !strings.Contains(prompts[0], "Normalize incoming ticket fields.") || !strings.Contains(prompts[0], `"ticket"`) {
		t.Fatalf("unexpected AI hook prompt: %#v", prompts)
	}
	if strings.Contains(prompts[0], "sk-or-secret") {
		t.Fatalf("prompt leaked OpenRouter secret: %s", prompts[0])
	}

	if _, err := hooks.Create(ctx, admin, tracker.CreateHookInput{
		ProjectID: project.ID,
		Name:      "ai reject",
		Event:     tracker.HookEventTicketCreate,
		Phase:     tracker.HookPhaseBefore,
		Enabled:   true,
		Position:  20,
		Engine: tracker.HookEngineSpec{
			Type:       tracker.HookEngineAI,
			Prompt:     "Reject disallowed tickets.",
			ProviderID: provider.ID,
		},
	}); err != nil {
		t.Fatalf("create AI reject hook: %v", err)
	}
	ticketsBeforeReject := countTrackerRows(t, ctx, db.SQL, "tickets")
	if _, err := service.CreateTicket(ctx, member, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Blocked"}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected AI hook validation error, got %v", err)
	}
	if countTrackerRows(t, ctx, db.SQL, "tickets") != ticketsBeforeReject {
		t.Fatalf("expected rejected AI hook to leave ticket count unchanged")
	}
}

func TestTicketHookPreviewDoesNotPersistLastError(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	hooks := tracker.NewHookService(db.SQL, evaluator)
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow), tracker.WithHookService(hooks))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "HPV", Name: "Hook Preview"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	hook, err := hooks.Create(ctx, admin, tracker.CreateHookInput{
		ProjectID: project.ID,
		Name:      "preview",
		Event:     tracker.HookEventTicketCreate,
		Phase:     tracker.HookPhaseBefore,
		Enabled:   false,
		Position:  10,
		Engine: tracker.HookEngineSpec{
			Type: tracker.HookEngineLua,
			Script: `
rayboard.log("previewing " .. ticket.title)
ticket.priority = "High"
return { ticket = ticket }
`,
		},
	})
	if err != nil {
		t.Fatalf("create hook: %v", err)
	}

	preview, err := hooks.Preview(ctx, admin, hook.ID, tracker.PreviewHookInput{
		Ticket: map[string]any{"title": "Preview me"},
	})
	if err != nil {
		t.Fatalf("preview hook: %v", err)
	}
	ticket, ok := preview.Output["ticket"].(map[string]any)
	if !ok || ticket["priority"] != "High" || preview.Error != "" || !slices.Equal(preview.Logs, []string{"previewing Preview me"}) {
		t.Fatalf("unexpected preview result: %#v", preview)
	}
	stored, err := hooks.Get(ctx, admin, hook.ID)
	if err != nil {
		t.Fatalf("get hook: %v", err)
	}
	if stored.LastError != "" {
		t.Fatalf("preview should not persist last error, got %q", stored.LastError)
	}

	badScript := `return { reject = { message = "blocked in preview" } }`
	if _, err := hooks.Update(ctx, admin, hook.ID, tracker.UpdateHookInput{Engine: &tracker.HookEngineSpec{Type: tracker.HookEngineLua, Script: badScript}}); err != nil {
		t.Fatalf("update hook script: %v", err)
	}
	rejected, err := hooks.Preview(ctx, admin, hook.ID, tracker.PreviewHookInput{
		Ticket: map[string]any{"title": "Rejected"},
	})
	if err != nil {
		t.Fatalf("preview rejected hook: %v", err)
	}
	if rejected.Error == "" || rejected.Output["reject"] == nil {
		t.Fatalf("expected preview error and reject output, got %#v", rejected)
	}
	stored, err = hooks.Get(ctx, admin, hook.ID)
	if err != nil {
		t.Fatalf("get hook after reject: %v", err)
	}
	if stored.LastError != "" {
		t.Fatalf("rejected preview should not persist last error, got %q", stored.LastError)
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

	scheduled, err := service.ScheduleRoadmap(ctx, admin, project.ID, tracker.RoadmapScheduleInput{
		TicketID:  epic.ID,
		StartDate: "2026-08-01",
		DueDate:   "2026-08-31",
	})
	if err != nil {
		t.Fatalf("schedule roadmap: %v", err)
	}
	if len(scheduled) != 1 || scheduled[0].Epic.StartDate != "2026-08-01" || scheduled[0].Epic.DueDate != "2026-08-31" {
		t.Fatalf("unexpected scheduled roadmap: %#v", scheduled)
	}
	cleared, err := service.ScheduleRoadmap(ctx, admin, project.ID, tracker.RoadmapScheduleInput{
		TicketID: epic.ID,
	})
	if err != nil {
		t.Fatalf("clear roadmap schedule: %v", err)
	}
	if len(cleared) != 1 || cleared[0].Epic.StartDate != "" || cleared[0].Epic.DueDate != "" {
		t.Fatalf("expected cleared schedule, got %#v", cleared)
	}
	if _, err := service.ScheduleRoadmap(ctx, admin, project.ID, tracker.RoadmapScheduleInput{}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected empty schedule validation, got %v", err)
	}
	if _, err := service.ScheduleRoadmap(ctx, admin, project.ID, tracker.RoadmapScheduleInput{
		TicketID:  childTodo.ID,
		StartDate: "2026-08-01",
	}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected non-epic schedule validation, got %v", err)
	}
	if _, err := service.ScheduleRoadmap(ctx, admin, project.ID, tracker.RoadmapScheduleInput{
		TicketID:  epic.ID,
		StartDate: "2026-09-01",
		DueDate:   "2026-08-01",
	}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected schedule date range validation, got %v", err)
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
	activeReport, err := service.GetSprintReport(ctx, admin, sprint.ID)
	if err != nil {
		t.Fatalf("get active sprint report: %v", err)
	}
	if activeReport.Scope != tracker.SprintReportScopeCurrent || activeReport.SnapshotAt != nil || activeReport.Progress.Total != 1 || activeReport.Progress.Done != 0 || len(activeReport.Tickets) != 1 || activeReport.Tickets[0].ID != ticket.ID {
		t.Fatalf("unexpected active sprint report: %#v", activeReport)
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
	completedReport, err := service.GetSprintReport(ctx, admin, sprint.ID)
	if err != nil {
		t.Fatalf("get completed sprint report: %v", err)
	}
	if completedReport.Scope != tracker.SprintReportScopeSnapshot || completedReport.SnapshotAt == nil || !completedReport.SnapshotAt.Equal(*completed.CompletedAt) || completedReport.Progress.Total != 1 || len(completedReport.Tickets) != 1 || completedReport.Tickets[0].ID != ticket.ID {
		t.Fatalf("unexpected completed sprint report: %#v", completedReport)
	}
	removed, err := service.SetTicketSprint(ctx, admin, ticket.ID, "")
	if err != nil {
		t.Fatalf("remove ticket sprint: %v", err)
	}
	if removed.SprintID != "" {
		t.Fatalf("expected sprint removal, got %#v", removed)
	}
	completedReportAfterMove, err := service.GetSprintReport(ctx, admin, sprint.ID)
	if err != nil {
		t.Fatalf("get completed sprint report after move: %v", err)
	}
	if completedReportAfterMove.Scope != tracker.SprintReportScopeSnapshot || completedReportAfterMove.Progress.Total != 1 || len(completedReportAfterMove.Tickets) != 1 || completedReportAfterMove.Tickets[0].ID != ticket.ID {
		t.Fatalf("expected completed report to keep committed scope, got %#v", completedReportAfterMove)
	}

	emptyStarted, err := service.StartSprint(ctx, admin, other.ID)
	if err != nil {
		t.Fatalf("start empty sprint: %v", err)
	}
	emptyCompleted, err := service.CompleteSprint(ctx, admin, emptyStarted.ID)
	if err != nil {
		t.Fatalf("complete empty sprint: %v", err)
	}
	emptyReport, err := service.GetSprintReport(ctx, admin, emptyStarted.ID)
	if err != nil {
		t.Fatalf("get empty completed sprint report: %v", err)
	}
	if emptyReport.Scope != tracker.SprintReportScopeSnapshot || emptyReport.SnapshotAt == nil || !emptyReport.SnapshotAt.Equal(*emptyCompleted.CompletedAt) || emptyReport.Progress.Total != 0 || len(emptyReport.Tickets) != 0 {
		t.Fatalf("unexpected empty completed sprint report: %#v", emptyReport)
	}
}

func TestSprintReportAnalytics(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)

	now := fixedNow()
	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(func() time.Time { return now }))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "RPT", Name: "Reports"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	sprint, err := service.CreateSprint(ctx, admin, tracker.CreateSprintInput{
		ProjectID: project.ID,
		Name:      "Analytics Sprint",
		StartDate: "2026-06-16",
		EndDate:   "2026-06-16",
	})
	if err != nil {
		t.Fatalf("create sprint: %v", err)
	}
	todoPoints := 2.0
	todo, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Todo report ticket", StoryPoints: &todoPoints})
	if err != nil {
		t.Fatalf("create todo ticket: %v", err)
	}
	donePoints := 3.0
	done, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Done report ticket", Status: "done", StoryPoints: &donePoints})
	if err != nil {
		t.Fatalf("create done ticket: %v", err)
	}
	postWindowPoints := 5.0
	postWindow, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Late done report ticket", StoryPoints: &postWindowPoints})
	if err != nil {
		t.Fatalf("create late done ticket: %v", err)
	}
	if _, err := service.SetTicketSprint(ctx, admin, todo.ID, sprint.ID); err != nil {
		t.Fatalf("assign todo ticket: %v", err)
	}
	if _, err := service.SetTicketSprint(ctx, admin, done.ID, sprint.ID); err != nil {
		t.Fatalf("assign done ticket: %v", err)
	}
	if _, err := service.SetTicketSprint(ctx, admin, postWindow.ID, sprint.ID); err != nil {
		t.Fatalf("assign late done ticket: %v", err)
	}
	now = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	postWindowStatus := "done"
	if _, err := service.UpdateTicket(ctx, admin, postWindow.ID, tracker.UpdateTicketInput{Status: &postWindowStatus}); err != nil {
		t.Fatalf("mark late ticket done: %v", err)
	}

	report, err := service.GetSprintReport(ctx, admin, sprint.ID)
	if err != nil {
		t.Fatalf("get sprint report: %v", err)
	}
	if report.Progress.StoryPointsTotal != 10 || report.Progress.StoryPointsDone != 8 || report.Progress.StoryPointsRemaining != 2 || report.Progress.StoryPointsUnestimated != 0 {
		t.Fatalf("unexpected sprint story point progress: %#v", report.Progress)
	}
	if report.Analytics.Velocity.Completed != 3 || report.Analytics.Velocity.Unit != "points" {
		t.Fatalf("unexpected sprint velocity: %#v", report.Analytics.Velocity)
	}
	if len(report.Analytics.Burndown) != 1 || report.Analytics.Burndown[0].Date != "2026-06-16" || report.Analytics.Burndown[0].Remaining != 7 {
		t.Fatalf("unexpected sprint burndown: %#v", report.Analytics.Burndown)
	}
	if len(report.Analytics.Burnup) != 1 || report.Analytics.Burnup[0].Date != "2026-06-16" || report.Analytics.Burnup[0].Total != 10 || report.Analytics.Burnup[0].Done != 3 {
		t.Fatalf("unexpected sprint burnup: %#v", report.Analytics.Burnup)
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
		WIPLimits:   map[string]int{"blocked": 0},
	})
	if err != nil {
		t.Fatalf("create board: %v", err)
	}
	if board.ID == "" || len(board.Columns) != 3 || board.Columns[1].StatusSlug != "blocked" || board.Columns[1].WIPLimit == nil || *board.Columns[1].WIPLimit != 0 {
		t.Fatalf("unexpected board: %#v", board)
	}
	if _, err := service.CreateBoard(ctx, admin, tracker.CreateBoardInput{
		ProjectID:   project.ID,
		Name:        "Invalid limit",
		StatusSlugs: []string{"todo", "blocked"},
		WIPLimits:   map[string]int{"done": -1},
	}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected validation for negative WIP limit, got %v", err)
	}
	if _, err := service.CreateBoard(ctx, admin, tracker.CreateBoardInput{
		ProjectID:   project.ID,
		Name:        "Unknown limit",
		StatusSlugs: []string{"todo", "blocked"},
		WIPLimits:   map[string]int{"done": 1},
	}); !errors.Is(err, tracker.ErrValidation) {
		t.Fatalf("expected validation for unknown WIP limit status, got %v", err)
	}

	boardTickets, err := service.ListBoardTickets(ctx, admin, board.ID)
	if err != nil {
		t.Fatalf("list board tickets: %v", err)
	}
	if len(boardTickets.Columns) != 3 || len(boardTickets.Columns[1].Tickets) != 1 || boardTickets.Columns[1].Tickets[0].Status != "blocked" || boardTickets.Columns[1].TicketCount != 1 || !boardTickets.Columns[1].OverWIPLimit {
		t.Fatalf("unexpected board tickets: %#v", boardTickets)
	}

	updatedLimits := map[string]int{"blocked": 2}
	limitOnly, err := service.UpdateBoard(ctx, admin, board.ID, tracker.UpdateBoardInput{WIPLimits: &updatedLimits})
	if err != nil {
		t.Fatalf("update board limits: %v", err)
	}
	if limitOnly.Columns[1].WIPLimit == nil || *limitOnly.Columns[1].WIPLimit != 2 {
		t.Fatalf("unexpected limit-only board update: %#v", limitOnly)
	}

	name := "Triage Updated"
	columns := []string{"blocked", "done"}
	updated, err := service.UpdateBoard(ctx, admin, board.ID, tracker.UpdateBoardInput{Name: &name, StatusSlugs: &columns})
	if err != nil {
		t.Fatalf("update board: %v", err)
	}
	if updated.Name != name || len(updated.Columns) != 2 || updated.Columns[0].StatusSlug != "blocked" || updated.Columns[0].WIPLimit == nil || *updated.Columns[0].WIPLimit != 2 {
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

func TestVersionReportSummarizesAssignedTickets(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	seedRole(t, ctx, db.SQL, authz.RoleProjectOwner)
	seedRole(t, ctx, db.SQL, authz.RoleProjectMember)

	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(
		authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope()),
	))
	service := tracker.NewService(db.SQL, evaluator, tracker.WithNow(fixedNow))
	admin := principal("user-admin")
	project, err := service.CreateProject(ctx, admin, tracker.CreateProjectInput{Key: "REL", Name: "Release Reports"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	version, err := service.CreateVersion(ctx, admin, tracker.CreateVersionInput{ProjectID: project.ID, Name: "2026.7"})
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	otherVersion, err := service.CreateVersion(ctx, admin, tracker.CreateVersionInput{ProjectID: project.ID, Name: "2026.8"})
	if err != nil {
		t.Fatalf("create other version: %v", err)
	}
	directReleasedVersion, err := service.CreateVersion(ctx, admin, tracker.CreateVersionInput{ProjectID: project.ID, Name: "2026.9", Status: tracker.VersionStatusReleased})
	if err != nil {
		t.Fatalf("create directly released version: %v", err)
	}
	component, err := service.CreateComponent(ctx, admin, tracker.CreateComponentInput{ProjectID: project.ID, Name: "API"})
	if err != nil {
		t.Fatalf("create component: %v", err)
	}

	doneStatus := "done"
	versionID := version.ID
	componentID := component.ID
	firstPoints := 2.0
	first, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Fix blocker", VersionID: version.ID, StoryPoints: &firstPoints})
	if err != nil {
		t.Fatalf("create first ticket: %v", err)
	}
	secondPoints := 5.0
	second, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Ship feature", ComponentID: component.ID, VersionID: version.ID, StoryPoints: &secondPoints})
	if err != nil {
		t.Fatalf("create second ticket: %v", err)
	}
	second, err = service.UpdateTicket(ctx, admin, second.ID, tracker.UpdateTicketInput{Status: &doneStatus, ComponentID: &componentID, VersionID: &versionID})
	if err != nil {
		t.Fatalf("mark second ticket done: %v", err)
	}
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Needs estimate", VersionID: version.ID}); err != nil {
		t.Fatalf("create unestimated ticket: %v", err)
	}
	zeroPoints := 0.0
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Zero point estimate", VersionID: version.ID, StoryPoints: &zeroPoints}); err != nil {
		t.Fatalf("create zero point ticket: %v", err)
	}
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "Other release", VersionID: otherVersion.ID}); err != nil {
		t.Fatalf("create other version ticket: %v", err)
	}
	if _, err := service.CreateTicket(ctx, admin, tracker.CreateTicketInput{ProjectID: project.ID, Title: "After direct release", VersionID: directReleasedVersion.ID}); err != nil {
		t.Fatalf("create ticket after direct release: %v", err)
	}
	directReleasedReport, err := service.GetVersionReport(ctx, admin, directReleasedVersion.ID)
	if err != nil {
		t.Fatalf("get directly released version report: %v", err)
	}
	if directReleasedReport.Scope != tracker.VersionReportScopeSnapshot ||
		directReleasedReport.SnapshotAt == nil ||
		directReleasedReport.Progress.Total != 0 ||
		directReleasedReport.ScopeChanges.Current != 1 ||
		directReleasedReport.ScopeChanges.Added != 1 ||
		len(directReleasedReport.Tickets) != 0 {
		t.Fatalf("expected directly released version to keep empty release snapshot, got %#v", directReleasedReport)
	}

	report, err := service.GetVersionReport(ctx, admin, version.ID)
	if err != nil {
		t.Fatalf("get version report: %v", err)
	}
	if report.Version.ID != version.ID ||
		report.Scope != tracker.VersionReportScopeCurrent ||
		report.SnapshotAt != nil ||
		report.Progress.Total != 4 ||
		report.Progress.Done != 1 ||
		report.Progress.Open != 3 ||
		report.Progress.UnassignedComponent != 3 ||
		report.Progress.StoryPointsTotal != 7 ||
		report.Progress.StoryPointsDone != 5 ||
		report.Progress.StoryPointsRemaining != 2 ||
		report.Progress.StoryPointsUnestimated != 1 ||
		report.Progress.Total-report.Progress.StoryPointsUnestimated != 3 ||
		report.Progress.ByStatus["todo"] != 3 ||
		report.Progress.ByStatus["done"] != 1 ||
		report.ScopeChanges.Current != 4 ||
		report.ScopeChanges.Unchanged != 4 ||
		len(report.Tickets) != 4 {
		t.Fatalf("unexpected version report: %#v", report)
	}
	reportIDs := map[string]bool{}
	for _, ticket := range report.Tickets {
		reportIDs[ticket.ID] = true
		if ticket.VersionID != version.ID {
			t.Fatalf("unexpected ticket version in report: %#v", ticket)
		}
	}
	if !reportIDs[first.ID] || !reportIDs[second.ID] {
		t.Fatalf("expected assigned tickets in report, got %#v", report.Tickets)
	}

	released := tracker.VersionStatusReleased
	updatedVersion, err := service.UpdateVersion(ctx, admin, version.ID, tracker.UpdateVersionInput{Status: &released})
	if err != nil {
		t.Fatalf("release version: %v", err)
	}
	if updatedVersion.Status != tracker.VersionStatusReleased {
		t.Fatalf("expected released version, got %#v", updatedVersion)
	}
	otherVersionID := otherVersion.ID
	if _, err := service.UpdateTicket(ctx, admin, first.ID, tracker.UpdateTicketInput{VersionID: &otherVersionID}); err != nil {
		t.Fatalf("move first ticket after release: %v", err)
	}
	if _, err := db.SQL.ExecContext(ctx, "UPDATE tickets SET deleted_at = ? WHERE id = ?", fixedNow().UTC().Format(time.RFC3339Nano), second.ID); err != nil {
		t.Fatalf("delete second ticket after release: %v", err)
	}
	snapshotReport, err := service.GetVersionReport(ctx, admin, version.ID)
	if err != nil {
		t.Fatalf("get released version report: %v", err)
	}
	if snapshotReport.Scope != tracker.VersionReportScopeSnapshot || snapshotReport.SnapshotAt == nil || !snapshotReport.SnapshotAt.Equal(updatedVersion.UpdatedAt) {
		t.Fatalf("unexpected version report scope: %#v", snapshotReport)
	}
	if snapshotReport.Progress.Total != 3 ||
		snapshotReport.Progress.Done != 0 ||
		snapshotReport.Progress.Open != 3 ||
		snapshotReport.Progress.StoryPointsTotal != 2 ||
		snapshotReport.Progress.StoryPointsDone != 0 ||
		snapshotReport.Progress.StoryPointsRemaining != 2 ||
		snapshotReport.Progress.StoryPointsUnestimated != 1 ||
		snapshotReport.ScopeChanges.Current != 2 ||
		snapshotReport.ScopeChanges.Snapshot != 3 ||
		snapshotReport.ScopeChanges.Removed != 1 ||
		snapshotReport.ScopeChanges.Unchanged != 2 ||
		len(snapshotReport.Tickets) != 3 {
		t.Fatalf("unexpected released snapshot report after move/delete: %#v", snapshotReport)
	}

	member := principal("user-member")
	if _, err := service.GetVersionReport(ctx, member, version.ID); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden version report, got %v", err)
	}
	evaluator.BindRole(authz.UserBinding("user-member", authz.RoleProjectMember, authz.ProjectScope(project.ID)))
	if _, err := service.GetVersionReport(ctx, member, version.ID); err != nil {
		t.Fatalf("expected project member to read version report: %v", err)
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
