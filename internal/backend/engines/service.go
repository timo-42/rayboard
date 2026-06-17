package engines

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/luasandbox"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	lua "github.com/yuin/gopher-lua"
)

const (
	EngineLua  = "lua"
	EngineAI   = "ai"
	EngineWASM = "wasm"
)

const maxWASMModuleBytes = 4 << 20
const maxWASMOutputBytes = 1 << 20

var (
	ErrValidation = errors.New("engines: validation failed")
	ErrNotFound   = errors.New("engines: not found")
)

type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	if e == nil || e.Message == "" {
		return ErrValidation.Error()
	}
	return fmt.Sprintf("%s: %s", ErrValidation, e.Message)
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation
}

type EngineSpec struct {
	Type         string `json:"type"`
	Script       string `json:"script,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	ProviderID   string `json:"provider_id,omitempty"`
	ModuleBase64 string `json:"module_base64,omitempty"`
}

type TestInput struct {
	ProjectID   string
	ActorUserID string
	Surface     string
	Context     map[string]any
	Input       map[string]any
	DryRun      bool
	Engine      EngineSpec
}

type Service struct {
	db         *sql.DB
	authorizer authz.Evaluator
	runs       *automation.RunStore
	openrouter *openrouter.Service
	now        func() time.Time
}

type Option func(*Service)

func NewService(db *sql.DB, authorizer authz.Evaluator, runStore *automation.RunStore, options ...Option) *Service {
	service := &Service{
		db:         db,
		authorizer: authorizer,
		runs:       runStore,
		now:        func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithNow(now func() time.Time) Option {
	return func(service *Service) {
		if now != nil {
			service.now = now
		}
	}
}

func WithOpenRouterService(openRouterService *openrouter.Service) Option {
	return func(service *Service) {
		service.openrouter = openRouterService
	}
}

func (s *Service) Test(ctx context.Context, principal authz.Principal, input TestInput) (automation.Run, error) {
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	input.ActorUserID = strings.TrimSpace(input.ActorUserID)
	if input.ActorUserID == "" {
		input.ActorUserID = principal.UserID
	}
	input.Surface = strings.TrimSpace(input.Surface)
	if input.Surface == "" {
		input.Surface = "scratch"
	}
	if input.Context == nil {
		input.Context = map[string]any{}
	}
	input.Engine = normalizeEngine(input.Engine)
	if input.Input == nil {
		input.Input = map[string]any{}
	}
	input.DryRun = true

	if err := s.validate(ctx, input); err != nil {
		return automation.Run{}, err
	}
	if err := s.requireManage(principal, input.ProjectID); err != nil {
		return automation.Run{}, err
	}

	run, err := s.runs.Start(ctx, automation.StartInput{
		TriggerType: "engine_test",
		ProjectID:   input.ProjectID,
		Engine:      input.Engine.Type,
		ActorUserID: input.ActorUserID,
		Input: map[string]any{
			"surface": input.Surface,
			"context": input.normalizedContext(),
			"input":   input.Input,
			"dry_run": input.DryRun,
		},
		Limits: s.runLimits(ctx, input),
	})
	if err != nil {
		return automation.Run{}, err
	}

	output, logs, execErr := s.execute(ctx, input)
	if execErr == nil {
		execErr = validateSurfaceOutput(input, output)
	}
	finish := automation.FinishInput{
		Status: automation.StatusSucceeded,
		Output: output,
		Logs:   logs,
	}
	if execErr != nil {
		finish.Status = automation.StatusFailed
		finish.Error = execErr.Error()
	}
	finished, finishErr := s.runs.Finish(ctx, run.ID, finish)
	if finishErr != nil {
		return automation.Run{}, finishErr
	}
	if execErr != nil {
		return finished, execErr
	}
	return finished, nil
}

func (s *Service) execute(ctx context.Context, input TestInput) (map[string]any, []string, error) {
	switch input.Engine.Type {
	case EngineLua:
		return s.executeLua(ctx, input)
	case EngineAI:
		return s.executeAI(ctx, input)
	case EngineWASM:
		return s.executeWASM(ctx, input)
	default:
		return map[string]any{}, nil, fmt.Errorf("%w: unsupported engine", ErrValidation)
	}
}

func (s *Service) executeLua(ctx context.Context, input TestInput) (map[string]any, []string, error) {
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	sandbox := luasandbox.New(luasandbox.DefaultJSONLimits())
	defer sandbox.Close()
	sandbox.L.SetContext(runCtx)

	logs := []string{}
	registerLuaHelpers(sandbox, input, &logs)
	if err := sandbox.L.DoString(input.Engine.Script); err != nil {
		return map[string]any{}, logs, err
	}
	result := sandbox.L.Get(-1)
	if result == lua.LNil {
		return map[string]any{"ok": true}, logs, nil
	}
	converted, err := sandbox.JSON.ToGo(result)
	if err != nil {
		return map[string]any{}, logs, err
	}
	output, ok := converted.(map[string]any)
	if !ok {
		return map[string]any{}, logs, fmt.Errorf("%w: Lua engine test must return a table/object", ErrValidation)
	}
	return output, logs, nil
}

func (s *Service) executeAI(ctx context.Context, input TestInput) (map[string]any, []string, error) {
	if s.openrouter == nil {
		return map[string]any{}, nil, fmt.Errorf("%w: OpenRouter service is not configured", ErrValidation)
	}
	prompt, err := engineTestPrompt(input)
	if err != nil {
		return map[string]any{}, nil, err
	}
	result, err := s.openrouter.CompleteJSON(ctx, openrouter.CompletionInput{
		ProviderID: input.Engine.ProviderID,
		Prompt:     prompt,
	})
	if err != nil {
		return map[string]any{}, nil, err
	}
	output := map[string]any{
		"ok":          true,
		"provider_id": result.ProviderID,
		"model":       result.Model,
		"output":      result.Output,
	}
	if result.ResponseID != "" {
		output["response_id"] = result.ResponseID
	}
	if len(result.Usage) > 0 {
		output["usage"] = result.Usage
	}
	return output, nil, nil
}

func (s *Service) executeWASM(ctx context.Context, input TestInput) (map[string]any, []string, error) {
	moduleBytes, err := decodeWASMModule(input.Engine.ModuleBase64)
	if err != nil {
		return map[string]any{}, nil, err
	}
	stdin, err := json.Marshal(map[string]any{
		"surface": input.Surface,
		"context": input.normalizedContext(),
		"input":   input.Input,
		"dry_run": true,
	})
	if err != nil {
		return map[string]any{}, nil, fmt.Errorf("encode wasm engine input: %w", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	runtime := wazero.NewRuntimeWithConfig(runCtx, wazero.NewRuntimeConfigInterpreter())
	defer runtime.Close(runCtx)
	wasi_snapshot_preview1.MustInstantiate(runCtx, runtime)

	stdout := &limitedBuffer{limit: maxWASMOutputBytes}
	stderr := &limitedBuffer{limit: maxWASMOutputBytes}
	config := wazero.NewModuleConfig().
		WithStdin(bytes.NewReader(stdin)).
		WithStdout(stdout).
		WithStderr(stderr)
	if _, err := runtime.InstantiateWithConfig(runCtx, moduleBytes, config); err != nil {
		return map[string]any{}, wasmLogs(stderr.String()), fmt.Errorf("%w: run wasm module: %v", ErrValidation, err)
	}

	outputText := strings.TrimSpace(stdout.String())
	if outputText == "" {
		return map[string]any{}, wasmLogs(stderr.String()), fmt.Errorf("%w: wasm stdout must be a JSON object", ErrValidation)
	}
	var output map[string]any
	decoder := json.NewDecoder(io.LimitReader(strings.NewReader(outputText), 1<<20))
	if err := decoder.Decode(&output); err != nil {
		return map[string]any{}, wasmLogs(stderr.String()), fmt.Errorf("%w: wasm stdout must be a JSON object: %v", ErrValidation, err)
	}
	if output == nil {
		output = map[string]any{}
	}
	return output, wasmLogs(stderr.String()), nil
}

func registerLuaHelpers(sandbox *luasandbox.Sandbox, input TestInput, logs *[]string) {
	L := sandbox.L
	contextValue, err := sandbox.JSON.FromGo(input.normalizedContext())
	if err == nil {
		L.SetGlobal("context", contextValue)
	}
	inputValue, err := sandbox.JSON.FromGo(input.Input)
	if err == nil {
		L.SetGlobal("input", inputValue)
	}
	rayboard := L.GetGlobal("rayboard")
	table, ok := rayboard.(*lua.LTable)
	if !ok {
		table = L.NewTable()
		L.SetGlobal("rayboard", table)
	}
	L.SetField(table, "log", L.NewFunction(func(L *lua.LState) int {
		message := strings.TrimSpace(L.CheckString(1))
		if message != "" && len(*logs) < 200 {
			*logs = append(*logs, message)
		}
		return 0
	}))
}

func (s *Service) validate(ctx context.Context, input TestInput) error {
	fields := map[string]string{}
	if input.ActorUserID == "" {
		fields["actor_user_id"] = "Required"
	} else if err := s.actorCanRun(ctx, input.ActorUserID); err != nil {
		fields["actor_user_id"] = "Actor must exist and be enabled"
	}
	if !validSurface(input.Surface) {
		fields["surface"] = "Must be scratch, cron, ticket_hook_before, ticket_hook_after, custom_create_page, incoming_webhook, outgoing_webhook, or notification_hook"
	}
	switch input.Engine.Type {
	case EngineLua:
		if strings.TrimSpace(input.Engine.Script) == "" {
			fields["engine.script"] = "Required for lua engine"
		}
	case EngineAI:
		if strings.TrimSpace(input.Engine.Prompt) == "" {
			fields["engine.prompt"] = "Required for ai engine"
		}
		if strings.TrimSpace(input.Engine.ProviderID) == "" {
			fields["engine.provider_id"] = "Required for ai engine"
		} else if err := s.validateAIProvider(ctx, input.Engine.ProviderID); err != nil {
			fields["engine.provider_id"] = err.Error()
		}
	case EngineWASM:
		if strings.TrimSpace(input.Engine.ModuleBase64) == "" {
			fields["engine.module_base64"] = "Required for wasm engine"
		} else if _, err := decodeWASMModule(input.Engine.ModuleBase64); err != nil {
			fields["engine.module_base64"] = err.Error()
		}
	default:
		fields["engine.type"] = "Must be lua, ai, or wasm"
	}
	if len(fields) > 0 {
		return &ValidationError{Message: "Invalid engine test", Fields: fields}
	}
	return nil
}

func validSurface(surface string) bool {
	switch surface {
	case "scratch",
		"cron",
		"ticket_hook_before",
		"ticket_hook_after",
		"incoming_webhook",
		"outgoing_webhook",
		"notification_hook",
		"custom_create_page":
		return true
	default:
		return false
	}
}

func validateSurfaceOutput(input TestInput, output map[string]any) error {
	switch input.Surface {
	case "custom_create_page":
		return validateCustomCreatePageOutput(surfaceOutput(input, output))
	default:
		return nil
	}
}

func surfaceOutput(input TestInput, output map[string]any) map[string]any {
	if input.Engine.Type == EngineAI {
		if nested, ok := output["output"].(map[string]any); ok {
			return nested
		}
		return map[string]any{}
	}
	return output
}

func validateCustomCreatePageOutput(output map[string]any) error {
	fields := map[string]string{}
	recognized := false
	if rawLayout, ok := output["field_layout"]; ok {
		recognized = true
		if err := validateCreatePageFieldLayout(rawLayout); err != nil {
			fields["field_layout"] = err.Error()
		}
	}
	if rawDefaults, ok := output["defaults"]; ok {
		recognized = true
		if _, ok := rawDefaults.(map[string]any); !ok {
			fields["defaults"] = "Must be an object"
		}
	}
	if rawDescription, ok := output["description"]; ok {
		recognized = true
		if _, ok := rawDescription.(string); !ok {
			fields["description"] = "Must be a string"
		}
	}
	if !recognized {
		fields["output"] = "Must include field_layout, defaults, or description"
	}
	if len(fields) > 0 {
		return &ValidationError{Message: "Invalid custom create page output", Fields: fields}
	}
	return nil
}

func validateCreatePageFieldLayout(value any) error {
	items, ok := value.([]any)
	if !ok {
		return errors.New("Must be an array of objects")
	}
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			return errors.New("Must be an array of objects")
		}
		if _, hasHTML := object["html"]; hasHTML {
			return errors.New("Raw HTML fields are not allowed")
		}
		if nested, ok := object["fields"]; ok {
			if err := validateCreatePageFieldLayout(nested); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) validateAIProvider(ctx context.Context, providerID string) error {
	if s.openrouter == nil {
		return errors.New("OpenRouter service is not configured")
	}
	provider, err := s.openrouter.GetExecutionProvider(ctx, providerID)
	if err != nil {
		return err
	}
	if !provider.Enabled {
		return errors.New("OpenRouter provider is disabled")
	}
	if strings.TrimSpace(provider.APIKey) == "" {
		return errors.New("OpenRouter provider API key is not configured")
	}
	if strings.TrimSpace(provider.DefaultModel) == "" {
		return errors.New("OpenRouter provider default model is required")
	}
	if provider.DefaultTimeoutSeconds <= 0 {
		return errors.New("OpenRouter provider timeout must be greater than zero")
	}
	if provider.MaxOutputTokens <= 0 {
		return errors.New("OpenRouter provider max output tokens must be greater than zero")
	}
	return nil
}

func (s *Service) actorCanRun(ctx context.Context, userID string) error {
	var disabled bool
	if err := s.db.QueryRowContext(ctx, `
		SELECT is_disabled
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`, userID).Scan(&disabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("get engine test actor: %w", err)
	}
	if disabled {
		return fmt.Errorf("%w: actor is disabled", ErrValidation)
	}
	return nil
}

func (s *Service) requireManage(principal authz.Principal, projectID string) error {
	if projectID != "" {
		return s.authorizer.Require(principal, authz.PermissionAutomationsManage, authz.ProjectScope(projectID))
	}
	return s.authorizer.Require(principal, authz.PermissionAutomationsManage, authz.GlobalScope())
}

func (s *Service) runLimits(ctx context.Context, input TestInput) map[string]any {
	limits := map[string]any{
		"timeout_seconds": 30,
		"dry_run":         true,
	}
	if input.Engine.Type == EngineWASM {
		limits["max_module_bytes"] = maxWASMModuleBytes
		limits["max_stdout_bytes"] = maxWASMOutputBytes
		limits["max_stderr_bytes"] = maxWASMOutputBytes
		return limits
	}
	if input.Engine.Type != EngineAI || s.openrouter == nil {
		return limits
	}
	provider, err := s.openrouter.GetExecutionProvider(ctx, input.Engine.ProviderID)
	if err != nil {
		return limits
	}
	limits["timeout_seconds"] = provider.DefaultTimeoutSeconds
	limits["max_output_tokens"] = provider.MaxOutputTokens
	limits["provider_id"] = provider.ID
	limits["model"] = provider.DefaultModel
	return limits
}

func engineTestPrompt(input TestInput) (string, error) {
	payload, err := json.Marshal(map[string]any{
		"surface":      input.Surface,
		"context":      input.normalizedContext(),
		"input":        input.Input,
		"instructions": []string{"Return only a JSON object.", "Do not request or assume access to secrets.", "Do not perform mutations; describe intended actions as output data only."},
	})
	if err != nil {
		return "", fmt.Errorf("encode engine test prompt context: %w", err)
	}
	return strings.TrimSpace(input.Engine.Prompt) + "\n\nRayboard engine test input:\n" + string(payload), nil
}

func (input TestInput) normalizedContext() map[string]any {
	context := map[string]any{}
	for key, value := range input.Context {
		context[key] = value
	}
	context["surface"] = input.Surface
	context["project_id"] = input.ProjectID
	context["actor_user_id"] = input.ActorUserID
	context["dry_run"] = true
	return context
}

func normalizeEngine(engine EngineSpec) EngineSpec {
	engine.Type = strings.TrimSpace(engine.Type)
	engine.ProviderID = strings.TrimSpace(engine.ProviderID)
	if engine.Type == "" {
		engine.Type = EngineLua
	}
	switch engine.Type {
	case EngineLua:
		engine.Prompt = ""
		engine.ProviderID = ""
		engine.ModuleBase64 = ""
	case EngineAI:
		engine.Script = ""
		engine.ModuleBase64 = ""
	case EngineWASM:
		engine.Script = ""
		engine.Prompt = ""
		engine.ProviderID = ""
	}
	return engine
}

func decodeWASMModule(encoded string) ([]byte, error) {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return nil, fmt.Errorf("%w: wasm module is required", ErrValidation)
	}
	moduleBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: wasm module must be base64 encoded", ErrValidation)
	}
	if len(moduleBytes) == 0 {
		return nil, fmt.Errorf("%w: wasm module is empty", ErrValidation)
	}
	if len(moduleBytes) > maxWASMModuleBytes {
		return nil, fmt.Errorf("%w: wasm module must be at most %d bytes", ErrValidation, maxWASMModuleBytes)
	}
	if len(moduleBytes) < 4 || string(moduleBytes[:4]) != "\x00asm" {
		return nil, fmt.Errorf("%w: wasm module has invalid magic header", ErrValidation)
	}
	return moduleBytes, nil
}

func wasmLogs(stderr string) []string {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return nil
	}
	lines := strings.Split(stderr, "\n")
	logs := make([]string, 0, min(len(lines), 200))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			logs = append(logs, line)
		}
		if len(logs) >= 200 {
			break
		}
	}
	return logs
}

type limitedBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (w *limitedBuffer) Write(p []byte) (int, error) {
	if w.limit <= 0 {
		return len(p), nil
	}
	remaining := w.limit - w.buffer.Len()
	if remaining <= 0 {
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = w.buffer.Write(p[:remaining])
		return len(p), nil
	}
	_, _ = w.buffer.Write(p)
	return len(p), nil
}

func (w *limitedBuffer) String() string {
	return w.buffer.String()
}
