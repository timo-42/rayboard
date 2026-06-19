package tracker

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
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
	HookEventTicketCreate = "ticket_create"
	HookEventTicketUpdate = "ticket_update"

	HookPhaseBefore = "before"
	HookPhaseAfter  = "after"

	HookEngineLua = "lua"
	HookEngineAI  = "ai"
)

type HookEngineSpec struct {
	Type       string
	Script     string
	Prompt     string
	ProviderID string
}

type Hook struct {
	ID        string
	ProjectID string
	Name      string
	Event     string
	Phase     string
	Enabled   bool
	Position  int
	Engine    HookEngineSpec
	LastError string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateHookInput struct {
	ProjectID string
	Name      string
	Event     string
	Phase     string
	Enabled   bool
	Position  int
	Engine    HookEngineSpec
}

type ListHooksInput struct {
	ProjectID string
	Event     string
	Phase     string
	Limit     int
	Offset    int
}

type UpdateHookInput struct {
	Name     *string
	Event    *string
	Phase    *string
	Enabled  *bool
	Position *int
	Engine   *HookEngineSpec
}

type PreviewHookInput struct {
	Ticket  map[string]any
	Current map[string]any
}

type HookResult struct {
	Output map[string]any
	Logs   []string
}

type HookPreview struct {
	Hook   Hook
	Input  map[string]any
	Output map[string]any
	Logs   []string
	Error  string
}

type HookService struct {
	db         *sql.DB
	authorizer authz.Evaluator
	runs       *automation.RunStore
	openrouter *openrouter.Service
	now        func() time.Time
}

type HookOption func(*HookService)

func NewHookService(db *sql.DB, authorizer authz.Evaluator, options ...HookOption) *HookService {
	service := &HookService{
		db:         db,
		authorizer: authorizer,
		now:        func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithHookOpenRouterService(openRouterService *openrouter.Service) HookOption {
	return func(service *HookService) {
		service.openrouter = openRouterService
	}
}

func WithHookRunStore(runStore *automation.RunStore) HookOption {
	return func(service *HookService) {
		service.runs = runStore
	}
}

func (s *HookService) Create(ctx context.Context, principal authz.Principal, input CreateHookInput) (Hook, error) {
	hook := Hook{
		ID:        newHookID("hook"),
		ProjectID: strings.TrimSpace(input.ProjectID),
		Name:      normalizeHookName(input.Name),
		Event:     strings.TrimSpace(input.Event),
		Phase:     strings.TrimSpace(input.Phase),
		Enabled:   input.Enabled,
		Position:  input.Position,
		Engine:    normalizeHookEngine(input.Engine),
		CreatedAt: s.now().UTC(),
		UpdatedAt: s.now().UTC(),
	}
	if hook.Position == 0 {
		hook.Position = 100
	}
	if err := s.requireManage(principal, hook.ProjectID); err != nil {
		return Hook{}, err
	}
	if err := s.validate(ctx, hook); err != nil {
		return Hook{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO ticket_hooks (
			id, project_id, name, event, phase, enabled, position, engine_type,
			lua_script, ai_prompt, ai_provider_id, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, hook.ID, hook.ProjectID, hook.Name, hook.Event, hook.Phase, hook.Enabled, hook.Position,
		hook.Engine.Type, nullableHookString(hook.Engine.Script), nullableHookString(hook.Engine.Prompt),
		nullableHookString(hook.Engine.ProviderID), formatHookTime(hook.CreatedAt), formatHookTime(hook.UpdatedAt)); err != nil {
		if isHookUniqueConstraint(err) {
			return Hook{}, validationFailed(map[string]string{"name": "Hook name already exists for this event and phase"})
		}
		return Hook{}, fmt.Errorf("insert ticket hook: %w", err)
	}
	return hook, nil
}

func (s *HookService) List(ctx context.Context, principal authz.Principal, input ListHooksInput) ([]Hook, error) {
	projectID := strings.TrimSpace(input.ProjectID)
	event := strings.TrimSpace(input.Event)
	phase := strings.TrimSpace(input.Phase)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if event != "" && !validHookEvent(event) {
		return nil, validationFailed(map[string]string{"event": "Must be ticket_create or ticket_update"})
	}
	if phase != "" && !validHookPhase(phase) {
		return nil, validationFailed(map[string]string{"phase": "Must be before or after"})
	}
	if err := validateListInput(input.Limit, input.Offset); err != nil {
		return nil, err
	}
	if err := s.requireManage(principal, projectID); err != nil {
		return nil, err
	}
	limit, offset := normalizeListWindow(input.Limit, input.Offset)
	where := []string{"project_id = ?", "deleted_at IS NULL"}
	args := []any{projectID}
	if event != "" {
		where = append(where, "event = ?")
		args = append(args, event)
	}
	if phase != "" {
		where = append(where, "phase = ?")
		args = append(args, phase)
	}
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, name, event, phase, enabled, position, engine_type,
			COALESCE(lua_script, ''), COALESCE(ai_prompt, ''), COALESCE(ai_provider_id, ''),
			COALESCE(last_error, ''), created_at, updated_at
		FROM ticket_hooks
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY event ASC, phase ASC, position ASC, name ASC, id ASC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list ticket hooks: %w", err)
	}
	defer rows.Close()

	var hooks []Hook
	for rows.Next() {
		hook, err := scanHook(rows)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket hooks: %w", err)
	}
	return hooks, nil
}

func (s *HookService) Get(ctx context.Context, principal authz.Principal, hookID string) (Hook, error) {
	hook, err := s.get(ctx, hookID)
	if err != nil {
		return Hook{}, err
	}
	if err := s.requireManage(principal, hook.ProjectID); err != nil {
		return Hook{}, err
	}
	return hook, nil
}

func (s *HookService) Update(ctx context.Context, principal authz.Principal, hookID string, input UpdateHookInput) (Hook, error) {
	current, err := s.get(ctx, hookID)
	if err != nil {
		return Hook{}, err
	}
	if err := s.requireManage(principal, current.ProjectID); err != nil {
		return Hook{}, err
	}
	updated := current
	if input.Name != nil {
		updated.Name = normalizeHookName(*input.Name)
	}
	if input.Event != nil {
		updated.Event = strings.TrimSpace(*input.Event)
	}
	if input.Phase != nil {
		updated.Phase = strings.TrimSpace(*input.Phase)
	}
	if input.Enabled != nil {
		updated.Enabled = *input.Enabled
	}
	if input.Position != nil {
		updated.Position = *input.Position
	}
	if updated.Position == 0 {
		updated.Position = 100
	}
	if input.Engine != nil {
		updated.Engine = normalizeHookEngine(*input.Engine)
	}
	updated.UpdatedAt = s.now().UTC()
	if err := s.validate(ctx, updated); err != nil {
		return Hook{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE ticket_hooks
		SET name = ?, event = ?, phase = ?, enabled = ?, position = ?, engine_type = ?,
			lua_script = ?, ai_prompt = ?, ai_provider_id = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, updated.Name, updated.Event, updated.Phase, updated.Enabled, updated.Position,
		updated.Engine.Type, nullableHookString(updated.Engine.Script), nullableHookString(updated.Engine.Prompt),
		nullableHookString(updated.Engine.ProviderID), formatHookTime(updated.UpdatedAt), updated.ID)
	if err != nil {
		if isHookUniqueConstraint(err) {
			return Hook{}, validationFailed(map[string]string{"name": "Hook name already exists for this event and phase"})
		}
		return Hook{}, fmt.Errorf("update ticket hook: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Hook{}, fmt.Errorf("check ticket hook update: %w", err)
	}
	if affected == 0 {
		return Hook{}, notFound("ticket_hook", hookID)
	}
	return updated, nil
}

func (s *HookService) Delete(ctx context.Context, principal authz.Principal, hookID string) error {
	hook, err := s.get(ctx, hookID)
	if err != nil {
		return err
	}
	if err := s.requireManage(principal, hook.ProjectID); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE ticket_hooks
		SET enabled = 0, deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatHookTime(s.now().UTC()), formatHookTime(s.now().UTC()), hook.ID)
	if err != nil {
		return fmt.Errorf("delete ticket hook: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check ticket hook delete: %w", err)
	}
	if affected == 0 {
		return notFound("ticket_hook", hookID)
	}
	return nil
}

func (s *HookService) Preview(ctx context.Context, principal authz.Principal, hookID string, input PreviewHookInput) (HookPreview, error) {
	hook, err := s.get(ctx, hookID)
	if err != nil {
		return HookPreview{}, err
	}
	if err := s.requireManage(principal, hook.ProjectID); err != nil {
		return HookPreview{}, err
	}
	fields := map[string]string{}
	if input.Ticket == nil {
		fields["ticket"] = "Required"
	}
	if hook.Event == HookEventTicketUpdate && input.Current == nil {
		fields["current"] = "Required for ticket_update hooks"
	}
	if len(fields) > 0 {
		return HookPreview{}, validationFailed(fields)
	}
	executionInput := map[string]any{}
	executionInput["ticket"] = input.Ticket
	if input.Current != nil {
		executionInput["current"] = input.Current
	}
	result, err := s.execute(ctx, principal, hook, copyHookMap(executionInput))
	preview := HookPreview{
		Hook:   hook,
		Input:  executionInput,
		Output: result.Output,
		Logs:   result.Logs,
	}
	if err != nil {
		preview.Error = err.Error()
		if message := hookRejectMessage(result.Output); message != "" {
			preview.Error = message
		}
	}
	if preview.Output == nil {
		preview.Output = map[string]any{}
	}
	if preview.Logs == nil {
		preview.Logs = []string{}
	}
	return preview, nil
}

func (s *HookService) ListRuns(ctx context.Context, principal authz.Principal, hookID string, limit int, offset int) ([]automation.Run, error) {
	hook, err := s.Get(ctx, principal, hookID)
	if err != nil {
		return nil, err
	}
	if s.runs == nil {
		return []automation.Run{}, nil
	}
	return s.runs.List(ctx, automation.ListInput{
		TriggerType: "ticket_hook",
		TriggerRef:  hook.ID,
		ProjectID:   hook.ProjectID,
		Limit:       limit,
		Offset:      offset,
	})
}

func (s *HookService) RunBeforeCreate(ctx context.Context, principal authz.Principal, input CreateTicketInput) (CreateTicketInput, []HookResult, error) {
	results, transformed, err := s.run(ctx, principal, input.ProjectID, HookEventTicketCreate, HookPhaseBefore, map[string]any{
		"ticket": createTicketInputMap(input),
	})
	if err != nil {
		return input, results, err
	}
	ticket, ok := transformed["ticket"].(map[string]any)
	if !ok {
		return input, results, nil
	}
	return createTicketInputFromMap(ticket, input), results, nil
}

func (s *HookService) RunBeforeUpdate(ctx context.Context, principal authz.Principal, current Ticket, input UpdateTicketInput) (UpdateTicketInput, []HookResult, error) {
	results, transformed, err := s.run(ctx, principal, current.ProjectID, HookEventTicketUpdate, HookPhaseBefore, map[string]any{
		"current": ticketMap(current),
		"ticket":  updateTicketInputMap(input),
	})
	if err != nil {
		return input, results, err
	}
	ticket, ok := transformed["ticket"].(map[string]any)
	if !ok {
		return input, results, nil
	}
	return updateTicketInputFromMap(ticket, input), results, nil
}

func (s *HookService) RunAfterCreate(ctx context.Context, principal authz.Principal, ticket Ticket) []HookResult {
	results, _, _ := s.run(ctx, principal, ticket.ProjectID, HookEventTicketCreate, HookPhaseAfter, map[string]any{
		"ticket": ticketMap(ticket),
	})
	return results
}

func (s *HookService) RunAfterUpdate(ctx context.Context, principal authz.Principal, current Ticket, updated Ticket) []HookResult {
	results, _, _ := s.run(ctx, principal, updated.ProjectID, HookEventTicketUpdate, HookPhaseAfter, map[string]any{
		"current": ticketMap(current),
		"ticket":  ticketMap(updated),
	})
	return results
}

func (s *HookService) run(ctx context.Context, principal authz.Principal, projectID string, event string, phase string, input map[string]any) ([]HookResult, map[string]any, error) {
	if s == nil {
		return nil, input, nil
	}
	hooks, err := s.enabledHooks(ctx, projectID, event, phase)
	if err != nil {
		return nil, input, err
	}
	results := make([]HookResult, 0, len(hooks))
	current := copyHookMap(input)
	for _, hook := range hooks {
		result, err := s.executeRecorded(ctx, principal, hook, current)
		results = append(results, result)
		lastError := ""
		if err != nil {
			lastError = err.Error()
		}
		if recordErr := s.recordHookResult(ctx, hook.ID, lastError); recordErr != nil && err == nil {
			return results, current, recordErr
		}
		if err != nil {
			return results, current, err
		}
		if ticket, ok := result.Output["ticket"].(map[string]any); ok {
			current["ticket"] = ticket
		}
	}
	return results, current, nil
}

func (s *HookService) executeRecorded(ctx context.Context, principal authz.Principal, hook Hook, input map[string]any) (HookResult, error) {
	if s.runs == nil {
		return s.execute(ctx, principal, hook, input)
	}
	run, err := s.runs.Start(ctx, automation.StartInput{
		TriggerType: "ticket_hook",
		TriggerRef:  hook.ID,
		ProjectID:   hook.ProjectID,
		TicketID:    hookTicketID(input),
		Engine:      hook.Engine.Type,
		ActorUserID: principal.UserID,
		Input:       copyHookMap(input),
	})
	if err != nil {
		return HookResult{Output: map[string]any{}, Logs: nil}, err
	}
	result, execErr := s.execute(ctx, principal, hook, input)
	status := automation.StatusSucceeded
	errorMessage := ""
	if execErr != nil {
		status = automation.StatusFailed
		errorMessage = execErr.Error()
	}
	if result.Output == nil {
		result.Output = map[string]any{}
	}
	_, finishErr := s.runs.Finish(ctx, run.ID, automation.FinishInput{
		Status: status,
		Output: result.Output,
		Error:  errorMessage,
		Logs:   result.Logs,
	})
	if finishErr != nil {
		return result, finishErr
	}
	return result, execErr
}

func (s *HookService) execute(ctx context.Context, principal authz.Principal, hook Hook, input map[string]any) (HookResult, error) {
	switch hook.Engine.Type {
	case HookEngineLua:
		return s.executeLua(ctx, principal, hook, input)
	case HookEngineAI:
		return s.executeAI(ctx, principal, hook, input)
	default:
		return HookResult{Output: map[string]any{}, Logs: nil}, validationFailed(map[string]string{"engine": "Unsupported ticket hook engine"})
	}
}

func (s *HookService) executeLua(ctx context.Context, principal authz.Principal, hook Hook, input map[string]any) (HookResult, error) {
	runCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	sandbox := luasandbox.New(luasandbox.DefaultJSONLimits())
	defer sandbox.Close()
	sandbox.L.SetContext(runCtx)

	contextValue, err := sandbox.JSON.FromGo(map[string]any{
		"event":      hook.Event,
		"phase":      hook.Phase,
		"project_id": hook.ProjectID,
		"hook_id":    hook.ID,
		"user_id":    principal.UserID,
	})
	if err != nil {
		return HookResult{}, err
	}
	sandbox.L.SetGlobal("context", contextValue)
	if ticket, ok := input["ticket"]; ok {
		ticketValue, err := sandbox.JSON.FromGo(ticket)
		if err != nil {
			return HookResult{}, err
		}
		sandbox.L.SetGlobal("ticket", ticketValue)
	}
	if current, ok := input["current"]; ok {
		currentValue, err := sandbox.JSON.FromGo(current)
		if err != nil {
			return HookResult{}, err
		}
		sandbox.L.SetGlobal("current", currentValue)
	}
	logs := []string{}
	registerHookLuaHelpers(sandbox, &logs)

	fn, err := sandbox.L.LoadString(hook.Engine.Script)
	if err != nil {
		return HookResult{Output: map[string]any{}, Logs: logs}, err
	}
	top := sandbox.L.GetTop()
	sandbox.L.Push(fn)
	if err := sandbox.L.PCall(0, lua.MultRet, nil); err != nil {
		return HookResult{Output: map[string]any{}, Logs: logs}, err
	}
	output := map[string]any{}
	if sandbox.L.GetTop() > top {
		value, err := sandbox.JSON.ToGo(sandbox.L.Get(-1))
		if err != nil {
			return HookResult{Output: map[string]any{}, Logs: logs}, err
		}
		if object, ok := value.(map[string]any); ok {
			output = object
		}
	}
	if message := hookRejectMessage(output); message != "" {
		return HookResult{Output: output, Logs: logs}, validationFailed(map[string]string{"hook": message})
	}
	return HookResult{Output: output, Logs: logs}, nil
}

func (s *HookService) executeAI(ctx context.Context, principal authz.Principal, hook Hook, input map[string]any) (HookResult, error) {
	if s.openrouter == nil {
		return HookResult{Output: map[string]any{}}, validationFailed(map[string]string{"engine": "OpenRouter service is not configured"})
	}
	prompt, err := hookAIPrompt(principal, hook, input)
	if err != nil {
		return HookResult{Output: map[string]any{}}, err
	}
	result, err := s.openrouter.CompleteJSON(ctx, openrouter.CompletionInput{
		ProviderID: hook.Engine.ProviderID,
		Prompt:     prompt,
	})
	if err != nil {
		return HookResult{Output: map[string]any{}}, err
	}
	output := result.Output
	if output == nil {
		output = map[string]any{}
	}
	if message := hookRejectMessage(output); message != "" {
		return HookResult{Output: output, Logs: nil}, validationFailed(map[string]string{"hook": message})
	}
	return HookResult{Output: output, Logs: nil}, nil
}

func hookAIPrompt(principal authz.Principal, hook Hook, input map[string]any) (string, error) {
	payload := map[string]any{
		"context": map[string]any{
			"event":      hook.Event,
			"phase":      hook.Phase,
			"project_id": hook.ProjectID,
			"hook_id":    hook.ID,
			"user_id":    principal.UserID,
		},
		"input": input,
		"instructions": []string{
			"Return only a JSON object.",
			"For before hooks, return {\"ticket\": {...}} to transform the pending ticket or {\"reject\":{\"message\":\"...\"}} to reject.",
			"For after hooks, return an object for run output only; after hooks cannot change committed tickets.",
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode ticket hook AI input: %w", err)
	}
	return strings.TrimSpace(hook.Engine.Prompt) + "\n\nRayboard hook input:\n" + string(data), nil
}

func registerHookLuaHelpers(sandbox *luasandbox.Sandbox, logs *[]string) {
	rayboard := sandbox.L.GetGlobal("rayboard")
	rayboardTable, ok := rayboard.(*lua.LTable)
	if !ok {
		rayboardTable = sandbox.L.NewTable()
		sandbox.L.SetGlobal("rayboard", rayboardTable)
	}
	sandbox.L.SetField(rayboardTable, "log", sandbox.L.NewFunction(func(L *lua.LState) int {
		if len(*logs) < 100 {
			*logs = append(*logs, L.CheckString(1))
		}
		return 0
	}))
}

func (s *HookService) get(ctx context.Context, hookID string) (Hook, error) {
	hookID = strings.TrimSpace(hookID)
	if hookID == "" {
		return Hook{}, validationFailed(map[string]string{"hook_id": "Required"})
	}
	hook, err := scanHook(s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, event, phase, enabled, position, engine_type,
			COALESCE(lua_script, ''), COALESCE(ai_prompt, ''), COALESCE(ai_provider_id, ''),
			COALESCE(last_error, ''), created_at, updated_at
		FROM ticket_hooks
		WHERE id = ? AND deleted_at IS NULL
	`, hookID))
	if errors.Is(err, sql.ErrNoRows) {
		return Hook{}, notFound("ticket_hook", hookID)
	}
	if err != nil {
		return Hook{}, fmt.Errorf("get ticket hook: %w", err)
	}
	return hook, nil
}

func (s *HookService) enabledHooks(ctx context.Context, projectID string, event string, phase string) ([]Hook, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, name, event, phase, enabled, position, engine_type,
			COALESCE(lua_script, ''), COALESCE(ai_prompt, ''), COALESCE(ai_provider_id, ''),
			COALESCE(last_error, ''), created_at, updated_at
		FROM ticket_hooks
		WHERE project_id = ? AND event = ? AND phase = ? AND enabled = 1 AND deleted_at IS NULL
		ORDER BY position ASC, name ASC, id ASC
	`, projectID, event, phase)
	if err != nil {
		return nil, fmt.Errorf("list enabled ticket hooks: %w", err)
	}
	defer rows.Close()

	var hooks []Hook
	for rows.Next() {
		hook, err := scanHook(rows)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket hooks: %w", err)
	}
	return hooks, nil
}

func (s *HookService) recordHookResult(ctx context.Context, hookID string, lastError string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE ticket_hooks
		SET last_error = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, nullableHookString(lastError), formatHookTime(s.now().UTC()), hookID)
	if err != nil {
		return fmt.Errorf("record ticket hook result: %w", err)
	}
	return nil
}

func (s *HookService) validate(ctx context.Context, hook Hook) error {
	fields := map[string]string{}
	if hook.ProjectID == "" {
		fields["project_id"] = "Required"
	} else {
		var exists bool
		if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE id = ? AND deleted_at IS NULL)", hook.ProjectID).Scan(&exists); err != nil {
			return fmt.Errorf("check ticket hook project: %w", err)
		}
		if !exists {
			fields["project_id"] = "Project not found"
		}
	}
	if hook.Name == "" {
		fields["name"] = "Required"
	}
	if !validHookEvent(hook.Event) {
		fields["event"] = "Must be ticket_create or ticket_update"
	}
	if !validHookPhase(hook.Phase) {
		fields["phase"] = "Must be before or after"
	}
	if err := s.validateHookEngine(ctx, hook.Engine); err != nil {
		fields["engine"] = err.Error()
	}
	if len(fields) > 0 {
		return validationFailed(fields)
	}
	return nil
}

func (s *HookService) requireManage(principal authz.Principal, projectID string) error {
	if s == nil || s.authorizer == nil {
		return errors.New("tracker: hook authorization evaluator is required")
	}
	return s.authorizer.Require(principal, authz.PermissionAutomationsManage, authz.ProjectScope(projectID))
}

func scanHook(scanner interface{ Scan(...any) error }) (Hook, error) {
	var hook Hook
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&hook.ID,
		&hook.ProjectID,
		&hook.Name,
		&hook.Event,
		&hook.Phase,
		&hook.Enabled,
		&hook.Position,
		&hook.Engine.Type,
		&hook.Engine.Script,
		&hook.Engine.Prompt,
		&hook.Engine.ProviderID,
		&hook.LastError,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Hook{}, err
	}
	created, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Hook{}, fmt.Errorf("parse ticket hook created time: %w", err)
	}
	updated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return Hook{}, fmt.Errorf("parse ticket hook updated time: %w", err)
	}
	hook.CreatedAt = created
	hook.UpdatedAt = updated
	return hook, nil
}

