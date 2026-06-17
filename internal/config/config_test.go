package config

import "testing"

func TestFromEnv(t *testing.T) {
	t.Setenv("RAYBOARD_FRONTEND_ADDR", "127.0.0.1:9000")
	t.Setenv("RAYBOARD_BACKEND_ADDR", "127.0.0.1:9001")
	t.Setenv("RAYBOARD_BACKEND_URL", "http://127.0.0.1:9001")
	t.Setenv("RAYBOARD_DB", "test.sqlite")
	t.Setenv("RAYBOARD_OUTGOING_WEBHOOK_BASE_URL", "https://hooks.example.test")

	cfg := FromEnv()

	if cfg.FrontendAddr != "127.0.0.1:9000" {
		t.Fatalf("unexpected frontend addr: %s", cfg.FrontendAddr)
	}
	if cfg.BackendAddr != "127.0.0.1:9001" {
		t.Fatalf("unexpected backend addr: %s", cfg.BackendAddr)
	}
	if cfg.BackendURL != "http://127.0.0.1:9001" {
		t.Fatalf("unexpected backend URL: %s", cfg.BackendURL)
	}
	if cfg.DBPath != "test.sqlite" {
		t.Fatalf("unexpected DB path: %s", cfg.DBPath)
	}
	if cfg.OutgoingWebhookBaseURL != "https://hooks.example.test" {
		t.Fatalf("unexpected outgoing webhook base URL: %s", cfg.OutgoingWebhookBaseURL)
	}
}
