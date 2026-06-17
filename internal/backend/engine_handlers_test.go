package backend

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/engines"
)

func TestEngineTestEndpoint(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	runStore := automation.NewRunStore(db.SQL)
	authService := auth.NewService(db.SQL)
	actor, err := authService.CreateUser(ctx, auth.CreateUserInput{Username: "engine-owner"})
	if err != nil {
		t.Fatalf("create engine owner: %v", err)
	}
	handler := NewHandler(
		WithAuthService(authService),
		WithAuthorizer(authorizer),
		WithEngineService(engines.NewService(db.SQL, authorizer, runStore)),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	sessionCookie := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrfCookie := responseCookie(t, login.Result(), csrfCookieName)

	missingCSRF := postJSON(t, handler, "/api/engines/test", map[string]any{
		"spec": map[string]any{
			"engine": map[string]string{"type": "lua", "script": `return { ok = true }`},
		},
	}, []*http.Cookie{sessionCookie})
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing CSRF status 403, got %d: %s", missingCSRF.Code, missingCSRF.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "/api/engines/test", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"surface": "ticket_hook_before",
			"context": map[string]any{"ticket_id": "ticket-1"},
			"input":   map[string]string{"title": "Preview"},
			"dry_run": false,
			"engine": map[string]string{
				"type":   "lua",
				"script": `rayboard.log("preview " .. input.title); return { ok = true, title = input.title, surface = context.surface, ticket_id = context.ticket_id, dry_run = context.dry_run }`,
			},
		},
	}))
	req.AddCookie(sessionCookie)
	req.AddCookie(csrfCookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrfCookie.Value)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected engine test status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Metadata struct {
			ID string `json:"id"`
		} `json:"metadata"`
		Spec struct {
			Engine struct {
				Type   string `json:"type"`
				Script string `json:"script"`
			} `json:"engine"`
			Surface string         `json:"surface"`
			Context map[string]any `json:"context"`
			DryRun  bool           `json:"dry_run"`
		} `json:"spec"`
		Status struct {
			State  string         `json:"state"`
			Output map[string]any `json:"output"`
			Logs   []string       `json:"logs"`
			Engine map[string]any `json:"engine"`
		} `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode engine test response: %v", err)
	}
	if body.Metadata.ID == "" || body.Status.State != automation.StatusSucceeded {
		t.Fatalf("unexpected response metadata/status: %#v", body)
	}
	if body.Spec.Engine.Type != "lua" || body.Spec.Engine.Script != "" {
		t.Fatalf("expected engine source to be redacted, got %#v", body.Spec.Engine)
	}
	if body.Spec.Surface != "ticket_hook_before" || body.Spec.Context["ticket_id"] != "ticket-1" || !body.Spec.DryRun {
		t.Fatalf("expected normalized engine test spec, got %#v", body.Spec)
	}
	if body.Status.Output["ok"] != true || body.Status.Output["title"] != "Preview" || body.Status.Output["surface"] != "ticket_hook_before" || body.Status.Output["ticket_id"] != "ticket-1" || body.Status.Output["dry_run"] != true {
		t.Fatalf("unexpected engine output: %#v", body.Status.Output)
	}
	if len(body.Status.Logs) != 1 || body.Status.Logs[0] != "preview Preview" || body.Status.Engine["type"] != "lua" || body.Status.Engine["surface"] != "ticket_hook_before" || body.Status.Engine["dry_run"] != true {
		t.Fatalf("unexpected engine status metadata: %#v", body.Status)
	}

	scratchReq := httptest.NewRequest(http.MethodPost, "/api/engines/test", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"engine": map[string]string{
				"type":   "lua",
				"script": `return { surface = context.surface, value = input.value }`,
			},
			"input": map[string]string{"value": "scratch"},
		},
	}))
	scratchReq.AddCookie(sessionCookie)
	scratchReq.AddCookie(csrfCookie)
	scratchReq.Header.Set("Content-Type", "application/json")
	scratchReq.Header.Set("X-CSRF-Token", csrfCookie.Value)
	scratchRec := httptest.NewRecorder()
	handler.ServeHTTP(scratchRec, scratchReq)
	if scratchRec.Code != http.StatusOK {
		t.Fatalf("expected scratch engine test status 200, got %d: %s", scratchRec.Code, scratchRec.Body.String())
	}
	var scratchBody struct {
		Spec struct {
			Surface string `json:"surface"`
		} `json:"spec"`
		Status struct {
			Output map[string]any `json:"output"`
			Engine map[string]any `json:"engine"`
		} `json:"status"`
	}
	if err := json.Unmarshal(scratchRec.Body.Bytes(), &scratchBody); err != nil {
		t.Fatalf("decode scratch engine test response: %v", err)
	}
	if scratchBody.Spec.Surface != "scratch" || scratchBody.Status.Output["surface"] != "scratch" || scratchBody.Status.Output["value"] != "scratch" || scratchBody.Status.Engine["surface"] != "scratch" {
		t.Fatalf("expected scratch engine response, got %#v", scratchBody)
	}

	contextActorReq := httptest.NewRequest(http.MethodPost, "/api/engines/test", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"context": map[string]any{"actor_user_id": actor.ID},
			"engine": map[string]string{
				"type":   "lua",
				"script": `return { actor_user_id = context.actor_user_id }`,
			},
		},
	}))
	contextActorReq.AddCookie(sessionCookie)
	contextActorReq.AddCookie(csrfCookie)
	contextActorReq.Header.Set("Content-Type", "application/json")
	contextActorReq.Header.Set("X-CSRF-Token", csrfCookie.Value)
	contextActorRec := httptest.NewRecorder()
	handler.ServeHTTP(contextActorRec, contextActorReq)
	if contextActorRec.Code != http.StatusOK {
		t.Fatalf("expected context actor engine test status 200, got %d: %s", contextActorRec.Code, contextActorRec.Body.String())
	}
	var contextActorBody struct {
		Spec struct {
			ActorUserID string         `json:"actor_user_id"`
			Context     map[string]any `json:"context"`
		} `json:"spec"`
		Status struct {
			Output map[string]any `json:"output"`
			Engine map[string]any `json:"engine"`
		} `json:"status"`
	}
	if err := json.Unmarshal(contextActorRec.Body.Bytes(), &contextActorBody); err != nil {
		t.Fatalf("decode context actor engine response: %v", err)
	}
	if contextActorBody.Spec.ActorUserID != actor.ID || contextActorBody.Spec.Context["actor_user_id"] != actor.ID || contextActorBody.Status.Output["actor_user_id"] != actor.ID {
		t.Fatalf("expected actor fallback from context, got %#v", contextActorBody)
	}

	validateReq := httptest.NewRequest(http.MethodPost, "/api/engines/test", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"surface":       "custom_create_page",
			"validate_only": true,
			"engine": map[string]string{
				"type":   "lua",
				"script": `error("must not execute")`,
			},
		},
	}))
	validateReq.AddCookie(sessionCookie)
	validateReq.AddCookie(csrfCookie)
	validateReq.Header.Set("Content-Type", "application/json")
	validateReq.Header.Set("X-CSRF-Token", csrfCookie.Value)
	validateRec := httptest.NewRecorder()
	handler.ServeHTTP(validateRec, validateReq)
	if validateRec.Code != http.StatusOK {
		t.Fatalf("expected validate-only engine status 200, got %d: %s", validateRec.Code, validateRec.Body.String())
	}
	var validateBody struct {
		Spec struct {
			ValidateOnly bool `json:"validate_only"`
			Engine       struct {
				Script string `json:"script"`
			} `json:"engine"`
		} `json:"spec"`
		Status struct {
			State  string         `json:"state"`
			Mode   string         `json:"mode"`
			Output map[string]any `json:"output"`
			Engine map[string]any `json:"engine"`
		} `json:"status"`
	}
	if err := json.Unmarshal(validateRec.Body.Bytes(), &validateBody); err != nil {
		t.Fatalf("decode validate-only engine response: %v", err)
	}
	if !validateBody.Spec.ValidateOnly || validateBody.Spec.Engine.Script != "" {
		t.Fatalf("expected validate-only redacted spec, got %#v", validateBody.Spec)
	}
	if validateBody.Status.State != automation.StatusSucceeded || validateBody.Status.Mode != "validated" || validateBody.Status.Output["validated"] != true || validateBody.Status.Engine["validate_only"] != true {
		t.Fatalf("unexpected validate-only response: %#v", validateBody.Status)
	}

	wasmModule := tinyEngineHandlerWASIBase64(`{"ok":true,"value":"wasm","surface":"scratch"}`+"\n", "wasm preview\n")
	wasmReq := httptest.NewRequest(http.MethodPost, "/api/engines/test", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"engine": map[string]string{
				"type":          "wasm",
				"module_base64": wasmModule,
			},
			"input": map[string]string{"value": "wasm"},
		},
	}))
	wasmReq.AddCookie(sessionCookie)
	wasmReq.AddCookie(csrfCookie)
	wasmReq.Header.Set("Content-Type", "application/json")
	wasmReq.Header.Set("X-CSRF-Token", csrfCookie.Value)
	wasmRec := httptest.NewRecorder()
	handler.ServeHTTP(wasmRec, wasmReq)
	if wasmRec.Code != http.StatusOK {
		t.Fatalf("expected wasm engine test status 200, got %d: %s", wasmRec.Code, wasmRec.Body.String())
	}
	var wasmBody struct {
		Spec struct {
			Engine struct {
				Type         string `json:"type"`
				ModuleBase64 string `json:"module_base64"`
			} `json:"engine"`
			Surface string `json:"surface"`
		} `json:"spec"`
		Status struct {
			Output map[string]any `json:"output"`
			Logs   []string       `json:"logs"`
			Engine map[string]any `json:"engine"`
		} `json:"status"`
	}
	if err := json.Unmarshal(wasmRec.Body.Bytes(), &wasmBody); err != nil {
		t.Fatalf("decode wasm engine test response: %v", err)
	}
	if wasmBody.Spec.Engine.Type != "wasm" || wasmBody.Spec.Engine.ModuleBase64 != "" || wasmBody.Spec.Surface != "scratch" {
		t.Fatalf("expected redacted wasm spec, got %#v", wasmBody.Spec)
	}
	if wasmBody.Status.Output["ok"] != true || wasmBody.Status.Output["value"] != "wasm" || wasmBody.Status.Engine["type"] != "wasm" {
		t.Fatalf("unexpected wasm engine output: %#v", wasmBody.Status)
	}
	if len(wasmBody.Status.Logs) != 1 || wasmBody.Status.Logs[0] != "wasm preview" {
		t.Fatalf("unexpected wasm logs: %#v", wasmBody.Status.Logs)
	}

	customReq := httptest.NewRequest(http.MethodPost, "/api/engines/test", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"surface": "custom_create_page",
			"engine": map[string]string{
				"type": "lua",
				"script": `