func createTicketInputMap(input CreateTicketInput) map[string]any {
	return map[string]any{
		"project_id":       input.ProjectID,
		"title":            input.Title,
		"description":      input.Description,
		"status":           input.Status,
		"priority":         input.Priority,
		"type":             input.Type,
		"reporter_id":      input.ReporterID,
		"assignee_id":      input.AssigneeID,
		"parent_ticket_id": input.ParentTicketID,
		"sprint_id":        input.SprintID,
		"component_id":     input.ComponentID,
		"version_id":       input.VersionID,
		"rank":             input.Rank,
		"start_date":       input.StartDate,
		"due_date":         input.DueDate,
		"story_points":     floatValue(input.StoryPoints),
		"labels":           anyStringSlice(input.Labels),
		"custom_fields":    input.CustomFields,
	}
}

func updateTicketInputMap(input UpdateTicketInput) map[string]any {
	result := map[string]any{}
	setOptionalString(result, "title", input.Title)
	setOptionalString(result, "description", input.Description)
	setOptionalString(result, "status", input.Status)
	setOptionalString(result, "priority", input.Priority)
	setOptionalString(result, "type", input.Type)
	setOptionalString(result, "assignee_id", input.AssigneeID)
	setOptionalString(result, "parent_ticket_id", input.ParentTicketID)
	setOptionalString(result, "sprint_id", input.SprintID)
	setOptionalString(result, "component_id", input.ComponentID)
	setOptionalString(result, "version_id", input.VersionID)
	setOptionalString(result, "rank", input.Rank)
	setOptionalString(result, "start_date", input.StartDate)
	setOptionalString(result, "due_date", input.DueDate)
	if input.StoryPointsSet {
		result["story_points"] = floatValue(input.StoryPoints)
	}
	if input.Labels != nil {
		result["labels"] = anyStringSlice(*input.Labels)
	}
	if input.CustomFields != nil {
		result["custom_fields"] = *input.CustomFields
	}
	return result
}

