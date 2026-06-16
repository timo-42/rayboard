package notifications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	PreferenceScopeUser    = "user"
	PreferenceScopeProject = "project"
)

type Preferences struct {
	ID                       string
	ScopeType                string
	UserID                   string
	ProjectID                string
	InAppEnabled             bool
	ExternalEnabled          bool
	AssignmentEnabled        bool
	CommentEnabled           bool
	StatusChangeEnabled      bool
	SprintChangeEnabled      bool
	ReleaseChangeEnabled     bool
	AutomationFailureEnabled bool
	CreatedAt                time.Time
	UpdatedAt                time.Time
	Customized               bool
}

type UpdatePreferencesInput struct {
	InAppEnabled             *bool
	ExternalEnabled          *bool
	AssignmentEnabled        *bool
	CommentEnabled           *bool
	StatusChangeEnabled      *bool
	SprintChangeEnabled      *bool
	ReleaseChangeEnabled     *bool
	AutomationFailureEnabled *bool
}

func (s *Service) GetUserPreferences(ctx context.Context, userID string) (Preferences, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return Preferences{}, fmt.Errorf("%w: user_id is required", ErrValidation)
	}
	if err := s.requireUser(ctx, userID); err != nil {
		return Preferences{}, err
	}
	preferences, err := s.getPreferences(ctx, PreferenceScopeUser, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return defaultPreferences(PreferenceScopeUser, userID, ""), nil
		}
		return Preferences{}, err
	}
	return preferences, nil
}

func (s *Service) UpdateUserPreferences(ctx context.Context, userID string, input UpdatePreferencesInput) (Preferences, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return Preferences{}, fmt.Errorf("%w: user_id is required", ErrValidation)
	}
	if err := s.requireUser(ctx, userID); err != nil {
		return Preferences{}, err
	}
	return s.upsertPreferences(ctx, PreferenceScopeUser, userID, "", input)
}

func (s *Service) GetProjectPreferences(ctx context.Context, projectID string) (Preferences, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return Preferences{}, fmt.Errorf("%w: project_id is required", ErrValidation)
	}
	if err := s.requireProject(ctx, projectID); err != nil {
		return Preferences{}, err
	}
	preferences, err := s.getPreferences(ctx, PreferenceScopeProject, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return defaultPreferences(PreferenceScopeProject, "", projectID), nil
		}
		return Preferences{}, err
	}
	return preferences, nil
}

func (s *Service) UpdateProjectPreferences(ctx context.Context, projectID string, input UpdatePreferencesInput) (Preferences, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return Preferences{}, fmt.Errorf("%w: project_id is required", ErrValidation)
	}
	if err := s.requireProject(ctx, projectID); err != nil {
		return Preferences{}, err
	}
	return s.upsertPreferences(ctx, PreferenceScopeProject, "", projectID, input)
}

