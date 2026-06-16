package tracker

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/events"
)

const (
	defaultTicketStatus = "todo"

	activityTicketCreated = "ticket.created"
	activityTicketUpdated = "ticket.updated"
)

var (
	projectKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{0,15}$`)
	slugPattern       = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)
)

type Service struct {
	db         *sql.DB
	repo       *Repository
	authorizer authz.Evaluator
	now        func() time.Time
	eventBus   *events.Bus
}

type Option func(*Service)

func NewService(db *sql.DB, authorizer authz.Evaluator, options ...Option) *Service {
	service := &Service{
		db:         db,
		repo:       NewRepository(db),
		authorizer: authorizer,
		now:        func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithNow(now func() time.Time) Option {
	return func(service *Service) {
		if now != nil {
			service.now = now
		}
	}
}

func WithEventBus(bus *events.Bus) Option {
	return func(service *Service) {
		service.eventBus = bus
	}
}

func (s *Service) CreateProject(ctx context.Context, principal authz.Principal, input CreateProjectInput) (Project, error) {
	if err := s.require(principal, authz.PermissionProjectsWrite, authz.GlobalScope()); err != nil {
		return Project{}, err
	}

	project, err := s.buildProject(principal, input)
	if err != nil {
		return Project{}, err
	}

	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		if err := s.requireUser(ctx, tx, "lead_user_id", project.LeadUserID); err != nil {
			return err
		}
		if err := s.requireUser(ctx, tx, "created_by", project.CreatedBy); err != nil {
			return err
		}
		if err := s.repo.insertProject(ctx, tx, project); err != nil {
			return err
		}
		if err := s.repo.bindProjectOwner(ctx, tx, project.ID, project.LeadUserID, project.CreatedAt); err != nil {
			return err
		}
		return s.seedDefaultProjectWorkflow(ctx, tx, project)
	}); err != nil {
		return Project{}, err
	}

	s.publish(ctx, events.Event{
		Type:      "project.created",
		ActorID:   actorID(principal),
		ProjectID: project.ID,
		ObjectID:  project.ID,
		At:        project.CreatedAt,
		Data: map[string]any{
			"key":  project.Key,
			"name": project.Name,
		},
	})

	return project, nil
}

func (s *Service) ListProjects(ctx context.Context, principal authz.Principal, input ListProjectsInput) ([]Project, error) {
	if err := validateListInput(input.Limit, input.Offset); err != nil {
		return nil, err
	}
	if s.authorizer == nil {
		return nil, errors.New("tracker: authorization evaluator is required")
	}

	projects, err := s.repo.ListProjects(ctx, input)
	if err != nil {
		return nil, err
	}

	visible := projects[:0]
	for _, project := range projects {
		if s.authorizer.Can(principal, authz.PermissionProjectsRead, authz.ProjectScope(project.ID)) {
			visible = append(visible, project)
		}
	}
	return visible, nil
}

func (s *Service) GetProject(ctx context.Context, principal authz.Principal, projectID string) (Project, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return Project{}, validationFailed(map[string]string{"project_id": "Required"})
	}

	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return Project{}, err
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(project.ID)); err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *Service) GetProjectByKey(ctx context.Context, principal authz.Principal, key string) (Project, error) {
	key = normalizeProjectKey(key)
	if key == "" {
		return Project{}, validationFailed(map[string]string{"key": "Required"})
	}

	project, err := s.repo.GetProjectByKey(ctx, key)
	if err != nil {
		return Project{}, err
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(project.ID)); err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *Service) CreateTicket(ctx context.Context, principal authz.Principal, input CreateTicketInput) (Ticket, error) {
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	if input.ProjectID == "" {
		return Ticket{}, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionTicketsWrite, authz.ProjectScope(input.ProjectID)); err != nil {
		return Ticket{}, err
	}

	var ticket Ticket
	var eventData map[string]any
	for attempt := 0; attempt < 3; attempt++ {
		created, data, err := s.createTicketOnce(ctx, principal, input)
		if err == nil {
			ticket = created
			eventData = data
			break
		}
		if attempt < 2 && isTicketKeyConflict(err) {
			continue
		}
		return Ticket{}, err
	}

	s.publish(ctx, events.Event{
		Type:      activityTicketCreated,
		ActorID:   actorID(principal),
		ProjectID: ticket.ProjectID,
		ObjectID:  ticket.ID,
		At:        ticket.CreatedAt,
		Data:      eventData,
	})

	return ticket, nil
}

func (s *Service) ListTickets(ctx context.Context, principal authz.Principal, input ListTicketsInput) ([]Ticket, error) {
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	if input.ProjectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}

	fields := map[string]string{}
	input.Status = normalizeSlug(input.Status)
	if input.Status != "" && !slugPattern.MatchString(input.Status) {
		fields["status"] = "Must be a lowercase slug"
	}
	input.AssigneeID = strings.TrimSpace(input.AssigneeID)
	input.Label = normalizeSlug(input.Label)
	if input.Label != "" && !slugPattern.MatchString(input.Label) {
		fields["label"] = "Must be a lowercase slug"
	}
	addListFieldErrors(fields, input.Limit, input.Offset)
	if len(fields) > 0 {
		return nil, validationFailed(fields)
	}

	if _, err := s.repo.GetProject(ctx, input.ProjectID); err != nil {
		return nil, err
	}
	tickets, err := s.repo.ListTickets(ctx, input)
	if err != nil {
		return nil, err
	}
	return s.attachTicketDetailsToTickets(ctx, tickets)
}

func (s *Service) GetTicket(ctx context.Context, principal authz.Principal, ticketID string) (Ticket, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return Ticket{}, validationFailed(map[string]string{"ticket_id": "Required"})
	}

	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(ticket.ProjectID)); err != nil {
		return Ticket{}, err
	}
	return s.attachTicketDetails(ctx, ticket)
}

func (s *Service) UpdateTicket(ctx context.Context, principal authz.Principal, ticketID string, input UpdateTicketInput) (Ticket, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return Ticket{}, validationFailed(map[string]string{"ticket_id": "Required"})
	}

	current, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return Ticket{}, err
	}
	if err := s.require(principal, authz.PermissionTicketsWrite, authz.ProjectScope(current.ProjectID)); err != nil {
		return Ticket{}, err
	}

	var updated Ticket
	var eventData map[string]any
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		current, err := s.repo.getTicket(ctx, tx, ticketID)
		if err != nil {
			return err
		}
		project, err := s.repo.getProject(ctx, tx, current.ProjectID)
		if err != nil {
			return err
		}
		if project.ArchivedAt != nil {
			return &ConflictError{Resource: "project", Field: "archived_at", Value: project.ID, Message: "project is archived"}
		}

		currentLabels, err := s.loadTicketLabelsFrom(ctx, tx, current.ID)
		if err != nil {
			return err
		}
		current.Labels = currentLabels

		candidate, changes, err := s.applyTicketUpdate(ctx, tx, current, input)
		if err != nil {
			return err
		}
		var customValues map[string]customFieldValue
		customChanged := input.CustomFields != nil
		if customChanged {
			customValues, err = s.validateCustomFieldValues(ctx, tx, current.ProjectID, *input.CustomFields, true)
			if err != nil {
				return err
			}
		}
		labelsChanged := input.Labels != nil && !equalStringSlices(current.Labels, candidate.Labels)
		if len(changes) == 0 && !customChanged && !labelsChanged {
			updated = current
			return nil
		}

		candidate.UpdatedAt = s.now().UTC()
		if err := s.repo.updateTicket(ctx, tx, candidate); err != nil {
			return err
		}
		if customChanged {
			if err := s.replaceTicketCustomFieldValues(ctx, tx, candidate.ID, customValues, candidate.UpdatedAt); err != nil {
				return err
			}
			candidate.CustomFields = customFieldValueMap(customValues)
		}
		if labelsChanged {
			if err := s.replaceTicketLabels(ctx, tx, candidate.ID, candidate.Labels, candidate.UpdatedAt); err != nil {
				return err
			}
		}

		eventData = map[string]any{"changes": changes}
		if customChanged {
			eventData["custom_fields"] = "updated"
		}
		if labelsChanged {
			eventData["labels"] = candidate.Labels
		}
		activityID, err := newID("activity")
		if err != nil {
			return err
		}
		if err := s.repo.insertTicketActivity(ctx, tx, TicketActivity{
			ID:           activityID,
			TicketID:     candidate.ID,
			ActorID:      actorID(principal),
			ActivityType: activityTicketUpdated,
			Data:         eventData,
			CreatedAt:    candidate.UpdatedAt,
		}); err != nil {
			return err
		}

		updated = candidate
		return nil
	}); err != nil {
		return Ticket{}, err
	}

	if eventData != nil {
		s.publish(ctx, events.Event{
			Type:      activityTicketUpdated,
			ActorID:   actorID(principal),
			ProjectID: updated.ProjectID,
			ObjectID:  updated.ID,
			At:        updated.UpdatedAt,
			Data:      eventData,
		})
	}

	return s.attachTicketDetails(ctx, updated)
}

func (s *Service) ListTicketActivity(ctx context.Context, principal authz.Principal, ticketID string) ([]TicketActivity, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return nil, validationFailed(map[string]string{"ticket_id": "Required"})
	}

	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(ticket.ProjectID)); err != nil {
		return nil, err
	}
	return s.repo.ListTicketActivity(ctx, ticketID)
}

func (s *Service) buildProject(principal authz.Principal, input CreateProjectInput) (Project, error) {
	fields := map[string]string{}

	key := normalizeProjectKey(input.Key)
	if key == "" {
		fields["key"] = "Required"
	} else if !projectKeyPattern.MatchString(key) {
		fields["key"] = "Must start with a letter and contain only uppercase letters or digits"
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		fields["name"] = "Required"
	} else if len(name) > 200 {
		fields["name"] = "Must be 200 characters or fewer"
	}

	description := strings.TrimSpace(input.Description)
	if len(description) > 8000 {
		fields["description"] = "Must be 8000 characters or fewer"
	}

	leadUserID := strings.TrimSpace(input.LeadUserID)
	if leadUserID == "" {
		leadUserID = principal.UserID
	}
	if leadUserID == "" {
		fields["lead_user_id"] = "Required"
	}

	createdBy := actorID(principal)
	if createdBy == "" {
		fields["created_by"] = "Required"
	}

	if len(fields) > 0 {
		return Project{}, validationFailed(fields)
	}

	id, err := newID("project")
	if err != nil {
		return Project{}, err
	}
	now := s.now().UTC()
	return Project{
		ID:          id,
		Key:         key,
		Name:        name,
		Description: description,
		LeadUserID:  leadUserID,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Service) createTicketOnce(ctx context.Context, principal authz.Principal, input CreateTicketInput) (Ticket, map[string]any, error) {
	var created Ticket
	var eventData map[string]any

	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		project, err := s.repo.getProject(ctx, tx, input.ProjectID)
		if err != nil {
			return err
		}
		if project.ArchivedAt != nil {
			return &ConflictError{Resource: "project", Field: "archived_at", Value: project.ID, Message: "project is archived"}
		}

		ticket, err := s.buildTicket(ctx, tx, principal, project, input)
		if err != nil {
			return err
		}
		customValues, err := s.validateCustomFieldValues(ctx, tx, project.ID, input.CustomFields, true)
		if err != nil {
			return err
		}
		key, err := s.repo.nextTicketKey(ctx, tx, project)
		if err != nil {
			return err
		}
		ticket.Key = key

		if err := s.repo.insertTicket(ctx, tx, ticket); err != nil {
			return err
		}
		if err := s.replaceTicketLabels(ctx, tx, ticket.ID, ticket.Labels, ticket.CreatedAt); err != nil {
			return err
		}
		if err := s.replaceTicketCustomFieldValues(ctx, tx, ticket.ID, customValues, ticket.CreatedAt); err != nil {
			return err
		}
		ticket.CustomFields = customFieldValueMap(customValues)

		eventData = map[string]any{
			"key":    ticket.Key,
			"title":  ticket.Title,
			"status": ticket.Status,
		}
		if len(ticket.Labels) > 0 {
			eventData["labels"] = ticket.Labels
		}
		activityID, err := newID("activity")
		if err != nil {
			return err
		}
		if err := s.repo.insertTicketActivity(ctx, tx, TicketActivity{
			ID:           activityID,
			TicketID:     ticket.ID,
			ActorID:      actorID(principal),
			ActivityType: activityTicketCreated,
			Data:         eventData,
			CreatedAt:    ticket.CreatedAt,
		}); err != nil {
			return err
		}

		created = ticket
		return nil
	}); err != nil {
		return Ticket{}, nil, err
	}

	return created, eventData, nil
}

func (s *Service) buildTicket(ctx context.Context, tx *sql.Tx, principal authz.Principal, project Project, input CreateTicketInput) (Ticket, error) {
	fields := map[string]string{}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		fields["title"] = "Required"
	} else if len(title) > 500 {
		fields["title"] = "Must be 500 characters or fewer"
	}

	description := strings.TrimSpace(input.Description)
	if len(description) > 20000 {
		fields["description"] = "Must be 20000 characters or fewer"
	}

	status := normalizeSlug(input.Status)
	if status == "" {
		status = defaultTicketStatus
	}
	validateSlugField(fields, "status", status, true)

	priority := normalizeSlug(input.Priority)
	validateSlugField(fields, "priority", priority, false)

	ticketType := normalizeSlug(input.Type)
	validateSlugField(fields, "type", ticketType, false)

	reporterID := strings.TrimSpace(input.ReporterID)
	if reporterID == "" {
		reporterID = principal.UserID
	}
	if reporterID == "" {
		fields["reporter_id"] = "Required"
	}

	assigneeID := strings.TrimSpace(input.AssigneeID)
	parentTicketID := strings.TrimSpace(input.ParentTicketID)
	sprintID := strings.TrimSpace(input.SprintID)
	componentID := strings.TrimSpace(input.ComponentID)
	versionID := strings.TrimSpace(input.VersionID)
	rank := strings.TrimSpace(input.Rank)
	startDate := strings.TrimSpace(input.StartDate)
	dueDate := strings.TrimSpace(input.DueDate)
	if len(rank) > 200 {
		fields["rank"] = "Must be 200 characters or fewer"
	}
	validateTicketDates(fields, startDate, dueDate)
	labels, err := normalizeTicketLabels(input.Labels)
	if err != nil {
		return Ticket{}, err
	}

	if len(fields) > 0 {
		return Ticket{}, validationFailed(fields)
	}

	if err := s.requireUser(ctx, tx, "reporter_id", reporterID); err != nil {
		return Ticket{}, err
	}
	if err := s.requireUser(ctx, tx, "assignee_id", assigneeID); err != nil {
		return Ticket{}, err
	}
	if err := s.requireParentTicket(ctx, tx, "parent_ticket_id", parentTicketID, project.ID, ""); err != nil {
		return Ticket{}, err
	}
	if err := s.requireSprint(ctx, tx, "sprint_id", sprintID, project.ID); err != nil {
		return Ticket{}, err
	}
	if err := s.requireComponent(ctx, tx, "component_id", componentID, project.ID); err != nil {
		return Ticket{}, err
	}
	if err := s.requireVersion(ctx, tx, "version_id", versionID, project.ID); err != nil {
		return Ticket{}, err
	}

	id, err := newID("ticket")
	if err != nil {
		return Ticket{}, err
	}
	now := s.now().UTC()
	return Ticket{
		ID:             id,
		ProjectID:      project.ID,
		Title:          title,
		Description:    description,
		Status:         status,
		Priority:       priority,
		Type:           ticketType,
		ReporterID:     reporterID,
		AssigneeID:     assigneeID,
		ParentTicketID: parentTicketID,
		SprintID:       sprintID,
		ComponentID:    componentID,
		VersionID:      versionID,
		Rank:           rank,
		StartDate:      startDate,
		DueDate:        dueDate,
		Labels:         labels,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

type ticketFieldChange struct {
	Old string `json:"old"`
	New string `json:"new"`
}

func (s *Service) applyTicketUpdate(ctx context.Context, tx *sql.Tx, current Ticket, input UpdateTicketInput) (Ticket, map[string]ticketFieldChange, error) {
	next := current
	fields := map[string]string{}
	changes := map[string]ticketFieldChange{}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			fields["title"] = "Required"
		} else if len(title) > 500 {
			fields["title"] = "Must be 500 characters or fewer"
		} else {
			next.Title = title
		}
	}

	if input.Description != nil {
		description := strings.TrimSpace(*input.Description)
		if len(description) > 20000 {
			fields["description"] = "Must be 20000 characters or fewer"
		} else {
			next.Description = description
		}
	}

	if input.Status != nil {
		status := normalizeSlug(*input.Status)
		if status == "" {
			fields["status"] = "Required"
		} else {
			validateSlugField(fields, "status", status, true)
			next.Status = status
		}
	}

	if input.Priority != nil {
		priority := normalizeSlug(*input.Priority)
		validateSlugField(fields, "priority", priority, false)
		next.Priority = priority
	}

	if input.Type != nil {
		ticketType := normalizeSlug(*input.Type)
		validateSlugField(fields, "type", ticketType, false)
		next.Type = ticketType
	}

	if input.AssigneeID != nil {
		next.AssigneeID = strings.TrimSpace(*input.AssigneeID)
	}

	if input.ParentTicketID != nil {
		next.ParentTicketID = strings.TrimSpace(*input.ParentTicketID)
	}

	if input.SprintID != nil {
		next.SprintID = strings.TrimSpace(*input.SprintID)
	}

	if input.ComponentID != nil {
		next.ComponentID = strings.TrimSpace(*input.ComponentID)
	}

	if input.VersionID != nil {
		next.VersionID = strings.TrimSpace(*input.VersionID)
	}

	if input.Rank != nil {
		rank := strings.TrimSpace(*input.Rank)
		if len(rank) > 200 {
			fields["rank"] = "Must be 200 characters or fewer"
		} else {
			next.Rank = rank
		}
	}

	if input.StartDate != nil {
		next.StartDate = strings.TrimSpace(*input.StartDate)
	}

	if input.DueDate != nil {
		next.DueDate = strings.TrimSpace(*input.DueDate)
	}
	if input.Labels != nil {
		labels, err := normalizeTicketLabels(*input.Labels)
		if err != nil {
			return Ticket{}, nil, err
		}
		next.Labels = labels
	}
	validateTicketDates(fields, next.StartDate, next.DueDate)

	if len(fields) > 0 {
		return Ticket{}, nil, validationFailed(fields)
	}

	if err := s.requireUser(ctx, tx, "assignee_id", next.AssigneeID); err != nil {
		return Ticket{}, nil, err
	}
	if err := s.requireParentTicket(ctx, tx, "parent_ticket_id", next.ParentTicketID, current.ProjectID, current.ID); err != nil {
		return Ticket{}, nil, err
	}
	if err := s.requireSprint(ctx, tx, "sprint_id", next.SprintID, current.ProjectID); err != nil {
		return Ticket{}, nil, err
	}
	if err := s.requireComponent(ctx, tx, "component_id", next.ComponentID, current.ProjectID); err != nil {
		return Ticket{}, nil, err
	}
	if err := s.requireVersion(ctx, tx, "version_id", next.VersionID, current.ProjectID); err != nil {
		return Ticket{}, nil, err
	}

	addChange(changes, "title", current.Title, next.Title)
	addChange(changes, "description", current.Description, next.Description)
	addChange(changes, "status", current.Status, next.Status)
	addChange(changes, "priority", current.Priority, next.Priority)
	addChange(changes, "type", current.Type, next.Type)
	addChange(changes, "assignee_id", current.AssigneeID, next.AssigneeID)
	addChange(changes, "parent_ticket_id", current.ParentTicketID, next.ParentTicketID)
	addChange(changes, "sprint_id", current.SprintID, next.SprintID)
	addChange(changes, "component_id", current.ComponentID, next.ComponentID)
	addChange(changes, "version_id", current.VersionID, next.VersionID)
	addChange(changes, "rank", current.Rank, next.Rank)
	addChange(changes, "start_date", current.StartDate, next.StartDate)
	addChange(changes, "due_date", current.DueDate, next.DueDate)
	if input.Labels != nil && !equalStringSlices(current.Labels, next.Labels) {
		changes["labels"] = ticketFieldChange{Old: strings.Join(current.Labels, ","), New: strings.Join(next.Labels, ",")}
	}

	return next, changes, nil
}

func validateTicketDates(fields map[string]string, startDate string, dueDate string) {
	validateTicketDate(fields, "start_date", startDate)
	validateTicketDate(fields, "due_date", dueDate)
	if strings.TrimSpace(startDate) == "" || strings.TrimSpace(dueDate) == "" {
		return
	}
	start, startErr := time.Parse(dateOnlyLayout, strings.TrimSpace(startDate))
	due, dueErr := time.Parse(dateOnlyLayout, strings.TrimSpace(dueDate))
	if startErr == nil && dueErr == nil && due.Before(start) {
		fields["due_date"] = "Must be on or after start_date"
	}
}

func validateTicketDate(fields map[string]string, field string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if _, err := time.Parse(dateOnlyLayout, value); err != nil {
		fields[field] = "Must use YYYY-MM-DD"
	}
}

func (s *Service) requireComponent(ctx context.Context, q sqlRunner, field string, componentID string, projectID string) error {
	if strings.TrimSpace(componentID) == "" {
		return nil
	}
	var exists int
	if err := q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM project_components
		WHERE id = ? AND project_id = ?
	`, componentID, projectID).Scan(&exists); err != nil {
		return fmt.Errorf("check component: %w", err)
	}
	if exists != 1 {
		return validationFailed(map[string]string{field: "Component not found in project"})
	}
	return nil
}