func ticketMap(ticket Ticket) map[string]any {
	encoded, err := json.Marshal(ticket)
	if err != nil {
		return map[string]any{}
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return map[string]any{}
	}
	return decoded
}

func hookTicketID(input map[string]any) string {
	for _, key := range []string{"ticket", "current"} {
		object, ok := input[key].(map[string]any)
		if !ok {
			continue
		}
		id, _ := object["id"].(string)
		if strings.TrimSpace(id) != "" {
			return strings.TrimSpace(id)
		}
	}
	return ""
}

func createTicketInputFromMap(input map[string]any, fallback CreateTicketInput) CreateTicketInput {
	return CreateTicketInput{
		ProjectID:      stringFromMap(input, "project_id", fallback.ProjectID),
		Title:          stringFromMap(input, "title", fallback.Title),
		Description:    stringFromMap(input, "description", fallback.Description),
		Status:         stringFromMap(input, "status", fallback.Status),
		Priority:       stringFromMap(input, "priority", fallback.Priority),
		Type:           stringFromMap(input, "type", fallback.Type),
		ReporterID:     stringFromMap(input, "reporter_id", fallback.ReporterID),
		AssigneeID:     stringFromMap(input, "assignee_id", fallback.AssigneeID),
		ParentTicketID: stringFromMap(input, "parent_ticket_id", fallback.ParentTicketID),
		SprintID:       stringFromMap(input, "sprint_id", fallback.SprintID),
		ComponentID:    stringFromMap(input, "component_id", fallback.ComponentID),
		VersionID:      stringFromMap(input, "version_id", fallback.VersionID),
		Rank:           stringFromMap(input, "rank", fallback.Rank),
		StartDate:      stringFromMap(input, "start_date", fallback.StartDate),
		DueDate:        stringFromMap(input, "due_date", fallback.DueDate),
		StoryPoints:    floatFromMap(input, "story_points", fallback.StoryPoints),
		Labels:         stringSliceFromMap(input, "labels", fallback.Labels),
		CustomFields:   objectFromMap(input, "custom_fields", fallback.CustomFields),
	}
}

