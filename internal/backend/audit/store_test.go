package audit_test

import (
	"context"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestStoreRecordAndList(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/rayboard.sqlite")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	auditStore := audit.NewStore(db.SQL, audit.WithNow(func() time.Time { return now }))
	entry, err := auditStore.Record(ctx, audit.RecordInput{
		EventType:   "auth.api_token_created",
		ActorID:     "user_admin",
		AuthKind:    authz.AuthKindSession,
		SubjectType: "api_token",
		SubjectID:   "token_1",
		Payload: map[string]any{
			"token_name": "demo",
		},
	})
	if err != nil {
		t.Fatalf("record audit entry: %v", err)
	}
	if entry.ID == "" || entry.Outcome != audit.OutcomeSuccess || !entry.OccurredAt.Equal(now) {
		t.Fatalf("unexpected recorded entry: %#v", entry)
	}

	entries, err := auditStore.List(ctx, 10)
	if err != nil {
		t.Fatalf("list audit entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %#v", entries)
	}
	got := entries[0]
	if got.EventType != "auth.api_token_created" || got.ActorID != "user_admin" || got.AuthKind != authz.AuthKindSession || got.SubjectID != "token_1" {
		t.Fatalf("unexpected listed entry: %#v", got)
	}
	if got.Payload["token_name"] != "demo" {
		t.Fatalf("unexpected payload: %#v", got.Payload)
	}
}

func TestStoreValidation(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, t.TempDir()+"/rayboard.sqlite")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	auditStore := audit.NewStore(db.SQL)
	if _, err := auditStore.Record(ctx, audit.RecordInput{SubjectType: "user"}); err == nil {
		t.Fatal("expected missing event type error")
	}
	if _, err := auditStore.Record(ctx, audit.RecordInput{EventType: "user.deleted"}); err == nil {
		t.Fatal("expected missing subject type error")
	}
	if _, err := auditStore.Record(ctx, audit.RecordInput{EventType: "user.deleted", SubjectType: "user", Outcome: "maybe"}); err == nil {
		t.Fatal("expected invalid outcome error")
	}
}
