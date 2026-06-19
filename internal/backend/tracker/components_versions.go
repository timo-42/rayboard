package tracker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/timo-42/rayboard/internal/backend/authz"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	VersionStatusPlanned  = "planned"
	VersionStatusReleased = "released"
	VersionStatusArchived = "archived"
)

func (s *Service) ListComponents(ctx context.Context, principal authz.Principal, projectID string) ([]Component, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, name, description, owner_user_id, default_assignee_id, created_at, updated_at
		FROM project_components
		WHERE project_id = ?
		ORDER BY name ASC, id ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list components: %w", err)
	}
	defer rows.Close()
	var components []Component
	for rows.Next() {
		component, err := scanComponent(rows)
		if err != nil {
			return nil, err
		}
		components = append(components, component)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate components: %w", err)
	}
	return components, nil
}

func (s *Service) CreateComponent(ctx context.Context, principal authz.Principal, input CreateComponentInput) (Component, error) {
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	if input.ProjectID == "" {
		return Component{}, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionProjectsWrite, authz.ProjectScope(input.ProjectID)); err != nil {
		return Component{}, err
	}
	if _, err := s.repo.GetProject(ctx, input.ProjectID); err != nil {
		return Component{}, err
	}
	component, err := s.buildComponent(ctx, input)
	if err != nil {
		return Component{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO project_components (
			id, project_id, name, description, owner_user_id, default_assignee_id, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, component.ID, component.ProjectID, component.Name, nullableString(component.Description), nullableString(component.OwnerUserID), nullableString(component.DefaultAssigneeID), formatTime(component.CreatedAt), formatTime(component.UpdatedAt)); err != nil {
		if isUniqueConstraint(err) {
			return Component{}, conflict("component", "name", component.Name)
		}
		return Component{}, fmt.Errorf("insert component: %w", err)
	}
	return component, nil
}

func (s *Service) GetComponent(ctx context.Context, principal authz.Principal, componentID string) (Component, error) {
	component, err := s.getComponent(ctx, componentID)
	if err != nil {
		return Component{}, err
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(component.ProjectID)); err != nil {
		return Component{}, err
	}
	return component, nil
}

func (s *Service) UpdateComponent(ctx context.Context, principal authz.Principal, componentID string, input UpdateComponentInput) (Component, error) {
	current, err := s.getComponent(ctx, componentID)
	if err != nil {
		return Component{}, err
	}
	if err := s.require(principal, authz.PermissionProjectsWrite, authz.ProjectScope(current.ProjectID)); err != nil {
		return Component{}, err
	}
	updated := current
	if input.Name != nil {
		updated.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		updated.Description = strings.TrimSpace(*input.Description)
	}
	if input.OwnerUserID != nil {
		updated.OwnerUserID = strings.TrimSpace(*input.OwnerUserID)
	}
	if input.DefaultAssigneeID != nil {
		updated.DefaultAssigneeID = strings.TrimSpace(*input.DefaultAssigneeID)
	}
	if fields := componentFields(updated.Name, updated.Description); len(fields) > 0 {
		return Component{}, validationFailed(fields)
	}
	if err := s.requireExistingUser(ctx, "owner_user_id", updated.OwnerUserID); err != nil {
		return Component{}, err
	}
	if err := s.requireExistingUser(ctx, "default_assignee_id", updated.DefaultAssigneeID); err != nil {
		return Component{}, err
	}
	updated.UpdatedAt = s.now().UTC()
	if _, err := s.db.ExecContext(ctx, `
		UPDATE project_components
		SET name = ?, description = ?, owner_user_id = ?, default_assignee_id = ?, updated_at = ?
		WHERE id = ?
	`, updated.Name, nullableString(updated.Description), nullableString(updated.OwnerUserID), nullableString(updated.DefaultAssigneeID), formatTime(updated.UpdatedAt), updated.ID); err != nil {
		if isUniqueConstraint(err) {
			return Component{}, conflict("component", "name", updated.Name)
		}
		return Component{}, fmt.Errorf("update component: %w", err)
	}
	return updated, nil
}

func (s *Service) DeleteComponent(ctx context.Context, principal authz.Principal, componentID string) error {
	component, err := s.getComponent(ctx, componentID)
	if err != nil {
		return err
	}
	if err := s.require(principal, authz.PermissionProjectsWrite, authz.ProjectScope(component.ProjectID)); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM project_components WHERE id = ?", component.ID)
	if err != nil {
		return fmt.Errorf("delete component: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check component delete: %w", err)
	}
	if affected == 0 {
		return notFound("component", componentID)
	}
	return nil
}

func (s *Service) ListVersions(ctx context.Context, principal authz.Principal, projectID string, status string) ([]Version, error) {
	projectID = strings.TrimSpace(projectID)
	status = normalizeSlug(status)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if status != "" && !validVersionStatus(status) {
		return nil, validationFailed(map[string]string{"status": "Invalid version status"})
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	query := `
		SELECT id, project_id, name, description, status, target_date, release_date, created_at, updated_at
		FROM project_versions
		WHERE project_id = ?`
	args := []any{projectID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY name ASC, id ASC"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	defer rows.Close()
	var versions []Version
	for rows.Next() {
		version, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate versions: %w", err)
	}
	return versions, nil
}

func (s *Service) CreateVersion(ctx context.Context, principal authz.Principal, input CreateVersionInput) (Version, error) {
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	if input.ProjectID == "" {
		return Version{}, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionProjectsWrite, authz.ProjectScope(input.ProjectID)); err != nil {
		return Version{}, err
	}
	if _, err := s.repo.GetProject(ctx, input.ProjectID); err != nil {
		return Version{}, err
	}
	version, err := s.buildVersion(input)
	if err != nil {
		return Version{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO project_versions (
			id, project_id, name, description, status, target_date, release_date, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, version.ID, version.ProjectID, version.Name, nullableString(version.Description), version.Status, nullableString(version.TargetDate), nullableString(version.ReleaseDate), formatTime(version.CreatedAt), formatTime(version.UpdatedAt)); err != nil {
		if isUniqueConstraint(err) {
			return Version{}, conflict("version", "name", version.Name)
		}
		return Version{}, fmt.Errorf("insert version: %w", err)
	}
	return version, nil
}

func (s *Service) GetVersion(ctx context.Context, principal authz.Principal, versionID string) (Version, error) {
	version, err := s.getVersion(ctx, versionID)
	if err != nil {
		return Version{}, err
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(version.ProjectID)); err != nil {
		return Version{}, err
	}
	return version, nil
}

func (s *Service) GetVersionReport(ctx context.Context, principal authz.Principal, versionID string) (VersionReport, error) {
	version, err := s.getVersion(ctx, versionID)
	if err != nil {
		return VersionReport{}, err
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(version.ProjectID)); err != nil {
		return VersionReport{}, err
	}
	tickets, err := s.listVersionReportTickets(ctx, version.ID)
	if err != nil {
		return VersionReport{}, err
	}
	tickets, err = s.attachTicketDetailsAndWatcherStatus(ctx, principal, tickets)
	if err != nil {
		return VersionReport{}, err
	}
	return VersionReport{
		Version:  version,
		Progress: versionReportProgress(tickets),
		Tickets:  tickets,
	}, nil
}

func (s *Service) UpdateVersion(ctx context.Context, principal authz.Principal, versionID string, input UpdateVersionInput) (Version, error) {
	current, err := s.getVersion(ctx, versionID)
	if err != nil {
		return Version{}, err
	}
	if err := s.require(principal, authz.PermissionProjectsWrite, authz.ProjectScope(current.ProjectID)); err != nil {
		return Version{}, err
	}
	updated := current
	if input.Name != nil {
		updated.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		updated.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		updated.Status = normalizeSlug(*input.Status)
	}
	if input.TargetDate != nil {
		updated.TargetDate = strings.TrimSpace(*input.TargetDate)
	}
	if input.ReleaseDate != nil {
		updated.ReleaseDate = strings.TrimSpace(*input.ReleaseDate)
	}
	if fields := versionFields(updated.Name, updated.Description, updated.Status, updated.TargetDate, updated.ReleaseDate); len(fields) > 0 {
		return Version{}, validationFailed(fields)
	}
	updated.UpdatedAt = s.now().UTC()
	if _, err := s.db.ExecContext(ctx, `
		UPDATE project_versions
		SET name = ?, description = ?, status = ?, target_date = ?, release_date = ?, updated_at = ?
		WHERE id = ?
	`, updated.Name, nullableString(updated.Description), updated.Status, nullableString(updated.TargetDate), nullableString(updated.ReleaseDate), formatTime(updated.UpdatedAt), updated.ID); err != nil {
		if isUniqueConstraint(err) {
			return Version{}, conflict("version", "name", updated.Name)
		}
		return Version{}, fmt.Errorf("update version: %w", err)
	}
	return updated, nil
}

func (s *Service) DeleteVersion(ctx context.Context, principal authz.Principal, versionID string) error {
	version, err := s.getVersion(ctx, versionID)
	if err != nil {
		return err
	}
	if err := s.require(principal, authz.PermissionProjectsWrite, authz.ProjectScope(version.ProjectID)); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM project_versions WHERE id = ?", version.ID)
	if err != nil {
		return fmt.Errorf("delete version: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check version delete: %w", err)
	}
	if affected == 0 {
		return notFound("version", versionID)
	}
	return nil
}

func (s *Service) buildComponent(ctx context.Context, input CreateComponentInput) (Component, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.OwnerUserID = strings.TrimSpace(input.OwnerUserID)
	input.DefaultAssigneeID = strings.TrimSpace(input.DefaultAssigneeID)
	if fields := componentFields(input.Name, input.Description); len(fields) > 0 {
		return Component{}, validationFailed(fields)
	}
	if err := s.requireExistingUser(ctx, "owner_user_id", input.OwnerUserID); err != nil {
		return Component{}, err
	}
	if err := s.requireExistingUser(ctx, "default_assignee_id", input.DefaultAssigneeID); err != nil {
		return Component{}, err
	}
	id, err := newID("component")
	if err != nil {
		return Component{}, err
	}
	now := s.now().UTC()
	return Component{
		ID:                id,
		ProjectID:         input.ProjectID,
		Name:              input.Name,
		Description:       input.Description,
		OwnerUserID:       input.OwnerUserID,
		DefaultAssigneeID: input.DefaultAssigneeID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (s *Service) buildVersion(input CreateVersionInput) (Version, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = normalizeSlug(input.Status)
	if input.Status == "" {
		input.Status = VersionStatusPlanned
	}
	input.TargetDate = strings.TrimSpace(input.TargetDate)
	input.ReleaseDate = strings.TrimSpace(input.ReleaseDate)
	if fields := versionFields(input.Name, input.Description, input.Status, input.TargetDate, input.ReleaseDate); len(fields) > 0 {
		return Version{}, validationFailed(fields)
	}
	id, err := newID("version")
	if err != nil {
		return Version{}, err
	}
	now := s.now().UTC()
	return Version{
		ID:          id,
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: input.Description,
		Status:      input.Status,
		TargetDate:  input.TargetDate,
		ReleaseDate: input.ReleaseDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func componentFields(name string, description string) map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(name) == "" {
		fields["name"] = "Required"
	}
	if len(strings.TrimSpace(name)) > 200 {
		fields["name"] = "Must be 200 characters or fewer"
	}
	if len(strings.TrimSpace(description)) > 2000 {
		fields["description"] = "Must be 2000 characters or fewer"
	}
	return fields
}

func versionFields(name string, description string, status string, targetDate string, releaseDate string) map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(name) == "" {
		fields["name"] = "Required"
	}
	if len(strings.TrimSpace(name)) > 200 {
		fields["name"] = "Must be 200 characters or fewer"
	}
	if len(strings.TrimSpace(description)) > 2000 {
		fields["description"] = "Must be 2000 characters or fewer"
	}
	if !validVersionStatus(status) {
		fields["status"] = "Invalid version status"
	}
	validateSprintDate(fields, "target_date", targetDate)
	validateSprintDate(fields, "release_date", releaseDate)
	return fields
}

func validVersionStatus(status string) bool {
	switch status {
	case VersionStatusPlanned, VersionStatusReleased, VersionStatusArchived:
		return true
	default:
		return false
	}
}

func (s *Service) requireExistingUser(ctx context.Context, field string, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	var exists int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`, userID).Scan(&exists); err != nil {
		return fmt.Errorf("check user: %w", err)
	}
	if exists != 1 {
		return validationFailed(map[string]string{field: "User not found"})
	}
	return nil
}

func (s *Service) getComponent(ctx context.Context, componentID string) (Component, error) {
	componentID = strings.TrimSpace(componentID)
	if componentID == "" {
		return Component{}, validationFailed(map[string]string{"component_id": "Required"})
	}
	component, err := scanComponent(s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, description, owner_user_id, default_assignee_id, created_at, updated_at
		FROM project_components
		WHERE id = ?
	`, componentID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Component{}, notFound("component", componentID)
		}
		return Component{}, fmt.Errorf("get component: %w", err)
	}
	return component, nil
}

func (s *Service) getVersion(ctx context.Context, versionID string) (Version, error) {
	versionID = strings.TrimSpace(versionID)
	if versionID == "" {
		return Version{}, validationFailed(map[string]string{"version_id": "Required"})
	}
	version, err := scanVersion(s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, description, status, target_date, release_date, created_at, updated_at
		FROM project_versions
		WHERE id = ?
	`, versionID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Version{}, notFound("version", versionID)
		}
		return Version{}, fmt.Errorf("get version: %w", err)
	}
	return version, nil
}

func (s *Service) listVersionReportTickets(ctx context.Context, versionID string) ([]Ticket, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, key, title, description, status, priority, type,
			reporter_id, assignee_id, parent_ticket_id, sprint_id, component_id, version_id, rank, start_date, due_date, created_at, updated_at, deleted_at
		FROM tickets
		WHERE version_id = ? AND deleted_at IS NULL
		ORDER BY status ASC, created_at DESC, key DESC
	`, versionID)
	if err != nil {
		return nil, fmt.Errorf("list version report tickets: %w", err)
	}
	defer rows.Close()

	tickets := []Ticket{}
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, fmt.Errorf("scan version report ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate version report tickets: %w", err)
	}
	return tickets, nil
}

func versionReportProgress(tickets []Ticket) VersionReportProgress {
	progress := VersionReportProgress{ByStatus: map[string]int{}}
	for _, ticket := range tickets {
		progress.Total++
		progress.ByStatus[ticket.Status]++
		if ticket.Status == "done" {
			progress.Done++
		}
		if ticket.ComponentID == "" {
			progress.UnassignedComponent++
		}
	}
	progress.Open = progress.Total - progress.Done
	return progress
}

func scanComponent(scanner rowScanner) (Component, error) {
	var component Component
	var description sql.NullString
	var ownerUserID sql.NullString
	var defaultAssigneeID sql.NullString
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(&component.ID, &component.ProjectID, &component.Name, &description, &ownerUserID, &defaultAssigneeID, &createdAt, &updatedAt); err != nil {
		return Component{}, err
	}
	component.Description = nullString(description)
	component.OwnerUserID = nullString(ownerUserID)
	component.DefaultAssigneeID = nullString(defaultAssigneeID)
	var err error
	component.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Component{}, fmt.Errorf("parse component created_at: %w", err)
	}
	component.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Component{}, fmt.Errorf("parse component updated_at: %w", err)
	}
	return component, nil
}

func scanVersion(scanner rowScanner) (Version, error) {
	var version Version
	var description sql.NullString
	var targetDate sql.NullString
	var releaseDate sql.NullString
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(&version.ID, &version.ProjectID, &version.Name, &description, &version.Status, &targetDate, &releaseDate, &createdAt, &updatedAt); err != nil {
		return Version{}, err
	}
	version.Description = nullString(description)
	version.TargetDate = nullString(targetDate)
	version.ReleaseDate = nullString(releaseDate)
	var err error
	version.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Version{}, fmt.Errorf("parse version created_at: %w", err)
	}
	version.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Version{}, fmt.Errorf("parse version updated_at: %w", err)
	}
	return version, nil
}

func isUniqueConstraint(err error) bool {
	return isSQLiteCode(err, sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY) || isSQLiteCode(err, sqlite3.SQLITE_CONSTRAINT_UNIQUE)
}