func updateTicketInputFromMap(input map[string]any, fallback UpdateTicketInput) UpdateTicketInput {
	result := fallback
	updateOptionalString(input, "title", &result.Title)
	updateOptionalString(input, "description", &result.Description)
	updateOptionalString(input, "status", &result.Status)
	updateOptionalString(input, "priority", &result.Priority)
	updateOptionalString(input, "type", &result.Type)
	updateOptionalString(input, "assignee_id", &result.AssigneeID)
	updateOptionalString(input, "parent_ticket_id", &result.ParentTicketID)
	updateOptionalString(input, "sprint_id", &result.SprintID)
	updateOptionalString(input, "component_id", &result.ComponentID)
	updateOptionalString(input, "version_id", &result.VersionID)
	updateOptionalString(input, "rank", &result.Rank)
	updateOptionalString(input, "start_date", &result.StartDate)
	updateOptionalString(input, "due_date", &result.DueDate)
	if _, ok := input["story_points"]; ok {
		result.StoryPoints = floatFromMap(input, "story_points", nil)
		result.StoryPointsSet = true
	}
	if _, ok := input["labels"]; ok {
		labels := stringSliceFromMap(input, "labels", nil)
		result.Labels = &labels
	}
	if _, ok := input["custom_fields"]; ok {
		customFields := objectFromMap(input, "custom_fields", nil)
		result.CustomFields = &customFields
	}
	return result
}