func (s *Service) requireVersion(ctx context.Context, q sqlRunner, field string, versionID string, projectID string) error {
	if strings.TrimSpace(versionID) == "" {
		return nil
	}
	var exists int
	if err := q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM project_versions
		WHERE id = ? AND project_id = ?
	`, versionID, projectID).Scan(&exists); err != nil {
		return fmt.Errorf("check version: %w", err)
	}
	if exists != 1 {
		return validationFailed(map[string]string{field: "Version not found in project"})
	}
	return nil
}

func (s *Service) requireSprint(ctx context.Context, q sqlRunner, field string, sprintID string, projectID string) error {
	if strings.TrimSpace(sprintID) == "" {
		return nil
	}
	var exists int
	if err := q.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sprints
		WHERE id = ? AND project_id = ?
	`, sprintID, projectID).Scan(&exists); err != nil {
		return fmt.Errorf("check sprint: %w", err)
	}
	if exists != 1 {
		return validationFailed(map[string]string{field: "Sprint not found in project"})
	}
	return nil
}

func (s *Service) requireUser(ctx context.Context, q sqlRunner, field string, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return nil
	}
	exists, err := s.repo.userExists(ctx, q, userID)
	if err != nil {
		return err
	}
	if !exists {
		return validationFailed(map[string]string{field: "User not found"})
	}
	return nil
}