func (s *Service) upsertPreferences(ctx context.Context, scopeType string, userID string, projectID string, input UpdatePreferencesInput) (Preferences, error) {
	scopeKey := preferencesScopeKey(scopeType, userID, projectID)
	current, err := s.getPreferences(ctx, scopeType, scopeKey)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return Preferences{}, err
		}
		current = defaultPreferences(scopeType, userID, projectID)
	}
	applyPreferenceUpdate(&current, input)
	current.Customized = true
	now := s.now().UTC()
	if current.ID == "" {
		id, err := newID("pref")
		if err != nil {
			return Preferences{}, err
		}
		current.ID = id
		current.CreatedAt = now
	}
	current.UpdatedAt = now

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO notification_preferences (
			id, scope_type, scope_key, user_id, project_id,
			in_app_enabled, external_enabled, assignment_enabled, comment_enabled,
			status_change_enabled, sprint_change_enabled, release_change_enabled,
			automation_failure_enabled, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(scope_type, scope_key) DO UPDATE SET
			in_app_enabled = excluded.in_app_enabled,
			external_enabled = excluded.external_enabled,
			assignment_enabled = excluded.assignment_enabled,
			comment_enabled = excluded.comment_enabled,
			status_change_enabled = excluded.status_change_enabled,
			sprint_change_enabled = excluded.sprint_change_enabled,
			release_change_enabled = excluded.release_change_enabled,
			automation_failure_enabled = excluded.automation_failure_enabled,
			updated_at = excluded.updated_at
	`, current.ID, current.ScopeType, scopeKey, nullableString(current.UserID), nullableString(current.ProjectID),
		boolInt(current.InAppEnabled), boolInt(current.ExternalEnabled), boolInt(current.AssignmentEnabled), boolInt(current.CommentEnabled),
		boolInt(current.StatusChangeEnabled), boolInt(current.SprintChangeEnabled), boolInt(current.ReleaseChangeEnabled),
		boolInt(current.AutomationFailureEnabled), formatTime(current.CreatedAt), formatTime(current.UpdatedAt)); err != nil {
		return Preferences{}, fmt.Errorf("upsert notification preferences: %w", err)
	}
	return s.getPreferences(ctx, scopeType, scopeKey)
}

func (s *Service) getPreferences(ctx context.Context, scopeType string, scopeKey string) (Preferences, error) {
	preferences, err := scanPreferences(s.db.QueryRowContext(ctx, `
		SELECT id, scope_type, user_id, project_id,
			in_app_enabled, external_enabled, assignment_enabled, comment_enabled,
			status_change_enabled, sprint_change_enabled, release_change_enabled,
			automation_failure_enabled, created_at, updated_at
		FROM notification_preferences
		WHERE scope_type = ? AND scope_key = ?
	`, scopeType, scopeKey))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Preferences{}, ErrNotFound
		}
		return Preferences{}, fmt.Errorf("get notification preferences: %w", err)
	}
	preferences.Customized = true
	return preferences, nil
}

func defaultPreferences(scopeType string, userID string, projectID string) Preferences {
	return Preferences{
		ScopeType:                scopeType,
		UserID:                   userID,
		ProjectID:                projectID,
		InAppEnabled:             true,
		ExternalEnabled:          true,
		AssignmentEnabled:        true,
		CommentEnabled:           true,
		StatusChangeEnabled:      true,
		SprintChangeEnabled:      true,
		ReleaseChangeEnabled:     true,
		AutomationFailureEnabled: true,
	}
}

func applyPreferenceUpdate(preferences *Preferences, input UpdatePreferencesInput) {
	if input.InAppEnabled != nil {
		preferences.InAppEnabled = *input.InAppEnabled
	}
	if input.ExternalEnabled != nil {
		preferences.ExternalEnabled = *input.ExternalEnabled
	}
	if input.AssignmentEnabled != nil {
		preferences.AssignmentEnabled = *input.AssignmentEnabled
	}
	if input.CommentEnabled != nil {
		preferences.CommentEnabled = *input.CommentEnabled
	}
	if input.StatusChangeEnabled != nil {
		preferences.StatusChangeEnabled = *input.StatusChangeEnabled
	}
	if input.SprintChangeEnabled != nil {
		preferences.SprintChangeEnabled = *input.SprintChangeEnabled
	}
	if input.ReleaseChangeEnabled != nil {
		preferences.ReleaseChangeEnabled = *input.ReleaseChangeEnabled
	}
	if input.AutomationFailureEnabled != nil {
		preferences.AutomationFailureEnabled = *input.AutomationFailureEnabled
	}
}

func scanPreferences(scanner interface{ Scan(...any) error }) (Preferences, error) {
	var preferences Preferences
	var userID sql.NullString
	var projectID sql.NullString
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&preferences.ID,
		&preferences.ScopeType,
		&userID,
		&projectID,
		&preferences.InAppEnabled,
		&preferences.ExternalEnabled,
		&preferences.AssignmentEnabled,
		&preferences.CommentEnabled,
		&preferences.StatusChangeEnabled,
		&preferences.SprintChangeEnabled,
		&preferences.ReleaseChangeEnabled,
		&preferences.AutomationFailureEnabled,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Preferences{}, err
	}
	preferences.UserID = nullString(userID)
	preferences.ProjectID = nullString(projectID)
	created, err := parseTime(createdAt)
	if err != nil {
		return Preferences{}, err
	}
	updated, err := parseTime(updatedAt)
	if err != nil {
		return Preferences{}, err
	}
	preferences.CreatedAt = created
	preferences.UpdatedAt = updated
	return preferences, nil
}

func preferencesScopeKey(scopeType string, userID string, projectID string) string {
	if scopeType == PreferenceScopeProject {
		return projectID
	}
	return userID
}

func (s *Service) requireUser(ctx context.Context, userID string) error {
	var exists int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`, userID).Scan(&exists); err != nil {
		return fmt.Errorf("check notification preference user: %w", err)
	}
	if exists == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) requireProject(ctx context.Context, projectID string) error {
	var exists int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM projects
		WHERE id = ? AND deleted_at IS NULL
	`, projectID).Scan(&exists); err != nil {
		return fmt.Errorf("check notification preference project: %w", err)
	}
	if exists == 0 {
		return ErrNotFound
	}
	return nil
}