func copyHookMap(input map[string]any) map[string]any {
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func setOptionalString(result map[string]any, key string, value *string) {
	if value == nil {
		return
	}
	result[key] = *value
}

func updateOptionalString(input map[string]any, key string, target **string) {
	if _, ok := input[key]; !ok {
		return
	}
	value := stringFromMap(input, key, "")
	*target = &value
}

func stringFromMap(input map[string]any, key string, fallback string) string {
	value, ok := input[key]
	if !ok || value == nil {
		return fallback
	}
	text, ok := value.(string)
	if !ok {
		return fallback
	}
	return text
}

func stringSliceFromMap(input map[string]any, key string, fallback []string) []string {
	value, ok := input[key]
	if !ok || value == nil {
		return fallback
	}
	items, ok := value.([]any)
	if !ok {
		return fallback
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return fallback
		}
		result = append(result, text)
	}
	return result
}

func floatFromMap(input map[string]any, key string, fallback *float64) *float64 {
	value, ok := input[key]
	if !ok || value == nil {
		return fallback
	}
	switch typed := value.(type) {
	case float64:
		return &typed
	case int:
		result := float64(typed)
		return &result
	case int64:
		result := float64(typed)
		return &result
	default:
		return fallback
	}
}

func floatValue(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func objectFromMap(input map[string]any, key string, fallback map[string]any) map[string]any {
	value, ok := input[key]
	if !ok || value == nil {
		return fallback
	}
	object, ok := value.(map[string]any)
	if !ok {
		return fallback
	}
	return object
}

func anyStringSlice(values []string) []any {
	if values == nil {
		return nil
	}
	result := make([]any, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func hookRejectMessage(output map[string]any) string {
	value, ok := output["reject"]
	if !ok || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	reject, ok := value.(map[string]any)
	if !ok {
		return "rejected"
	}
	if message, ok := reject["message"].(string); ok && strings.TrimSpace(message) != "" {
		return strings.TrimSpace(message)
	}
	return "rejected"
}

func normalizeHookEngine(engine HookEngineSpec) HookEngineSpec {
	engine.Type = strings.ToLower(strings.TrimSpace(engine.Type))
	engine.Script = strings.TrimSpace(engine.Script)
	engine.Prompt = strings.TrimSpace(engine.Prompt)
	engine.ProviderID = strings.TrimSpace(engine.ProviderID)
	return engine
}

func (s *HookService) validateHookEngine(ctx context.Context, engine HookEngineSpec) error {
	switch engine.Type {
	case HookEngineLua:
		if strings.TrimSpace(engine.Script) == "" {
			return errors.New("lua engine requires script")
		}
	case HookEngineAI:
		if strings.TrimSpace(engine.Prompt) == "" {
			return errors.New("ai engine requires prompt")
		}
		if strings.TrimSpace(engine.ProviderID) == "" {
			return errors.New("ai engine requires provider_id")
		}
		if err := s.validateHookAIProvider(ctx, engine.ProviderID); err != nil {
			return err
		}
	default:
		return errors.New("engine type must be lua or ai")
	}
	return nil
}

func (s *HookService) validateHookAIProvider(ctx context.Context, providerID string) error {
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

func validHookEvent(event string) bool {
	return event == HookEventTicketCreate || event == HookEventTicketUpdate
}

func validHookPhase(phase string) bool {
	return phase == HookPhaseBefore || phase == HookPhaseAfter
}

func normalizeHookName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func nullableHookString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func formatHookTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func newHookID(prefix string) string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return prefix + "_fallback"
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(raw[:])
}

func isHookUniqueConstraint(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "unique constraint") || strings.Contains(text, "constraint failed")
}
