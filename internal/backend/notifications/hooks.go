package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/luasandbox"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	lua "github.com/yuin/gopher-lua"
)

const (
	HookEngineLua = "lua"
	HookEngineAI  = "ai"
)

type HookEngine struct {
	Type       string
	Script     string
	Prompt     string
	ProviderID string
}

type Hook struct {
	ID          string
	Name        string
	ScopeType   string
	ProjectID   string
	ActorUserID string
	EventTypes  []string
	Enabled     bool
	Engine      HookEngine
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListHooksInput struct {
	ScopeType string
	ProjectID string
}

type CreateHookInput struct {
	Name        string
	ScopeType   string
	ProjectID   string
	ActorUserID string
	EventTypes  []string
	Enabled     bool
	Engine      HookEngine
}

type UpdateHookInput struct {
	Name        *string
	ActorUserID *string
	EventTypes  *[]string
	Enabled     *bool
	Engine      *HookEngine
}

type PreviewHookInput struct {
	PolicyID       string
	EventType      string
	ProjectID      string
	SubjectType    string
	SubjectID      string
	Message        string
	Payload        map[string]any
	DestinationIDs []string
}

type PreviewHookResult struct {
	Run  automation.Run
	Plan HookPreviewPlan
}

type HookPreviewPlan struct {
	EventType      string
	ProjectID      string
	SubjectType    string
	SubjectID      string
	Message        string
	Payload        map[string]any
	DestinationIDs []string
	Suppressed     bool
}

func (plan HookPreviewPlan) Map() map[string]any {
	return map[string]any{
		"event_type":      plan.EventType,
		"project_id":      plan.ProjectID,
		"subject_type":    plan.SubjectType,
		"subject_id":      plan.SubjectID,
		"message":         plan.Message,
		"payload":         nonNilMap(plan.Payload),
		"destination_ids": stringAnySlice(plan.DestinationIDs),
		"suppressed":      plan.Suppressed,
	}
}

type hookPlan struct {
	EventType      string
	ProjectID      string
	SubjectType    string
	SubjectID      string
	Message        string
	Payload        map[string]any
	DestinationIDs []string
	Suppressed     bool
}

func (s *Service) ListHooks(ctx context.Context, input ListHooksInput) ([]Hook, error) {
	scopeType := strings.TrimSpace(input.ScopeType)
	if scopeType == "" {
		scopeType = PolicyScopeGlobal
	}
	projectID := strings.TrimSpace(input.ProjectID)
	if err := validatePolicyScope(scopeType, projectID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, scope_type, project_id, actor_user_id, event_types_json, enabled,
			engine_type, COALESCE(lua_script, ''), COALESCE(ai_prompt, ''), COALESCE(ai_provider_id, ''),
			COALESCE(last_error, ''), created_at, updated_at
		FROM notification_hooks
		WHERE deleted_at IS NULL AND scope_type = ? AND scope_key = ?
		ORDER BY name ASC
	`, scopeType, policyScopeKey(scopeType, projectID))
	if err != nil {
		return nil, fmt.Errorf("list notification hooks: %w", err)
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
		return nil, fmt.Errorf("iterate notification hooks: %w", err)
	}
	return hooks, nil
}

func (s *Service) CreateHook(ctx context.Context, input CreateHookInput) (Hook, error) {
	id, err := newID("notif_hook")
	if err != nil {
		return Hook{}, err
	}
	hook := Hook{
		ID:          id,
		Name:        normalizePolicyName(input.Name),
		ScopeType:   strings.TrimSpace(input.ScopeType),
		ProjectID:   strings.TrimSpace(input.ProjectID),
		ActorUserID: strings.TrimSpace(input.ActorUserID),
		EventTypes:  normalizePolicyEventTypes(input.EventTypes),
		Enabled:     input.Enabled,
		Engine:      normalizeHookEngine(input.Engine),
		CreatedAt:   s.now().UTC(),
		UpdatedAt:   s.now().UTC(),
	}
	if hook.ScopeType == "" {
		hook.ScopeType = PolicyScopeGlobal
	}
	if err := s.validateHook(ctx, hook); err != nil {
		return Hook{}, err
	}
	eventTypesJSON, err := marshalStringList(hook.EventTypes)
	if err != nil {
		return Hook{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO notification_hooks (
			id, name, scope_type, scope_key, project_id, actor_user_id, event_types_json,
			enabled, engine_type, lua_script, ai_prompt, ai_provider_id, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, hook.ID, hook.Name, hook.ScopeType, policyScopeKey(hook.ScopeType, hook.ProjectID), nullableString(hook.ProjectID),
		hook.ActorUserID, eventTypesJSON, hook.Enabled, hook.Engine.Type, nullableString(hook.Engine.Script),
		nullableString(hook.Engine.Prompt), nullableString(hook.Engine.ProviderID), formatTime(hook.CreatedAt), formatTime(hook.UpdatedAt)); err != nil {
		if isUniqueConstraint(err) {
			return Hook{}, fmt.Errorf("%w: hook name already exists in scope", ErrValidation)
		}
		return Hook{}, fmt.Errorf("insert notification hook: %w", err)
	}
	return hook, nil
}

func (s *Service) GetHook(ctx context.Context, hookID string) (Hook, error) {
	hook, err := scanHook(s.db.QueryRowContext(ctx, `
		SELECT id, name, scope_type, project_id, actor_user_id, event_types_json, enabled,
			engine_type, COALESCE(lua_script, ''), COALESCE(ai_prompt, ''), COALESCE(ai_provider_id, ''),
			COALESCE(last_error, ''), created_at, updated_at
		FROM notification_hooks
		WHERE id = ? AND deleted_at IS NULL
	`, hookID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Hook{}, ErrNotFound
		}
		return Hook{}, fmt.Errorf("get notification hook: %w", err)
	}
	return hook, nil
}

func (s *Service) UpdateHook(ctx context.Context, hookID string, input UpdateHookInput) (Hook, error) {
	hook, err := s.GetHook(ctx, hookID)
	if err != nil {
		return Hook{}, err
	}
	if input.Name != nil {
		hook.Name = normalizePolicyName(*input.Name)
	}
	if input.ActorUserID != nil {
		hook.ActorUserID = strings.TrimSpace(*input.ActorUserID)
	}
	if input.EventTypes != nil {
		hook.EventTypes = normalizePolicyEventTypes(*input.EventTypes)
	}
	if input.Enabled != nil {
		hook.Enabled = *input.Enabled
	}
	if input.Engine != nil {
		hook.Engine = normalizeHookEngine(*input.Engine)
	}
	hook.UpdatedAt = s.now().UTC()
	if err := s.validateHook(ctx, hook); err != nil {
		return Hook{}, err
	}
	eventTypesJSON, err := marshalStringList(hook.EventTypes)
	if err != nil {
		return Hook{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE notification_hooks
		SET name = ?, actor_user_id = ?, event_types_json = ?, enabled = ?, engine_type = ?,
			lua_script = ?, ai_prompt = ?, ai_provider_id = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, hook.Name, hook.ActorUserID, eventTypesJSON, hook.Enabled, hook.Engine.Type,
		nullableString(hook.Engine.Script), nullableString(hook.Engine.Prompt), nullableString(hook.Engine.ProviderID),
		formatTime(hook.UpdatedAt), hookID)
	if err != nil {
		if isUniqueConstraint(err) {
			return Hook{}, fmt.Errorf("%w: hook name already exists in scope", ErrValidation)
		}
		return Hook{}, fmt.Errorf("update notification hook: %w", err)
	}
	if err := requireRowsAffected(result, "notification hook update"); err != nil {
		return Hook{}, err
	}
	return s.GetHook(ctx, hookID)
}

func (s *Service) DeleteHook(ctx context.Context, hookID string) error {
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE notification_hooks
		SET deleted_at = ?, enabled = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatTime(now), formatTime(now), hookID)
	if err != nil {
		return fmt.Errorf("delete notification hook: %w", err)
	}
	return requireRowsAffected(result, "notification hook delete")
}

func (s *Service) PreviewHook(ctx context.Context, hookID string, input PreviewHookInput) (PreviewHookResult, error) {
	if s.runs == nil {
		return PreviewHookResult{}, fmt.Errorf("%w: notification hook run store is not configured", ErrValidation)
	}
	hook, err := s.GetHook(ctx, hookID)
	if err != nil {
		return PreviewHookResult{}, err
	}
	policy, err := s.previewPolicy(ctx, hook, input)
	if err != nil {
		return PreviewHookResult{}, err
	}
	plan := hookPlan{
		EventType:      previewEventType(input.EventType, hook.EventTypes),
		ProjectID:      strings.TrimSpace(input.ProjectID),
		SubjectType:    strings.TrimSpace(input.SubjectType),
		SubjectID:      strings.TrimSpace(input.SubjectID),
		Message:        strings.TrimSpace(input.Message),
		Payload:        nonNilMap(input.Payload),
		DestinationIDs: normalizeHookPreviewStrings(input.DestinationIDs),
	}
	if plan.ProjectID == "" {
		plan.ProjectID = hook.ProjectID
	}
	if plan.Message == "" {
		plan.Message = "Notification hook preview"
	}
	if len(plan.DestinationIDs) == 0 {
		plan.DestinationIDs = append([]string(nil), policy.DestinationIDs...)
	}
	if plan.EventType == "" {
		return PreviewHookResult{}, fmt.Errorf("%w: event type is required", ErrValidation)
	}
	if !slices.Contains(hook.EventTypes, plan.EventType) {
		return PreviewHookResult{}, fmt.Errorf("%w: event type is not configured for notification hook", ErrValidation)
	}
	run, err := s.runs.Start(ctx, automation.StartInput{
		TriggerType: "notification_hook_preview",
		TriggerRef:  hook.ID,
		ProjectID:   plan.ProjectID,
		Engine:      hook.Engine.Type,
		ActorUserID: hook.ActorUserID,
		Input:       notificationHookRunInput(policy, plan, true),
		Limits:      map[string]any{"timeout_seconds": 10, "dry_run": true},
	})
	if err != nil {
		return PreviewHookResult{}, err
	}
	output, execErr := s.executeNotificationHook(ctx, hook, policy, plan)
	resultPlan := applyHookOutput(plan, policy.DestinationIDs, output)
	finish := automation.FinishInput{
		Status: automation.StatusSucceeded,
		Output: map[string]any{
			"hook_output": output,
			"plan":        previewPlanFromInternal(resultPlan).Map(),
		},
	}
	if execErr != nil {
		finish.Status = automation.StatusFailed
		finish.Error = execErr.Error()
	}
	finished, finishErr := s.runs.Finish(ctx, run.ID, finish)
	if finishErr != nil {
		return PreviewHookResult{}, finishErr
	}
	result := PreviewHookResult{Run: finished, Plan: previewPlanFromInternal(resultPlan)}
	if execErr != nil {
		return result, execErr
	}
	return result, nil
}

func (s *Service) ListHookRuns(ctx context.Context, hookID string, limit int, offset int) ([]automation.Run, error) {
	if s.runs == nil {
		return nil, fmt.Errorf("%w: notification hook run store is not configured", ErrValidation)
	}
	hook, err := s.GetHook(ctx, hookID)
	if err != nil {
		return nil, err
	}
	return s.runs.List(ctx, automation.ListInput{TriggerRef: hook.ID, Limit: limit, Offset: offset})
}

func (s *Service) applyNotificationHooks(ctx context.Context, policy Policy, plan externalNotificationPlan) (hookPlan, error) {
	result := hookPlan{
		EventType:      plan.EventType,
		ProjectID:      plan.ProjectID,
		SubjectType:    plan.SubjectType,
		SubjectID:      plan.SubjectID,
		Message:        plan.Message,
		Payload:        nonNilMap(plan.Payload),
		DestinationIDs: append([]string(nil), policy.DestinationIDs...),
	}
	hooks, err := s.matchingHooks(ctx, plan.EventType, plan.ProjectID)
	if err != nil {
		return result, err
	}
	for _, hook := range hooks {
		output, err := s.executeNotificationHookRecorded(ctx, hook, policy, result, "notification_hook")
		if err != nil {
			_ = s.recordHookResult(ctx, hook.ID, err.Error())
			return result, err
		}
		result = applyHookOutput(result, policy.DestinationIDs, output)
		if result.Suppressed {
			break
		}
		_ = s.recordHookResult(ctx, hook.ID, "")
	}
	return result, nil
}

func (s *Service) executeNotificationHookRecorded(ctx context.Context, hook Hook, policy Policy, plan hookPlan, triggerType string) (map[string]any, error) {
	if s.runs == nil {
		return s.executeNotificationHook(ctx, hook, policy, plan)
	}
	run, err := s.runs.Start(ctx, automation.StartInput{
		TriggerType: triggerType,
		TriggerRef:  hook.ID,
		ProjectID:   plan.ProjectID,
		Engine:      hook.Engine.Type,
		ActorUserID: hook.ActorUserID,
		Input:       notificationHookRunInput(policy, plan, false),
		Limits: map[string]any{
			"timeout_seconds": 10,
		},
	})
	if err != nil {
		return nil, err
	}
	output, execErr := s.executeNotificationHook(ctx, hook, policy, plan)
	finish := automation.FinishInput{
		Status: automation.StatusSucceeded,
		Output: map[string]any{
			"hook_output": output,
		},
	}
	if execErr != nil {
		finish.Status = automation.StatusFailed
		finish.Error = execErr.Error()
	}
	if _, finishErr := s.runs.Finish(ctx, run.ID, finish); finishErr != nil {
		return nil, finishErr
	}
	return output, execErr
}

func (s *Service) previewPolicy(ctx context.Context, hook Hook, input PreviewHookInput) (Policy, error) {
	if strings.TrimSpace(input.PolicyID) != "" {
		policy, err := s.GetPolicy(ctx, input.PolicyID)
		if err != nil {
			return Policy{}, err
		}
		return policy, nil
	}
	return Policy{
		ID:             "preview",
		Name:           "preview",
		ScopeType:      hook.ScopeType,
		ProjectID:      hook.ProjectID,
		EventTypes:     append([]string(nil), hook.EventTypes...),
		DestinationIDs: normalizeHookPreviewStrings(input.DestinationIDs),
		Enabled:        true,
	}, nil
}

func previewEventType(value string, allowed []string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	if len(allowed) > 0 {
		return allowed[0]
	}
	return ""
}

func notificationHookRunInput(policy Policy, plan hookPlan, dryRun bool) map[string]any {
	return map[string]any{
		"policy_id":       policy.ID,
		"policy_name":     policy.Name,
		"event_type":      plan.EventType,
		"project_id":      plan.ProjectID,
		"subject_type":    plan.SubjectType,
		"subject_id":      plan.SubjectID,
		"message":         plan.Message,
		"payload":         nonNilMap(plan.Payload),
		"destination_ids": stringAnySlice(plan.DestinationIDs),
		"dry_run":         dryRun,
	}
}

func previewPlanFromInternal(plan hookPlan) HookPreviewPlan {
	return HookPreviewPlan{
		EventType:      plan.EventType,
		ProjectID:      plan.ProjectID,
		SubjectType:    plan.SubjectType,
		SubjectID:      plan.SubjectID,
		Message:        plan.Message,
		Payload:        nonNilMap(plan.Payload),
		DestinationIDs: append([]string(nil), plan.DestinationIDs...),
		Suppressed:     plan.Suppressed,
	}
}

func normalizeHookPreviewStrings(values []string) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func (s *Service) matchingHooks(ctx context.Context, eventType string, projectID string) ([]Hook, error) {
	var matched []Hook
	globalHooks, err := s.ListHooks(ctx, ListHooksInput{ScopeType: PolicyScopeGlobal})
	if err != nil {
		return nil, err
	}
	matched = appendMatchingHooks(matched, globalHooks, eventType)
	if projectID != "" {
		projectHooks, err := s.ListHooks(ctx, ListHooksInput{ScopeType: PolicyScopeProject, ProjectID: projectID})
		if err != nil {
			return nil, err
		}
		matched = appendMatchingHooks(matched, projectHooks, eventType)
	}
	return matched, nil
}

func appendMatchingHooks(result []Hook, hooks []Hook, eventType string) []Hook {
	for _, hook := range hooks {
		if hook.Enabled && slices.Contains(hook.EventTypes, eventType) {
			result = append(result, hook)
		}
	}
	return result
}

func (s *Service) executeNotificationHook(ctx context.Context, hook Hook, policy Policy, plan hookPlan) (map[string]any, error) {
	if err := s.requireActiveUser(ctx, hook.ActorUserID); err != nil {
		return nil, err
	}
	switch hook.Engine.Type {
	case HookEngineLua:
		return s.executeNotificationHookLua(ctx, hook, policy, plan)
	case HookEngineAI:
		return s.executeNotificationHookAI(ctx, hook, policy, plan)
	default:
		return nil, fmt.Errorf("%w: unsupported notification hook engine", ErrValidation)
	}
}

func (s *Service) executeNotificationHookLua(ctx context.Context, hook Hook, policy Policy, plan hookPlan) (map[string]any, error) {
	runCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	sandbox := luasandbox.New(luasandbox.DefaultJSONLimits())
	defer sandbox.Close()
	sandbox.L.SetContext(runCtx)
	input, err := notificationHookInput(hook, policy, plan)
	if err != nil {
		return nil, err
	}
	value, err := sandbox.JSON.FromGo(input)
	if err != nil {
		return nil, err
	}
	sandbox.L.SetGlobal("notification", value)
	fn, err := sandbox.L.LoadString(hook.Engine.Script)
	if err != nil {
		return nil, err
	}
	top := sandbox.L.GetTop()
	sandbox.L.Push(fn)
	if err := sandbox.L.PCall(0, lua.MultRet, nil); err != nil {
		return nil, err
	}
	if sandbox.L.GetTop() <= top {
		return map[string]any{}, nil
	}
	returned, err := sandbox.JSON.ToGo(sandbox.L.Get(-1))
	if err != nil {
		return nil, err
	}
	object, ok := returned.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: notification hook must return an object", ErrValidation)
	}
	return object, nil
}

func (s *Service) executeNotificationHookAI(ctx context.Context, hook Hook, policy Policy, plan hookPlan) (map[string]any, error) {
	if s.openrouter == nil {
		return nil, fmt.Errorf("%w: OpenRouter service is not configured", ErrValidation)
	}
	input, err := notificationHookInput(hook, policy, plan)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("encode notification hook AI input: %w", err)
	}
	result, err := s.openrouter.CompleteJSON(ctx, openrouter.CompletionInput{
		ProviderID: hook.Engine.ProviderID,
		Prompt:     strings.TrimSpace(hook.Engine.Prompt) + "\n\nRayboard notification hook input:\n" + string(data),
	})
	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

func notificationHookInput(hook Hook, policy Policy, plan hookPlan) (map[string]any, error) {
	return map[string]any{
		"context": map[string]any{
			"hook_id":    hook.ID,
			"scope_type": hook.ScopeType,
			"project_id": hook.ProjectID,
			"user_id":    hook.ActorUserID,
		},
		"policy": map[string]any{
			"id":              policy.ID,
			"name":            policy.Name,
			"scope_type":      policy.ScopeType,
			"project_id":      policy.ProjectID,
			"destination_ids": stringAnySlice(policy.DestinationIDs),
		},
		"plan": map[string]any{
			"event_type":      plan.EventType,
			"project_id":      plan.ProjectID,
			"subject_type":    plan.SubjectType,
			"subject_id":      plan.SubjectID,
			"message":         plan.Message,
			"payload":         nonNilMap(plan.Payload),
			"destination_ids": stringAnySlice(plan.DestinationIDs),
		},
		"instructions": []any{
			"Return only a JSON object.",
			"Allowed output fields are suppress, message, payload, and destination_ids.",
			"destination_ids may only contain ids already present in policy.destination_ids.",
		},
	}, nil
}

func applyHookOutput(plan hookPlan, allowedDestinations []string, output map[string]any) hookPlan {
	if value, ok := output["suppress"].(bool); ok && value {
		plan.Suppressed = true
		return plan
	}
	if message, ok := output["message"].(string); ok && strings.TrimSpace(message) != "" {
		plan.Message = strings.TrimSpace(message)
	}
	if payload, ok := output["payload"].(map[string]any); ok {
		plan.Payload = nonNilMap(payload)
	}
	if destinations, ok := stringSliceFromAny(output["destination_ids"]); ok {
		plan.DestinationIDs = filterAllowedDestinations(destinations, allowedDestinations)
		if len(plan.DestinationIDs) == 0 {
			plan.Suppressed = true
		}
	}
	return plan
}

func filterAllowedDestinations(values []string, allowed []string) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] || !slices.Contains(allowed, value) {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func stringSliceFromAny(value any) ([]string, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return nil, false
		}
		result = append(result, text)
	}
	return result, true
}

func stringAnySlice(values []string) []any {
	result := make([]any, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func (s *Service) validateHook(ctx context.Context, hook Hook) error {
	fields := map[string]string{}
	if err := validatePolicyScope(hook.ScopeType, hook.ProjectID); err != nil {
		return err
	}
	if hook.ScopeType == PolicyScopeProject {
		if err := s.requireProject(ctx, hook.ProjectID); err != nil {
			return err
		}
	}
	if hook.Name == "" {
		fields["name"] = "Required"
	}
	if hook.ActorUserID == "" {
		fields["actor_user_id"] = "Required"
	} else if err := s.requireActiveUser(ctx, hook.ActorUserID); err != nil {
		return err
	}
	if len(hook.EventTypes) == 0 {
		fields["event_types"] = "At least one event type is required"
	}
	for _, eventType := range hook.EventTypes {
		if !slices.Contains(allowedPolicyEvents, eventType) {
			fields["event_types"] = "Contains an unsupported event type"
			break
		}
	}
	if err := s.validateHookEngine(ctx, hook.Engine); err != nil {
		fields["engine"] = err.Error()
	}
	if len(fields) > 0 {
		return fmt.Errorf("%w: invalid notification hook", ErrValidation)
	}
	return nil
}

func (s *Service) validateHookEngine(ctx context.Context, engine HookEngine) error {
	switch engine.Type {
	case HookEngineLua:
		if strings.TrimSpace(engine.Script) == "" {
			return errors.New("lua engine requires script")
		}
	case HookEngineAI:
		if strings.TrimSpace(engine.Prompt) == "" || strings.TrimSpace(engine.ProviderID) == "" {
			return errors.New("ai engine requires prompt and provider_id")
		}
		if err := s.validateAIProvider(ctx, engine.ProviderID); err != nil {
			return err
		}
	default:
		return errors.New("engine type must be lua or ai")
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
	return nil
}

func (s *Service) requireActiveUser(ctx context.Context, userID string) error {
	var exists bool
	if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL AND is_disabled = 0)", userID).Scan(&exists); err != nil {
		return fmt.Errorf("check notification hook actor user: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: notification hook actor user is disabled or deleted", ErrValidation)
	}
	return nil
}

func (s *Service) recordHookResult(ctx context.Context, hookID string, lastError string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE notification_hooks
		SET last_error = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, nullableString(lastError), formatTime(s.now().UTC()), hookID)
	return err
}

func scanHook(scanner interface{ Scan(...any) error }) (Hook, error) {
	var hook Hook
	var projectID sql.NullString
	var eventTypesJSON string
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&hook.ID,
		&hook.Name,
		&hook.ScopeType,
		&projectID,
		&hook.ActorUserID,
		&eventTypesJSON,
		&hook.Enabled,
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
	eventTypes, err := unmarshalStringList(eventTypesJSON)
	if err != nil {
		return Hook{}, err
	}
	created, err := parseTime(createdAt)
	if err != nil {
		return Hook{}, err
	}
	updated, err := parseTime(updatedAt)
	if err != nil {
		return Hook{}, err
	}
	hook.ProjectID = nullString(projectID)
	hook.EventTypes = eventTypes
	hook.CreatedAt = created
	hook.UpdatedAt = updated
	return hook, nil
}

func normalizeHookEngine(engine HookEngine) HookEngine {
	engine.Type = strings.ToLower(strings.TrimSpace(engine.Type))
	engine.Script = strings.TrimSpace(engine.Script)
	engine.Prompt = strings.TrimSpace(engine.Prompt)
	engine.ProviderID = strings.TrimSpace(engine.ProviderID)
	return engine
}
