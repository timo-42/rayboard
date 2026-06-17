package settings_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/settings"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestGlobalSettingsDefaultsUpdateAndAttachmentPolicy(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	service := settings.NewService(db.SQL, settings.WithNow(fixedNow))

	global, err := service.GetGlobal(ctx)
	if err != nil {
		t.Fatalf("get default settings: %v", err)
	}
	if global.AttachmentMaxSizeBytes != 10<<20 || !global.DemoWarningEnabled || global.BackupEnabled {
		t.Fatalf("unexpected defaults: %#v", global)
	}

	maxSize := int64(1024)
	contentTypes := []string{"TEXT/PLAIN", "application/json", "text/plain"}
	webhooks := []string{"https://example.com/hooks/", "https://example.com/hooks"}
	note := "healthy"
	backup := true
	updated, err := service.UpdateGlobal(ctx, settings.UpdateGlobalInput{
		AttachmentMaxSizeBytes:        &maxSize,
		AttachmentAllowedContentTypes: &contentTypes,
		WebhookAllowedBaseURLs:        &webhooks,
		BackupEnabled:                 &backup,
		SystemHealthNote:              &note,
		UpdatedBy:                     "user-admin",
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if updated.UpdatedBy != "user-admin" || updated.SystemHealthNote != note {
		t.Fatalf("unexpected update metadata: %#v", updated)
	}
	if !slices.Equal(updated.AttachmentAllowedContentTypes, []string{"application/json", "text/plain"}) {
		t.Fatalf("expected normalized content types, got %#v", updated.AttachmentAllowedContentTypes)
	}
	if !slices.Equal(updated.WebhookAllowedBaseURLs, []string{"https://example.com/hooks"}) {
		t.Fatalf("expected normalized webhook URLs, got %#v", updated.WebhookAllowedBaseURLs)
	}

	policy, err := service.AttachmentPolicy(ctx)
	if err != nil {
		t.Fatalf("get attachment policy: %v", err)
	}
	if policy.MaxSizeBytes != maxSize || !slices.Equal(policy.AllowedContentTypes, []string{"application/json", "text/plain"}) {
		t.Fatalf("unexpected attachment policy: %#v", policy)
	}
	baseURLs, err := service.OutgoingWebhookBaseURLs(ctx)
	if err != nil {
		t.Fatalf("get outgoing webhook base URLs: %v", err)
	}
	if !slices.Equal(baseURLs, []string{"https://example.com/hooks"}) {
		t.Fatalf("unexpected outgoing webhook base URLs: %#v", baseURLs)
	}
	baseURLs[0] = "https://mutated.example.com"
	again, err := service.OutgoingWebhookBaseURLs(ctx)
	if err != nil {
		t.Fatalf("get outgoing webhook base URLs again: %v", err)
	}
	if !slices.Equal(again, []string{"https://example.com/hooks"}) {
		t.Fatalf("outgoing webhook base URLs should be cloned, got %#v", again)
	}
}

func TestGlobalSettingsValidation(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	service := settings.NewService(db.SQL)

	tooLarge := int64(101 << 20)
	if _, err := service.UpdateGlobal(ctx, settings.UpdateGlobalInput{AttachmentMaxSizeBytes: &tooLarge}); !errors.Is(err, settings.ErrValidation) {
		t.Fatalf("expected max size validation, got %v", err)
	}
	badTypes := []string{"not-a-type"}
	if _, err := service.UpdateGlobal(ctx, settings.UpdateGlobalInput{AttachmentAllowedContentTypes: &badTypes}); !errors.Is(err, settings.ErrValidation) {
		t.Fatalf("expected content type validation, got %v", err)
	}
	for _, badURLs := range [][]string{
		{"ftp://example.com"},
		{"https://user@example.com/hooks"},
		{"https://example.com/hooks?token=secret"},
		{"https://example.com/hooks#fragment"},
	} {
		if _, err := service.UpdateGlobal(ctx, settings.UpdateGlobalInput{WebhookAllowedBaseURLs: &badURLs}); !errors.Is(err, settings.ErrValidation) {
			t.Fatalf("expected URL validation for %#v, got %v", badURLs, err)
		}
	}
}

func openTestDB(t *testing.T, ctx context.Context) *store.DB {
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

func fixedNow() time.Time {
	return time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
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
