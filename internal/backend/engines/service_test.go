package engines_test

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/engines"
	"github.com/timo-42/rayboard/internal/backend/store"
)

func TestLuaEngineTestReturnsOutputAndLogs(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))

	run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
		Surface: "ticket_hook_before",
		Context: map[string]any{"ticket_id": "ticket-1", "dry_run": false},
		Input:   map[string]any{"title": "Example"},
		Engine: engines.EngineSpec{
			Type: "lua",
			Script: `
rayboard.log("checking " .. input.title)
return { ok = true, title = input.title, surface = context.surface, ticket_id = context.ticket_id, dry_run = context.dry_run }
`,
		},
	})
	if err != nil {
		t.Fatalf("test lua engine: %v", err)
	}
	if run.Status != automation.StatusSucceeded {
		t.Fatalf("expected succeeded run, got %#v", run)
	}
	output, _ := run.Output["output"].(map[string]any)
	if output["ok"] != true || output["title"] != "Example" || output["surface"] != "ticket_hook_before" || output["ticket_id"] != "ticket-1" || output["dry_run"] != true {
		t.Fatalf("unexpected output: %#v", run.Output)
	}
	logs, _ := run.Output["logs"].([]any)
	if !slices.Equal(anyStrings(logs), []string{"checking Example"}) {
		t.Fatalf("unexpected logs: %#v", run.Output)
	}
	if encoded := run.Input; encoded["engine"] != "lua" {
		t.Fatalf("expected run input to store only engine type, got %#v", encoded)
	}
	inputEnvelope, _ := run.Input["input"].(map[string]any)
	contextEnvelope, _ := inputEnvelope["context"].(map[string]any)
	if inputEnvelope["dry_run"] != true || contextEnvelope["ticket_id"] != "ticket-1" || contextEnvelope["dry_run"] != true {
		t.Fatalf("expected normalized dry-run context in run input, got %#v", run.Input)
	}
}

func TestLuaEngineTestDefaultsToScratchSurface(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))

	run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
		Engine: engines.EngineSpec{
			Type:   "lua",
			Script: `return { surface = context.surface, dry_run = context.dry_run }`,
		},
	})
	if err != nil {
		t.Fatalf("test scratch lua engine: %v", err)
	}
	output, _ := run.Output["output"].(map[string]any)
	if output["surface"] != "scratch" || output["dry_run"] != true {
		t.Fatalf("expected scratch dry-run output, got %#v", run.Output)
	}
	inputEnvelope, _ := run.Input["input"].(map[string]any)
	contextEnvelope, _ := inputEnvelope["context"].(map[string]any)
	if inputEnvelope["dry_run"] != true || contextEnvelope["surface"] != "scratch" {
		t.Fatalf("expected scratch run input context, got %#v", run.Input)
	}
}

func TestEngineValidateOnlyDoesNotExecute(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))

	run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
		Surface:      "custom_create_page",
		ValidateOnly: true,
		Engine: engines.EngineSpec{
			Type:   "lua",
			Script: `error("must not execute")`,
		},
	})
	if err != nil {
		t.Fatalf("validate engine: %v", err)
	}
	if run.Status != automation.StatusSucceeded {
		t.Fatalf("expected succeeded validation run, got %#v", run)
	}
	output, _ := run.Output["output"].(map[string]any)
	if output["validated"] != true || output["mode"] != "validated" || output["engine_type"] != "lua" || output["surface"] != "custom_create_page" {
		t.Fatalf("unexpected validate-only output: %#v", run.Output)
	}
	inputEnvelope, _ := run.Input["input"].(map[string]any)
	contextEnvelope, _ := inputEnvelope["context"].(map[string]any)
	if inputEnvelope["validate_only"] != true || contextEnvelope["validate_only"] != true {
		t.Fatalf("expected validate-only run input, got %#v", run.Input)
	}
}

func TestWASMEngineTestReturnsOutputAndLogs(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))

	module := tinyWASIBase64(`{"ok":true,"value":"wasm","surface":"ticket_hook_before"}`+"\n", "wasm checked\n")

	run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
		Surface: "ticket_hook_before",
		Input:   map[string]any{"title": "WASM Preview"},
		Engine: engines.EngineSpec{
			Type:         engines.EngineWASM,
			ModuleBase64: module,
		},
	})
	if err != nil {
		t.Fatalf("test wasm engine: %v", err)
	}
	if run.Status != automation.StatusSucceeded {
		t.Fatalf("expected succeeded run, got %#v", run)
	}
	output, _ := run.Output["output"].(map[string]any)
	if output["ok"] != true || output["value"] != "wasm" || output["surface"] != "ticket_hook_before" {
		t.Fatalf("unexpected wasm output: %#v", run.Output)
	}
	logs, _ := run.Output["logs"].([]any)
	if !slices.Equal(anyStrings(logs), []string{"wasm checked"}) {
		t.Fatalf("unexpected wasm logs: %#v", run.Output)
	}
	if encoded := run.Input; encoded["engine"] != "wasm" {
		t.Fatalf("expected run input to store only engine type, got %#v", encoded)
	}
}

