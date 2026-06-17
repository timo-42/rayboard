package openrouter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestCompleteJSONUsesProviderSecretAndParsesObject(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/rayboard.sqlite")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	var receivedAuth string
	var receivedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		receivedAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "gen_123",
			"choices": [{"message": {"role": "assistant", "content": "{\"accepted\":true,\"count\":2}"}}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 5}
		}`))
	}))
	defer server.Close()

	service := openrouter.NewService(db.SQL, openrouter.WithBaseURL(server.URL))
	provider, err := service.CreateProvider(ctx, openrouter.CreateProviderInput{
		Name:                  "Default",
		DefaultModel:          "openai/gpt-4.1-mini",
		APIKey:                "sk-or-secret",
		DefaultTimeoutSeconds: 12,
		MaxOutputTokens:       123,
		Enabled:               true,
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	result, err := service.CompleteJSON(ctx, openrouter.CompletionInput{
		ProviderID: provider.ID,
		Prompt:     "Return a test result as JSON.",
	})
	if err != nil {
		t.Fatalf("complete JSON: %v", err)
	}
	if receivedAuth != "Bearer sk-or-secret" {
		t.Fatalf("unexpected authorization header: %q", receivedAuth)
	}
	if receivedBody["model"] != "openai/gpt-4.1-mini" || receivedBody["max_tokens"] != float64(123) {
		t.Fatalf("unexpected OpenRouter request body: %#v", receivedBody)
	}
	responseFormat, ok := receivedBody["response_format"].(map[string]any)
	if !ok || responseFormat["type"] != "json_object" {
		t.Fatalf("expected json_object response format, got %#v", receivedBody["response_format"])
	}
	if result.ResponseID != "gen_123" || result.Model != "openai/gpt-4.1-mini" || result.Output["accepted"] != true || result.Output["count"] != float64(2) {
		t.Fatalf("unexpected completion result: %#v", result)
	}
	fetched, err := service.GetProvider(ctx, provider.ID)
	if err != nil {
		t.Fatalf("get provider: %v", err)
	}
	if fetched.APIKeySet != true {
		t.Fatalf("expected public provider to report key set")
	}
}

func TestCompleteJSONRejectsInvalidOutput(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/rayboard.sqlite")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"not-json"}}]}`))
	}))
	defer server.Close()

	service := openrouter.NewService(db.SQL, openrouter.WithBaseURL(server.URL))
	provider, err := service.CreateProvider(ctx, openrouter.CreateProviderInput{
		Name:         "Default",
		DefaultModel: "openai/gpt-4.1-mini",
		APIKey:       "sk-or-secret",
		Enabled:      true,
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if _, err := service.CompleteJSON(ctx, openrouter.CompletionInput{ProviderID: provider.ID, Prompt: "Return JSON"}); !errors.Is(err, openrouter.ErrValidation) {
		t.Fatalf("expected validation error for invalid output, got %v", err)
	}
}
