package tracker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

type CreatePage struct {
	ID           string
	ProjectID    string
	Name         string
	Slug         string
	Description  string
	Enabled      bool
	TargetType   string
	TargetStatus string
	FieldLayout  []map[string]any
	Defaults     map[string]any
	OwnerUserID  string
	CreatedBy    string
	UpdatedBy    string
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ListCreatePagesInput struct {
	ProjectID       string
	IncludeDisabled bool
	Limit           int
	Offset          int
}

type CreateCreatePageInput struct {
	ProjectID    string
	Name         string
	Slug         string
	Description  string
	Enabled      bool
	TargetType   string
	TargetStatus string
	FieldLayout  []map[string]any
	Defaults     map[string]any
	OwnerUserID  string
}

type UpdateCreatePageInput struct {
	Name         *string
	Slug         *string
	Description  *string
	Enabled      *bool
	TargetType   *string
	TargetStatus *string
	FieldLayout  *[]map[string]any
	Defaults     *map[string]any
	OwnerUserID  *string
}

type SubmitCreatePageInput struct {
	Ticket CreateTicketInput
}

type CreatePageService struct {
	db         *sql.DB
	tracker    *Service
	authorizer authz.Evaluator
	now        func() time.Time
}

func NewCreatePageService(db *sql.DB, trackerService *Service, authorizer authz.Evaluator) *CreatePageService {
	return &CreatePageService{
		db:         db,
		tracker:    trackerService,
		authorizer: authorizer,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func (s *CreatePageService) List(ctx context.Context, principal authz.Principal, input ListCreatePagesInput) ([]CreatePage, error) {
	projectID := strings.TrimSpace(input.ProjectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := validateListInput(input.Limit, input.Offset); err != nil {
		return nil, err
	}
	if err := s.requireManage(principal, projectID); err != nil {
		return nil, err
	}
	limit, offset := normalizeListWindow(input.Limit, input.Offset)
	where := []string{"project_id = ?", "deleted_at IS NULL"}
	args := []any{projectID}
	if !input.IncludeDisabled {
		where = append(where, "enabled = 1")
	}
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, name, slug, COALESCE(description, ''), enabled,
			COALESCE(target_type, ''), COALESCE(target_status, ''), field_layout_json,
			defaults_json, COALESCE(owner_user_id, ''), COALESCE(created_by, ''),
			COALESCE(updated_by, ''), deleted_at, created_at, updated_at
		FROM ticket_create_pages
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY slug ASC, id ASC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list ticket create pages: %w", err)
	}
	defer rows.Close()

	var pages []CreatePage
	for rows.Next() {
		page, err := scanCreatePage(rows)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket create pages: %w", err)
	}
	return pages, nil
}

func (s *CreatePageService) Create(ctx context.Context, principal authz.Principal, input CreateCreatePageInput) (CreatePage, error) {
	page, err := s.buildCreatePage(principal, input)
	if err != nil {
		return CreatePage{}, err
	}
	if err := s.requireManage(principal, page.ProjectID); err != nil {
		return CreatePage{}, err
	}
	if err := s.validate(ctx, page); err != nil {
		return CreatePage{}, err
	}
	layout, defaults, err := createPageJSON(page)
	if err != nil {
		return CreatePage{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO ticket_create_pages (
			id, project_id, name, slug, description, enabled, target_type, target_status,
			field_layout_json, defaults_json, owner_user_id, created_by, updated_by, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, page.ID, page.ProjectID, page.Name, page.Slug, nullableString(page.Description), page.Enabled,
		nullableString(page.TargetType), nullableString(page.TargetStatus), layout, defaults,
		nullableString(page.OwnerUserID), nullableString(page.CreatedBy), nullableString(page.UpdatedBy),
		formatTime(page.CreatedAt), formatTime(page.UpdatedAt)); err != nil {
		if isUniqueConstraint(err) {
			return CreatePage{}, conflict("ticket_create_page", "slug", page.Slug)
		}
		return CreatePage{}, fmt.Errorf("insert ticket create page: %w", err)
	}
	return page, nil
}

func (s *CreatePageService) Get(ctx context.Context, principal authz.Principal, pageID string) (CreatePage, error) {
	page, err := s.get(ctx, pageID)
	if err != nil {
		return CreatePage{}, err
	}
	if err := s.requireManage(principal, page.ProjectID); err != nil {
		return CreatePage{}, err
	}
	return page, nil
}

func (s *CreatePageService) Resolve(ctx context.Context, principal authz.Principal, projectID string, slug string) (CreatePage, error) {
	page, err := s.getBySlug(ctx, projectID, slug)
	if err != nil {
		return CreatePage{}, err
	}
	if !page.Enabled {
		return CreatePage{}, notFound("ticket_create_page", slug)
	}
	if err := s.requireProjectRead(principal, page.ProjectID); err != nil {
		return CreatePage{}, err
	}
	return page, nil
}

func (s *CreatePageService) Update(ctx context.Context, principal authz.Principal, pageID string, input UpdateCreatePageInput) (CreatePage, error) {
	current, err := s.get(ctx, pageID)
	if err != nil {
		return CreatePage{}, err
	}
	if err := s.requireManage(principal, current.ProjectID); err != nil {
		return CreatePage{}, err
	}
	updated := current
	if input.Name != nil {
		updated.Name = strings.TrimSpace(*input.Name)
	}
	if input.Slug != nil {
		updated.Slug = normalizeSlug(*input.Slug)
	}
	if input.Description != nil {
		updated.Description = strings.TrimSpace(*input.Description)
	}
	if input.Enabled != nil {
		updated.Enabled = *input.Enabled
	}
	if input.TargetType != nil {
		updated.TargetType = normalizeSlug(*input.TargetType)
	}
	if input.TargetStatus != nil {
		updated.TargetStatus = normalizeSlug(*input.TargetStatus)
	}
	if input.FieldLayout != nil {
		updated.FieldLayout = *input.FieldLayout
	}
	if input.Defaults != nil {
		updated.Defaults = *input.Defaults
	}
	if input.OwnerUserID != nil {
		updated.OwnerUserID = strings.TrimSpace(*input.OwnerUserID)
	}
	updated.UpdatedBy = actorID(principal)
	updated.UpdatedAt = s.now().UTC()
	if err := s.validate(ctx, updated); err != nil {
		return CreatePage{}, err
	}
	layout, defaults, err := createPageJSON(updated)
	if err != nil {
		return CreatePage{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE ticket_create_pages
		SET name = ?, slug = ?, description = ?, enabled = ?, target_type = ?, target_status = ?,
			field_layout_json = ?, defaults_json = ?, owner_user_id = ?, updated_by = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, updated.Name, updated.Slug, nullableString(updated.Description), updated.Enabled,
		nullableString(updated.TargetType), nullableString(updated.TargetStatus), layout, defaults,
		nullableString(updated.OwnerUserID), nullableString(updated.UpdatedBy), formatTime(updated.UpdatedAt), updated.ID)
	if err != nil {
		if isUniqueConstraint(err) {
			return CreatePage{}, conflict("ticket_create_page", "slug", updated.Slug)
		}
		return CreatePage{}, fmt.Errorf("update ticket create page: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return CreatePage{}, fmt.Errorf("check ticket create page update: %w", err)
	}
	if affected == 0 {
		return CreatePage{}, notFound("ticket_create_page", pageID)
	}
	return updated, nil
}

func (s *CreatePageService) Delete(ctx context.Context, principal authz.Principal, pageID string) error {
	page, err := s.get(ctx, pageID)
	if err != nil {
		return err
	}
	if err := s.requireManage(principal, page.ProjectID); err != nil {
		return err
	}
	deletedAt := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE ticket_create_pages
		SET enabled = 0, deleted_at = ?, updated_by = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatTime(deletedAt), nullableString(actorID(principal)), formatTime(deletedAt), page.ID)
	if err != nil {
		return fmt.Errorf("delete ticket create page: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check ticket create page delete: %w", err)
	}
	if affected == 0 {
		return notFound("ticket_create_page", pageID)
	}
	return nil
}

func (s *CreatePageService) Submit(ctx context.Context, principal authz.Principal, projectID string, slug string, input SubmitCreatePageInput) (Ticket, error) {
	page, err := s.Resolve(ctx, principal, projectID, slug)
	if err != nil {
		return Ticket{}, err
	}
	ticketInput := createTicketInputFromCreatePage(page, input.Ticket)
	return s.tracker.CreateTicket(ctx, principal, ticketInput)
}

func (s *CreatePageService) buildCreatePage(principal authz.Principal, input CreateCreatePageInput) (CreatePage, error) {
	id, err := newID("create_page")
	if err != nil {
		return CreatePage{}, err
	}
	now := s.now().UTC()
	return CreatePage{
		ID:           id,
		ProjectID:    strings.TrimSpace(input.ProjectID),
		Name:         strings.TrimSpace(input.Name),
		Slug:         normalizeSlug(input.Slug),
		Description:  strings.TrimSpace(input.Description),
		Enabled:      input.Enabled,
		TargetType:   normalizeSlug(input.TargetType),
		TargetStatus: normalizeSlug(input.TargetStatus),
		FieldLayout:  normalizeCreatePageLayout(input.FieldLayout),
		Defaults:     normalizeCreatePageDefaults(input.Defaults),
		OwnerUserID:  strings.TrimSpace(input.OwnerUserID),
		CreatedBy:    actorID(principal),
		UpdatedBy:    actorID(principal),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (s *CreatePageService) validate(ctx context.Context, page CreatePage) error {
	fields := map[string]string{}
	if page.ProjectID == "" {
		fields["project_id"] = "Required"
	} else {
		var exists bool
		if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE id = ? AND deleted_at IS NULL)", page.ProjectID).Scan(&exists); err != nil {
			return fmt.Errorf("check ticket create page project: %w", err)
		}
		if !exists {
			fields["project_id"] = "Project not found"
		}
	}
	if page.Name == "" {
		fields["name"] = "Required"
	}
	validateSlugField(fields, "slug", page.Slug, true)
	if page.TargetStatus != "" && !slugPattern.MatchString(page.TargetStatus) {
		fields["target_status"] = "Must be a lowercase slug"
	}
	if page.OwnerUserID != "" {
		var exists bool
		if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL)", page.OwnerUserID).Scan(&exists); err != nil {
			return fmt.Errorf("check ticket create page owner: %w", err)
		}
		if !exists {
			fields["owner_user_id"] = "User not found"
		}
	}
	if _, _, err := createPageJSON(page); err != nil {
		fields["schema"] = err.Error()
	}
	if len(fields) > 0 {
		return validationFailed(fields)
	}
	return nil
}

func (s *CreatePageService) get(ctx context.Context, pageID string) (CreatePage, error) {
	pageID = strings.TrimSpace(pageID)
	if pageID == "" {
		return CreatePage{}, validationFailed(map[string]string{"page_id": "Required"})
	}
	page, err := scanCreatePage(s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, slug, COALESCE(description, ''), enabled,
			COALESCE(target_type, ''), COALESCE(target_status, ''), field_layout_json,
			defaults_json, COALESCE(owner_user_id, ''), COALESCE(created_by, ''),
			COALESCE(updated_by, ''), deleted_at, created_at, updated_at
		FROM ticket_create_pages
		WHERE id = ? AND deleted_at IS NULL
	`, pageID))
	if errors.Is(err, sql.ErrNoRows) {
		return CreatePage{}, notFound("ticket_create_page", pageID)
	}
	if err != nil {
		return CreatePage{}, fmt.Errorf("get ticket create page: %w", err)
	}
	return page, nil
}

func (s *CreatePageService) getBySlug(ctx context.Context, projectID string, slug string) (CreatePage, error) {
	projectID = strings.TrimSpace(projectID)
	slug = normalizeSlug(slug)
	if projectID == "" || slug == "" {
		return CreatePage{}, validationFailed(map[string]string{"slug": "Required"})
	}
	page, err := scanCreatePage(s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, slug, COALESCE(description, ''), enabled,
			COALESCE(target_type, ''), COALESCE(target_status, ''), field_layout_json,
			defaults_json, COALESCE(owner_user_id, ''), COALESCE(created_by, ''),
			COALESCE(updated_by, ''), deleted_at, created_at, updated_at
		FROM ticket_create_pages
		WHERE project_id = ? AND slug = ? AND deleted_at IS NULL
	`, projectID, slug))
	if errors.Is(err, sql.ErrNoRows) {
		return CreatePage{}, notFound("ticket_create_page", slug)
	}
	if err != nil {
		return CreatePage{}, fmt.Errorf("resolve ticket create page: %w", err)
	}
	return page, nil
}

func (s *CreatePageService) requireManage(principal authz.Principal, projectID string) error {
	if s == nil || s.authorizer == nil {
		return errors.New("tracker: create page authorization evaluator is required")
	}
	return s.authorizer.Require(principal, authz.PermissionAutomationsManage, authz.ProjectScope(projectID))
}

func (s *CreatePageService) requireProjectRead(principal authz.Principal, projectID string) error {
	if s == nil || s.authorizer == nil {
		return errors.New("tracker: create page authorization evaluator is required")
	}
	return s.authorizer.Require(principal, authz.PermissionProjectsRead, authz.ProjectScope(projectID))
}

func createTicketInputFromCreatePage(page CreatePage, submitted CreateTicketInput) CreateTicketInput {
	merged := createTicketInputFromDefaults(page.ProjectID, page.Defaults)
	if page.TargetType != "" {
		merged.Type = page.TargetType
	}
	if page.TargetStatus != "" {
		merged.Status = page.TargetStatus
	}
	mergeCreateTicketInput(&merged, submitted)
	merged.ProjectID = page.ProjectID
	return merged
}

func createTicketInputFromDefaults(projectID string, defaults map[string]any) CreateTicketInput {
	return createTicketInputFromMap(defaults, CreateTicketInput{ProjectID: projectID})
}

func mergeCreateTicketInput(target *CreateTicketInput, source CreateTicketInput) {
	if source.Title != "" {
		target.Title = source.Title
	}
	if source.Description != "" {
		target.Description = source.Description
	}
	if source.Status != "" {
		target.Status = source.Status
	}
	if source.Priority != "" {
		target.Priority = source.Priority
	}
	if source.Type != "" {
		target.Type = source.Type
	}
	if source.ReporterID != "" {
		target.ReporterID = source.ReporterID
	}
	if source.AssigneeID != "" {
		target.AssigneeID = source.AssigneeID
	}
	if source.ParentTicketID != "" {
		target.ParentTicketID = source.ParentTicketID
	}
	if source.SprintID != "" {
		target.SprintID = source.SprintID
	}
	if source.ComponentID != "" {
		target.ComponentID = source.ComponentID
	}
	if source.VersionID != "" {
		target.VersionID = source.VersionID
	}
	if source.Rank != "" {
		target.Rank = source.Rank
	}
	if source.StartDate != "" {
		target.StartDate = source.StartDate
	}
	if source.DueDate != "" {
		target.DueDate = source.DueDate
	}
	if source.Labels != nil {
		target.Labels = source.Labels
	}
	if source.CustomFields != nil {
		target.CustomFields = source.CustomFields
	}
}

func scanCreatePage(scanner interface{ Scan(...any) error }) (CreatePage, error) {
	var page CreatePage
	var deletedAt sql.NullString
	var createdAt string
	var updatedAt string
	var layout string
	var defaults string
	if err := scanner.Scan(
		&page.ID,
		&page.ProjectID,
		&page.Name,
		&page.Slug,
		&page.Description,
		&page.Enabled,
		&page.TargetType,
		&page.TargetStatus,
		&layout,
		&defaults,
		&page.OwnerUserID,
		&page.CreatedBy,
		&page.UpdatedBy,
		&deletedAt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return CreatePage{}, err
	}
	if err := json.Unmarshal([]byte(layout), &page.FieldLayout); err != nil {
		return CreatePage{}, fmt.Errorf("decode ticket create page field layout: %w", err)
	}
	if err := json.Unmarshal([]byte(defaults), &page.Defaults); err != nil {
		return CreatePage{}, fmt.Errorf("decode ticket create page defaults: %w", err)
	}
	created, err := parseTime(createdAt)
	if err != nil {
		return CreatePage{}, fmt.Errorf("parse ticket create page created time: %w", err)
	}
	updated, err := parseTime(updatedAt)
	if err != nil {
		return CreatePage{}, fmt.Errorf("parse ticket create page updated time: %w", err)
	}
	page.CreatedAt = created
	page.UpdatedAt = updated
	page.DeletedAt = parseNullableTime(deletedAt)
	return page, nil
}

func createPageJSON(page CreatePage) (string, string, error) {
	layout, err := json.Marshal(normalizeCreatePageLayout(page.FieldLayout))
	if err != nil {
		return "", "", err
	}
	defaults, err := json.Marshal(normalizeCreatePageDefaults(page.Defaults))
	if err != nil {
		return "", "", err
	}
	return string(layout), string(defaults), nil
}

func normalizeCreatePageLayout(layout []map[string]any) []map[string]any {
	if layout == nil {
		return []map[string]any{}
	}
	return layout
}

func normalizeCreatePageDefaults(defaults map[string]any) map[string]any {
	if defaults == nil {
		return map[string]any{}
	}
	return defaults
}