func TestWASMEngineTestRejectsInvalidModuleAndOutput(t *testing.T) {
	tests := []struct {
		name   string
		module string
	}{
		{name: "not_base64", module: "not base64"},
		{name: "not_wasm", module: base64.StdEncoding.EncodeToString([]byte("nope"))},
		{name: "bad_output", module: tinyWASIBase64("not-json\n", "")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := openTestDB(t, ctx)
			seedUser(t, ctx, db.SQL, "user-admin")
			evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
			service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))
			run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
				Engine: engines.EngineSpec{
					Type:         engines.EngineWASM,
					ModuleBase64: tt.module,
				},
			})
			if !errors.Is(err, engines.ErrValidation) {
				t.Fatalf("expected wasm validation, got run=%#v err=%v", run, err)
			}
			if tt.name == "bad_output" && (run.ID == "" || run.Status != automation.StatusFailed) {
				t.Fatalf("expected bad output to persist failed run, got %#v", run)
			}
		})
	}
}

func TestLuaEngineTestValidatesCustomCreatePageOutput(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))

	run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
		Surface: "custom_create_page",
		Engine: engines.EngineSpec{
			Type: "lua",
			Script: `
return {
  field_layout = {
    { key = "title", type = "text", required = true },
    { key = "priority", type = "single-select", options = { "Low", "High" } },
  },
  defaults = { priority = "High" },
  description = "Dynamic form"
}
`,
		},
	})
	if err != nil {
		t.Fatalf("test custom create page engine: %v", err)
	}
	if run.Status != automation.StatusSucceeded {
		t.Fatalf("expected succeeded run, got %#v", run)
	}
	output, _ := run.Output["output"].(map[string]any)
	layout, _ := output["field_layout"].([]any)
	if len(layout) != 2 || output["description"] != "Dynamic form" {
		t.Fatalf("unexpected custom create page output: %#v", output)
	}
}

func TestLuaEngineTestRejectsInvalidCustomCreatePageOutput(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{name: "raw_html", script: `return { field_layout = { { html = "<strong>no</strong>" } } }`},
		{name: "layout_string", script: `return { field_layout = "bad" }`},
		{name: "layout_item_string", script: `return { field_layout = { "bad" } }`},
		{name: "defaults_string", script: `return { defaults = "bad" }`},
		{name: "description_table", script: `return { description = { text = "bad" } }`},
		{name: "unknown_only", script: `return { ok = true }`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := openTestDB(t, ctx)
			seedUser(t, ctx, db.SQL, "user-admin")
			evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
			service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))

			run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
				Surface: "custom_create_page",
				Engine: engines.EngineSpec{
					Type:   "lua",
					Script: tt.script,
				},
			})
			if !errors.Is(err, engines.ErrValidation) {
				t.Fatalf("expected validation error, got run=%#v err=%v", run, err)
			}
			if run.ID == "" || run.Status != automation.StatusFailed || !strings.Contains(run.Error, "Invalid custom create page output") {
				t.Fatalf("expected failed validation run, got run=%#v err=%v", run, err)
			}
			output, _ := run.Output["output"].(map[string]any)
			if len(output) == 0 {
				t.Fatalf("expected failed run to retain output preview, got %#v", run.Output)
			}
		})
	}
}

func TestLuaEngineTestRecordsFailureAsRunStatus(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-admin")
	evaluator := authz.NewInMemoryEvaluator(authz.WithBindings(authz.UserBinding("user-admin", authz.RoleGlobalAdmin, authz.GlobalScope())))
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL, automation.WithNow(fixedNow)))

	run, err := service.Test(ctx, principal("user-admin"), engines.TestInput{
		Engine: engines.EngineSpec{Type: "lua", Script: `error("boom")`},
	})
	if err == nil {
		t.Fatal("expected runtime error")
	}
	if run.ID == "" || run.Status != automation.StatusFailed || run.Error == "" {
		t.Fatalf("expected failed run with error, got run=%#v err=%v", run, err)
	}
}