func (s *Service) requireParentTicket(ctx context.Context, q sqlRunner, field string, ticketID string, projectID string, currentTicketID string) error {
	if strings.TrimSpace(ticketID) == "" {
		return nil
	}
	if ticketID == currentTicketID {
		return validationFailed(map[string]string{field: "Ticket cannot be its own parent"})
	}
	exists, err := s.repo.epicExistsInProject(ctx, q, ticketID, projectID)
	if err != nil {
		return err
	}
	if !exists {
		return validationFailed(map[string]string{field: "Epic not found in project"})
	}
	return nil
}

func (s *Service) require(principal authz.Principal, permission authz.Permission, scope authz.Scope) error {
	if s == nil || s.authorizer == nil {
		return errors.New("tracker: authorization evaluator is required")
	}
	return s.authorizer.Require(principal, permission, scope)
}

func (s *Service) withTx(ctx context.Context, fn func(*sql.Tx) error) (err error) {
	if s == nil || s.db == nil {
		return errors.New("tracker: nil database")
	}
	if fn == nil {
		return errors.New("tracker: transaction function is required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tracker transaction: %w", err)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback()
			panic(recovered)
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("commit tracker transaction: %w", commitErr)
		}
	}()

	err = fn(tx)
	return err
}

func (s *Service) publish(ctx context.Context, event events.Event) {
	if s == nil || s.eventBus == nil {
		return
	}
	_ = s.eventBus.Publish(ctx, event)
}

