package notifications

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestUserPreferencesDefaultsAndUpdate(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	seedNotificationUser(t, ctx, db, "user-1")

	now := time.Date(2026, 6, 17, 9, 0, 0, 0, time.UTC)
	service := NewService(db.SQL, WithNow(func() time.Time { return now }))

	defaults, err := service.GetUserPreferences(ctx, "user-1")
	if err != nil {
		t.Fatalf("get default preferences: %v", err)
	}
	if defaults.Customized || !defaults.InAppEnabled || !defaults.ExternalEnabled || !defaults.AssignmentEnabled {
		t.Fatalf("unexpected default preferences: %#v", defaults)
	}

	disabled := false
	updated, err := service.UpdateUserPreferences(ctx, "user-1", UpdatePreferencesInput{
		ExternalEnabled:     &disabled,
		StatusChangeEnabled: &disabled,
	})
	if err != nil {
		t.Fatalf("update preferences: %v", err)
	}
	if updated.ID == "" || !updated.Customized || updated.ExternalEnabled || updated.StatusChangeEnabled || !updated.InAppEnabled {
		t.Fatalf("unexpected updated preferences: %#v", updated)
	}
	if !updated.CreatedAt.Equal(now) || !updated.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected preference timestamps: %#v", updated)
	}

	got, err := service.GetUserPreferences(ctx, "user-1")
	if err != nil {
		t.Fatalf("get stored preferences: %v", err)
	}
	if got.ID != updated.ID || got.ExternalEnabled || got.StatusChangeEnabled {
		t.Fatalf("unexpected stored preferences: %#v", got)
	}
}

func TestProjectPreferencesDefaultsAndUpdate(t *testing.T) {
	ctx := context.Background()
	db := openNotificationTestDB(t, ctx)
	seedNotificationProject(t, ctx, db, "project-1", "CORE")

	service := NewService(db.SQL)
	defaults, err := service.GetProjectPreferences(ctx, "project-1")
	if err != nil {
		t.Fatalf("get project default preferences: %v", err)
	}
	if defaults.ScopeType != PreferenceScopeProject || defaults.ProjectID != "project-1" || defaults.Customized || !defaults.CommentEnabled {
		t.Fatalf("unexpected project defaults: %#v", defaults)
	}

	disabled := false
	updated, err := service.UpdateProjectPreferences(ctx, "project-1", UpdatePreferencesInput{
		CommentEnabled: &disabled,
	})
	if err != nil {
		t.Fatalf("update project preferences: %v", err)
	}
	if updated.CommentEnabled || !updated.Customized {
		t.Fatalf("unexpected project preferences: %#v", updated)
	}

	if _, err := service.GetProjectPreferences(ctx, "missing-project"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing project not found, got %v", err)
	}
}
