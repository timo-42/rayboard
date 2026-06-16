package app

import (
	"bytes"
	"context"
	"testing"
)

func TestMainUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Main(context.Background(), []string{"missing"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("expected usage output on stderr")
	}
}

func TestDemoSeedRequiresFreshReset(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Main(context.Background(), []string{
		"demo", "seed",
		"--backend-url", "http://127.0.0.1:8081",
		"--admin-user", "admin",
		"--admin-password", "secret",
	}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
}
