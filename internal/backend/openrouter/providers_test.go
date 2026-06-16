package openrouter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/openrouter"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestProviderLifecycleRedactsAPIKey(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/rayboard.sqlite")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	service := openrouter.NewService(db.SQL)

	created, err := service.CreateProvider(ctx, openrouter.CreateProviderInput{
		Name:                  "Default",
		DefaultModel:          "openai/gpt-4.1-mini",
		APIKey:                "sk-or-secret",
		AllowedModels:         []string{"openai/gpt-4.1-mini", "openai/gpt-4.1-mini", "anthropic/claude-sonnet-4"},
		DefaultTimeoutSeconds: 45,
		MaxOutputTokens:       4096,
		Enabled:               true,
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if created.ID == "" || created.Name != "default" || !created.APIKeySet {
		t.Fatalf("unexpected provider: %#v", created)
	}
	if len(created.AllowedModels) != 2 {
		t.Fatalf("expected allowed models to be normalized, got %#v", created.AllowedModels)
	}

	fetched, err := service.GetProvider(ctx, created.ID)
	if err != nil {
		t.Fatalf("get provider: %v", err)
	}
	if fetched.APIKeySet != true || fetched.DefaultModel != created.DefaultModel {
		t.Fatalf("unexpected fetched provider: %#v", fetched)
	}

	updatedModel := "openai/gpt-4.1"
	newKey := "sk-or-rotated"
	updated, err := service.UpdateProvider(ctx, created.ID, openrouter.UpdateProviderInput{
		DefaultModel: &updatedModel,
		APIKey:       &newKey,
	})
	if err != nil {
		t.Fatalf("update provider: %v", err)
	}
	if updated.DefaultModel != updatedModel || !updated.APIKeySet {
		t.Fatalf("unexpected updated provider: %#v", updated)
	}

	if err := service.DeleteProvider(ctx, created.ID); err != nil {
		t.Fatalf("delete provider: %v", err)
	}
	if _, err := service.GetProvider(ctx, created.ID); !errors.Is(err, openrouter.ErrNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestProviderValidation(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/rayboard.sqlite")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	service := openrouter.NewService(db.SQL)

	if _, err := service.CreateProvider(ctx, openrouter.CreateProviderInput{}); !errors.Is(err, openrouter.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if _, err := service.CreateProvider(ctx, openrouter.CreateProviderInput{
		Name:         "default",
		DefaultModel: "openai/gpt-4.1-mini",
		APIKey:       "sk-or-secret",
	}); err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if _, err := service.CreateProvider(ctx, openrouter.CreateProviderInput{
		Name:         "default",
		DefaultModel: "openai/gpt-4.1-mini",
		APIKey:       "sk-or-secret-2",
	}); !errors.Is(err, openrouter.ErrConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
}
