package engines

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/luasandbox"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	lua "github.com/yuin/gopher-lua"
)

const (
	EngineLua = "lua"
	EngineAI  = "ai"
)

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
	Type       string `json:"type"`
	Script     string `json:"script,omitempty"`
	Prompt     string `json:"prompt,omitempty"`
	ProviderID string `json:"provider_id,omitempty"`
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
		input.Surface = "generic"
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
		fields["surface"] = "Must be generic, cron, ticket_hook_before, ticket_hook_after, custom_create_page, incoming_webhook, outgoing_webhook, or notification_hook"
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
	default:
		fields["engine.type"] = "Must be lua or ai"
	}
	if len(fields) > 0 {
		return &ValidationError{Message: "Invalid engine test", Fields: fields}
	}
	return nil
}

func validSurface(surface string) bool {
	switch surface {
	case "generic",
		"cron",
		"ticket_hook",
		"ticket_hook_before",
		"ticket_hook_after",
		"webhook",
		"incoming_webhook",
		"outgoing_webhook",
		"notification_hook",
		"create_page",
		"custom_create_page":
		return true
	default:
		return false
	}
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
	case EngineAI:
		engine.Script = ""
	}
	return engine
}