func validateListInput(limit int, offset int) error {
	fields := map[string]string{}
	addListFieldErrors(fields, limit, offset)
	if len(fields) > 0 {
		return validationFailed(fields)
	}
	return nil
}

func addListFieldErrors(fields map[string]string, limit int, offset int) {
	if limit < 0 {
		fields["limit"] = "Must be non-negative"
	}
	if limit > maxListLimit {
		fields["limit"] = "Must be 200 or fewer"
	}
	if offset < 0 {
		fields["offset"] = "Must be non-negative"
	}
}

func validateSlugField(fields map[string]string, field string, value string, required bool) {
	if value == "" {
		if required {
			fields[field] = "Required"
		}
		return
	}
	if !slugPattern.MatchString(value) {
		fields[field] = "Must be a lowercase slug"
	}
}

func addChange(changes map[string]ticketFieldChange, field string, oldValue string, newValue string) {
	if oldValue == newValue {
		return
	}
	changes[field] = ticketFieldChange{Old: oldValue, New: newValue}
}

func actorID(principal authz.Principal) string {
	if principal.ActorUserID != "" {
		return principal.ActorUserID
	}
	return principal.UserID
}

func normalizeProjectKey(key string) string {
	return strings.ToUpper(strings.TrimSpace(key))
}

func normalizeSlug(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isTicketKeyConflict(err error) bool {
	var conflictErr *ConflictError
	return errors.As(err, &conflictErr) && conflictErr.Resource == "ticket" && conflictErr.Field == "key"
}

func newID(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(raw[:]), nil
}