return {
  field_layout = {
    { key = "title", type = "text", required = true },
    { key = "priority", type = "single-select", options = { "Low", "High" } },
  },
  defaults = { priority = "High" }
}
`,
			},
		},
	}))
	customReq.AddCookie(sessionCookie)
	customReq.AddCookie(csrfCookie)
	customReq.Header.Set("Content-Type", "application/json")
	customReq.Header.Set("X-CSRF-Token", csrfCookie.Value)
	customRec := httptest.NewRecorder()
	handler.ServeHTTP(customRec, customReq)
	if customRec.Code != http.StatusOK {
		t.Fatalf("expected custom create page engine status 200, got %d: %s", customRec.Code, customRec.Body.String())
	}
	var customBody struct {
		Status struct {
			State  string         `json:"state"`
			Output map[string]any `json:"output"`
		} `json:"status"`
	}
	if err := json.Unmarshal(customRec.Body.Bytes(), &customBody); err != nil {
		t.Fatalf("decode custom create page engine response: %v", err)
	}
	layout, _ := customBody.Status.Output["field_layout"].([]any)
	if customBody.Status.State != automation.StatusSucceeded || len(layout) != 2 || customBody.Status.Output["defaults"] == nil {
		t.Fatalf("expected validated custom create page output, got %#v", customBody.Status)
	}

	invalidReq := httptest.NewRequest(http.MethodPost, "/api/engines/test", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"surface": "custom_create_page",
			"engine": map[string]string{
				"type":   "lua",
				"script": `return { field_layout = { { html = "<strong>no</strong>" } } }`,
			},
		},
	}))
	invalidReq.AddCookie(sessionCookie)
	invalidReq.AddCookie(csrfCookie)
	invalidReq.Header.Set("Content-Type", "application/json")
	invalidReq.Header.Set("X-CSRF-Token", csrfCookie.Value)
	invalidRec := httptest.NewRecorder()
	handler.ServeHTTP(invalidRec, invalidReq)
	if invalidRec.Code != http.StatusOK {
		t.Fatalf("expected invalid custom create page engine status 200, got %d: %s", invalidRec.Code, invalidRec.Body.String())
	}
	var invalidBody struct {
		Status struct {
			State string `json:"state"`
			Error string `json:"error"`
		} `json:"status"`
	}
	if err := json.Unmarshal(invalidRec.Body.Bytes(), &invalidBody); err != nil {
		t.Fatalf("decode invalid custom create page engine response: %v", err)
	}
	if invalidBody.Status.State != automation.StatusFailed || !strings.Contains(invalidBody.Status.Error, "Invalid custom create page output") {
		t.Fatalf("expected failed validation response, got %#v", invalidBody.Status)
	}
}

func tinyEngineHandlerWASIBase64(stdout string, stderr string) string {
	return base64.StdEncoding.EncodeToString(tinyEngineHandlerWASIModule(stdout, stderr))
}

func tinyEngineHandlerWASIModule(stdout string, stderr string) []byte {
	var module []byte
	module = append(module, 0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00)
	module = append(module, engineHandlerWASMSection(1, engineHandlerWASMBytes(
		engineHandlerWASMU32(2),
		[]byte{0x60, 0x04, 0x7f, 0x7f, 0x7f, 0x7f, 0x01, 0x7f},
		[]byte{0x60, 0x00, 0x00},
	))...)
	module = append(module, engineHandlerWASMSection(2, engineHandlerWASMBytes(
		engineHandlerWASMU32(1),
		engineHandlerWASMName("wasi_snapshot_preview1"),
		engineHandlerWASMName("fd_write"),
		[]byte{0x00},
		engineHandlerWASMU32(0),
	))...)
	module = append(module, engineHandlerWASMSection(3, engineHandlerWASMBytes(engineHandlerWASMU32(1), engineHandlerWASMU32(1)))...)
	module = append(module, engineHandlerWASMSection(5, engineHandlerWASMBytes(engineHandlerWASMU32(1), []byte{0x00}, engineHandlerWASMU32(1)))...)
	module = append(module, engineHandlerWASMSection(7, engineHandlerWASMBytes(
		engineHandlerWASMU32(2),
		engineHandlerWASMName("memory"), []byte{0x02}, engineHandlerWASMU32(0),
		engineHandlerWASMName("_start"), []byte{0x00}, engineHandlerWASMU32(1),
	))...)
	var code []byte
	if stdout != "" {
		code = append(code, engineHandlerWASMFDWriteDefaultPtr(1, len(stdout), 1024)...)
	}
	if stderr != "" {
		code = append(code, engineHandlerWASMFDWrite(2, 2048, len(stderr), 1032)...)
	}
	body := engineHandlerWASMBytes(engineHandlerWASMU32(0), code, []byte{0x0b})
	module = append(module, engineHandlerWASMSection(10, engineHandlerWASMBytes(engineHandlerWASMU32(1), engineHandlerWASMU32(uint32(len(body))), body))...)
	module = append(module, engineHandlerWASMSection(11, engineHandlerWASMBytes(
		engineHandlerWASMU32(2),
		engineHandlerWASMData(0, []byte(stdout)),
		engineHandlerWASMData(2048, []byte(stderr)),
	))...)
	return module
}

func engineHandlerWASMFDWrite(fd int, ptr int, size int, iovec int) []byte {
	return engineHandlerWASMBytes(
		[]byte{0x41}, engineHandlerWASMU32(uint32(iovec)), []byte{0x41}, engineHandlerWASMU32(uint32(ptr)), []byte{0x36, 0x02, 0x00},
		[]byte{0x41}, engineHandlerWASMU32(uint32(iovec+4)), []byte{0x41}, engineHandlerWASMU32(uint32(size)), []byte{0x36, 0x02, 0x00},
		[]byte{0x41}, engineHandlerWASMU32(uint32(fd)),
		[]byte{0x41}, engineHandlerWASMU32(uint32(iovec)),
		[]byte{0x41}, engineHandlerWASMU32(1),
		[]byte{0x41}, engineHandlerWASMU32(160),
		[]byte{0x10}, engineHandlerWASMU32(0),
		[]byte{0x1a},
	)
}

func engineHandlerWASMFDWriteDefaultPtr(fd int, size int, iovec int) []byte {
	return engineHandlerWASMBytes(
		[]byte{0x41}, engineHandlerWASMU32(uint32(iovec+4)), []byte{0x41}, engineHandlerWASMU32(uint32(size)), []byte{0x36, 0x02, 0x00},
		[]byte{0x41}, engineHandlerWASMU32(uint32(fd)),
		[]byte{0x41}, engineHandlerWASMU32(uint32(iovec)),
		[]byte{0x41}, engineHandlerWASMU32(1),
		[]byte{0x41}, engineHandlerWASMU32(160),
		[]byte{0x10}, engineHandlerWASMU32(0),
		[]byte{0x1a},
	)
}

func engineHandlerWASMData(offset int, data []byte) []byte {
	return engineHandlerWASMBytes([]byte{0x00, 0x41}, engineHandlerWASMU32(uint32(offset)), []byte{0x0b}, engineHandlerWASMU32(uint32(len(data))), data)
}

func engineHandlerWASMSection(id byte, payload []byte) []byte {
	return engineHandlerWASMBytes([]byte{id}, engineHandlerWASMU32(uint32(len(payload))), payload)
}

func engineHandlerWASMName(value string) []byte {
	return engineHandlerWASMBytes(engineHandlerWASMU32(uint32(len(value))), []byte(value))
}

func engineHandlerWASMBytes(parts ...[]byte) []byte {
	var out []byte
	for _, part := range parts {
		out = append(out, part...)
	}
	return out
}

func engineHandlerWASMU32(value uint32) []byte {
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