func TestEngineTestRequiresAutomationPermission(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t, ctx)
	seedUser(t, ctx, db.SQL, "user-member")
	evaluator := authz.NewInMemoryEvaluator()
	service := engines.NewService(db.SQL, evaluator, automation.NewRunStore(db.SQL))

	_, err := service.Test(ctx, principal("user-member"), engines.TestInput{
		Engine: engines.EngineSpec{Type: "lua", Script: `return { ok = true }`},
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
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

func principal(userID string) authz.Principal {
	return authz.Principal{UserID: userID, ActorUserID: userID, AuthKind: authz.AuthKindSession}
}

func fixedNow() time.Time {
	return time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
}

func anyStrings(values []any) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		text, _ := value.(string)
		out = append(out, text)
	}
	return out
}

func tinyWASIBase64(stdout string, stderr string) string {
	return base64.StdEncoding.EncodeToString(tinyWASIModule(stdout, stderr))
}

func tinyWASIModule(stdout string, stderr string) []byte {
	var module []byte
	module = append(module, 0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00)
	module = append(module, wasmSection(1, wasmBytes(
		wasmU32(2),
		[]byte{0x60, 0x04, 0x7f, 0x7f, 0x7f, 0x7f, 0x01, 0x7f},
		[]byte{0x60, 0x00, 0x00},
	))...)
	module = append(module, wasmSection(2, wasmBytes(
		wasmU32(1),
		wasmName("wasi_snapshot_preview1"),
		wasmName("fd_write"),
		[]byte{0x00},
		wasmU32(0),
	))...)
	module = append(module, wasmSection(3, wasmBytes(wasmU32(1), wasmU32(1)))...)
	module = append(module, wasmSection(5, wasmBytes(wasmU32(1), []byte{0x00}, wasmU32(1)))...)
	module = append(module, wasmSection(7, wasmBytes(
		wasmU32(2),
		wasmName("memory"), []byte{0x02}, wasmU32(0),
		wasmName("_start"), []byte{0x00}, wasmU32(1),
	))...)
	var code []byte
	if stdout != "" {
		code = append(code, wasmFDWriteDefaultPtr(1, len(stdout), 1024)...)
	}
	if stderr != "" {
		code = append(code, wasmFDWrite(2, 2048, len(stderr), 1032)...)
	}
	body := wasmBytes(wasmU32(0), code, []byte{0x0b})
	module = append(module, wasmSection(10, wasmBytes(wasmU32(1), wasmU32(uint32(len(body))), body))...)
	module = append(module, wasmSection(11, wasmBytes(
		wasmU32(2),
		wasmData(0, []byte(stdout)),
		wasmData(2048, []byte(stderr)),
	))...)
	return module
}

func wasmFDWrite(fd int, ptr int, size int, iovec int) []byte {
	return wasmBytes(
		[]byte{0x41}, wasmU32(uint32(iovec)), []byte{0x41}, wasmU32(uint32(ptr)), []byte{0x36, 0x02, 0x00},
		[]byte{0x41}, wasmU32(uint32(iovec+4)), []byte{0x41}, wasmU32(uint32(size)), []byte{0x36, 0x02, 0x00},
		[]byte{0x41}, wasmU32(uint32(fd)),
		[]byte{0x41}, wasmU32(uint32(iovec)),
		[]byte{0x41}, wasmU32(1),
		[]byte{0x41}, wasmU32(160),
		[]byte{0x10}, wasmU32(0),
		[]byte{0x1a},
	)
}

func wasmFDWriteDefaultPtr(fd int, size int, iovec int) []byte {
	return wasmBytes(
		[]byte{0x41}, wasmU32(uint32(iovec+4)), []byte{0x41}, wasmU32(uint32(size)), []byte{0x36, 0x02, 0x00},
		[]byte{0x41}, wasmU32(uint32(fd)),
		[]byte{0x41}, wasmU32(uint32(iovec)),
		[]byte{0x41}, wasmU32(1),
		[]byte{0x41}, wasmU32(160),
		[]byte{0x10}, wasmU32(0),
		[]byte{0x1a},
	)
}

func wasmData(offset int, data []byte) []byte {
	return wasmBytes([]byte{0x00, 0x41}, wasmU32(uint32(offset)), []byte{0x0b}, wasmU32(uint32(len(data))), data)
}

func wasmSection(id byte, payload []byte) []byte {
	return wasmBytes([]byte{id}, wasmU32(uint32(len(payload))), payload)
}

func wasmName(value string) []byte {
	return wasmBytes(wasmU32(uint32(len(value))), []byte(value))
}

func wasmBytes(parts ...[]byte) []byte {
	var out []byte
	for _, part := range parts {
		out = append(out, part...)
	}
	return out
}

func wasmU32(value uint32) []byte {
	var out []byte
	for {
		b := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			b |= 0x80
		}
		out = append(out, b)
		if value == 0 {
			return out
		}
	}
}
