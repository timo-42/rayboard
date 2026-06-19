package webhooks

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/luasandbox"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/tracker"
	lua "github.com/yuin/gopher-lua"
)

const (
	DirectionIncoming = "incoming"
	DirectionOutgoing = "outgoing"

	EngineTypeLua = "lua"
	EngineTypeAI  = "ai"
)

var (
	ErrNotFound   = errors.New("webhooks: not found")
	ErrValidation = errors.New("webhooks: validation failed")
	ErrDelivery   = errors.New("webhooks: delivery failed")
)

type Service struct {
	db              *sql.DB
	authorizer      authz.Evaluator
	runs            *automation.RunStore
	tracker         *tracker.Service
	search          *search.Service
	comments        *comments.Service
	openrouter      *openrouter.Service
	httpClient      httpClient
	outgoingBaseURL string
	outgoingBases   OutgoingBaseURLProvider
	now             func() time.Time
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type OutgoingBaseURLProvider interface {
	OutgoingWebhookBaseURLs(context.Context) ([]string, error)
}

type Option func(*Service)

func NewService(db *sql.DB, authorizer authz.Evaluator, options ...Option) *Service {
	service := &Service{
		db:         db,
		authorizer: authorizer,
		httpClient: &http.Client{Timeout: 10 * time.Second},
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

func WithRunStore(runStore *automation.RunStore) Option {
	return func(service *Service) {
		service.runs = runStore
	}
}

func WithTrackerService(trackerService *tracker.Service) Option {
	return func(service *Service) {
		service.tracker = trackerService
	}
}

func WithSearchService(searchService *search.Service) Option {
	return func(service *Service) {
		service.search = searchService
	}
}

func WithCommentService(commentService *comments.Service) Option {
	return func(service *Service) {
		service.comments = commentService
	}
}

func WithOpenRouterService(openRouterService *openrouter.Service) Option {
	return func(service *Service) {
		service.openrouter = openRouterService
	}
}

func WithHTTPClient(client httpClient) Option {
	return func(service *Service) {
		if client != nil {
			service.httpClient = client
		}
	}
}

func WithOutgoingBaseURL(baseURL string) Option {
	return func(service *Service) {
		service.outgoingBaseURL = strings.TrimSpace(baseURL)
	}
}

func WithOutgoingBaseURLProvider(provider OutgoingBaseURLProvider) Option {
	return func(service *Service) {
		service.outgoingBases = provider
	}
}

type EngineSpec struct {
	Type       string
	Script     string
	Prompt     string
	ProviderID string
}

type Webhook struct {
	ID             string
	ProjectID      string
	Name           string
	Direction      string
	Enabled        bool
	ActorUserID    string
	EventTypes     []string
	Engine         EngineSpec
	TokenSet       bool
	TokenRotatedAt *time.Time
	LastError      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreatedWebhook struct {
	Webhook
	Token string
}

type ListInput struct {
	ProjectID string
	Direction string
	Limit     int
	Offset    int
}

type CreateInput struct {
	ProjectID   string
	Name        string
	Direction   string
	Enabled     bool
	ActorUserID string
	EventTypes  []string
	Engine      EngineSpec
}

type UpdateInput struct {
	Name        *string
	Direction   *string
	Enabled     *bool
	ActorUserID *string
	EventTypes  *[]string
	Engine      *EngineSpec
}

type IncomingInput struct {
	Headers map[string]string
	Query   map[string]string
	Payload map[string]any
}

type IncomingResult struct {
	Webhook Webhook
	Run     automation.Run
}

func (s *Service) List(ctx context.Context, principal authz.Principal, input ListInput) ([]Webhook, error) {
	projectID := strings.TrimSpace(input.ProjectID)
	if projectID == "" {
		return nil, fmt.Errorf("%w: project_id is required", ErrValidation)
	}
	if err := s.require(principal, projectID); err != nil {
		return nil, err
	}
	limit, offset, err := normalizeList(input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}
	where := []string{"project_id = ?", "deleted_at IS NULL"}
	args := []any{projectID}
	if direction := strings.TrimSpace(input.Direction); direction != "" {
		if !validDirection(direction) {
			return nil, fmt.Errorf("%w: invalid webhook direction", ErrValidation)
		}
		where = append(where, "direction = ?")
		args = append(args, direction)
	}
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, name, direction, enabled, actor_user_id, engine_type,
			COALESCE(lua_script, ''), COALESCE(ai_prompt, ''), COALESCE(ai_provider_id, ''),
			event_types_json,
			token_hash, token_rotated_at, COALESCE(last_error, ''), created_at, updated_at
		FROM webhooks
		WHERE `+joinAnd(where)+`
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}
	defer rows.Close()

	var hooks []Webhook
	for rows.Next() {
		hook, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate webhooks: %w", err)
	}
	return hooks, nil
}

func (s *Service) Create(ctx context.Context, principal authz.Principal, input CreateInput) (CreatedWebhook, error) {
	hook := Webhook{
		ID:          newID("webhook"),
		ProjectID:   strings.TrimSpace(input.ProjectID),
		Name:        normalizeName(input.Name),
		Direction:   strings.TrimSpace(input.Direction),
		Enabled:     input.Enabled,
		ActorUserID: strings.TrimSpace(input.ActorUserID),
		EventTypes:  normalizeEventTypes(input.EventTypes),
		Engine:      normalizeEngine(input.Engine),
		CreatedAt:   s.now().UTC(),
		UpdatedAt:   s.now().UTC(),
	}
	if hook.Direction == "" {
		hook.Direction = DirectionIncoming
	}
	if err := s.require(principal, hook.ProjectID); err != nil {
		return CreatedWebhook{}, err
	}
	if err := s.validateWebhook(ctx, hook, true); err != nil {
		return CreatedWebhook{}, err
	}
	token := ""
	tokenHash := any(nil)
	tokenRotatedAt := any(nil)
	if hook.Direction == DirectionIncoming {
		secret, err := randomSecret()
		if err != nil {
			return CreatedWebhook{}, err
		}
		token = "wh_" + secret
		tokenHash = hashSecret(token)
		tokenRotatedAt = formatTime(hook.CreatedAt)
		hook.TokenSet = true
		hook.TokenRotatedAt = &hook.CreatedAt
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO webhooks (
			id, project_id, name, direction, enabled, actor_user_id, engine_type,
			lua_script, ai_prompt, ai_provider_id, event_types_json, token_hash, token_rotated_at,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, hook.ID, hook.ProjectID, hook.Name, hook.Direction, hook.Enabled, hook.ActorUserID,
		hook.Engine.Type, nullableString(hook.Engine.Script), nullableString(hook.Engine.Prompt),
		nullableString(hook.Engine.ProviderID), mustEventTypesJSON(hook.EventTypes), tokenHash, tokenRotatedAt,
		formatTime(hook.CreatedAt), formatTime(hook.UpdatedAt)); err != nil {
		if isUniqueConstraint(err) {
			return CreatedWebhook{}, fmt.Errorf("%w: webhook name already exists in project", ErrValidation)
		}
		return CreatedWebhook{}, fmt.Errorf("insert webhook: %w", err)
	}
	return CreatedWebhook{Webhook: hook, Token: token}, nil
}

func (s *Service) Get(ctx context.Context, principal authz.Principal, webhookID string) (Webhook, error) {
	hook, err := s.get(ctx, webhookID)
	if err != nil {
		return Webhook{}, err
	}
	if err := s.require(principal, hook.ProjectID); err != nil {
		return Webhook{}, err
	}
	return hook, nil
}

func (s *Service) Update(ctx context.Context, principal authz.Principal, webhookID string, input UpdateInput) (Webhook, error) {
	current, err := s.get(ctx, webhookID)
	if err != nil {
		return Webhook{}, err
	}
	if err := s.require(principal, current.ProjectID); err != nil {
		return Webhook{}, err
	}
	if input.Name != nil {
		current.Name = normalizeName(*input.Name)
	}
	if input.Direction != nil {
		current.Direction = strings.TrimSpace(*input.Direction)
	}
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	if input.ActorUserID != nil {
		current.ActorUserID = strings.TrimSpace(*input.ActorUserID)
	}
	if input.EventTypes != nil {
		current.EventTypes = normalizeEventTypes(*input.EventTypes)
	}
	if input.Engine != nil {
		current.Engine = normalizeEngine(*input.Engine)
	}
	current.UpdatedAt = s.now().UTC()
	if err := s.validateWebhook(ctx, current, false); err != nil {
		return Webhook{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE webhooks
		SET name = ?, direction = ?, enabled = ?, actor_user_id = ?, engine_type = ?,
			lua_script = ?, ai_prompt = ?, ai_provider_id = ?, event_types_json = ?,
			token_hash = CASE WHEN ? = 'incoming' THEN token_hash ELSE NULL END,
			token_rotated_at = CASE WHEN ? = 'incoming' THEN token_rotated_at ELSE NULL END,
			updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, current.Name, current.Direction, current.Enabled, current.ActorUserID, current.Engine.Type,
		nullableString(current.Engine.Script), nullableString(current.Engine.Prompt),
		nullableString(current.Engine.ProviderID), mustEventTypesJSON(current.EventTypes),
		current.Direction, current.Direction, formatTime(current.UpdatedAt), webhookID)
	if err != nil {
		if isUniqueConstraint(err) {
			return Webhook{}, fmt.Errorf("%w: webhook name already exists in project", ErrValidation)
		}
		return Webhook{}, fmt.Errorf("update webhook: %w", err)
	}
	if err := requireRowsAffected(result, "webhook update"); err != nil {
		return Webhook{}, err
	}
	return s.get(ctx, webhookID)
}

func (s *Service) Delete(ctx context.Context, principal authz.Principal, webhookID string) error {
	hook, err := s.get(ctx, webhookID)
	if err != nil {
		return err
	}
	if err := s.require(principal, hook.ProjectID); err != nil {
		return err
	}
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE webhooks
		SET deleted_at = ?, enabled = 0, token_hash = NULL, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatTime(now), formatTime(now), webhookID)
	if err != nil {
		return fmt.Errorf("delete webhook: %w", err)
	}
	return requireRowsAffected(result, "webhook delete")
}

func (s *Service) RotateIncomingToken(ctx context.Context, principal authz.Principal, webhookID string) (CreatedWebhook, error) {
	hook, err := s.get(ctx, webhookID)
	if err != nil {
		return CreatedWebhook{}, err
	}
	if err := s.require(principal, hook.ProjectID); err != nil {
		return CreatedWebhook{}, err
	}
	if hook.Direction != DirectionIncoming {
		return CreatedWebhook{}, fmt.Errorf("%w: only incoming webhooks have tokens", ErrValidation)
	}
	secret, err := randomSecret()
	if err != nil {
		return CreatedWebhook{}, err
	}
	token := "wh_" + secret
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE webhooks
		SET token_hash = ?, token_rotated_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, hashSecret(token), formatTime(now), formatTime(now), webhookID)
	if err != nil {
		return CreatedWebhook{}, fmt.Errorf("rotate webhook token: %w", err)
	}
	if err := requireRowsAffected(result, "webhook token rotation"); err != nil {
		return CreatedWebhook{}, err
	}
	hook, err = s.get(ctx, webhookID)
	if err != nil {
		return CreatedWebhook{}, err
	}
	return CreatedWebhook{Webhook: hook, Token: token}, nil
}

func (s *Service) AuthenticateIncoming(ctx context.Context, webhookID string, token string) (Webhook, error) {
	if strings.TrimSpace(token) == "" {
		return Webhook{}, ErrNotFound
	}
	hook, err := scanWebhook(s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, direction, enabled, actor_user_id, engine_type,
			COALESCE(lua_script, ''), COALESCE(ai_prompt, ''), COALESCE(ai_provider_id, ''),
			event_types_json,
			token_hash, token_rotated_at, COALESCE(last_error, ''), created_at, updated_at
		FROM webhooks
		WHERE id = ? AND direction = ? AND enabled = 1 AND token_hash = ? AND deleted_at IS NULL
	`, webhookID, DirectionIncoming, hashSecret(token)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Webhook{}, ErrNotFound
		}
		return Webhook{}, fmt.Errorf("authenticate incoming webhook: %w", err)
	}
	return hook, nil
}

func (s *Service) ReceiveIncoming(ctx context.Context, webhookID string, token string, input IncomingInput) (IncomingResult, error) {
	hook, err := s.AuthenticateIncoming(ctx, webhookID, token)
	if err != nil {
		return IncomingResult{}, err
	}
	if s.runs == nil {
		return IncomingResult{}, errors.New("webhooks: run store is required")
	}
	run, err := s.runs.Start(ctx, automation.StartInput{
		TriggerType: "incoming_webhook",
		TriggerRef:  hook.ID,
		ProjectID:   hook.ProjectID,
		Engine:      hook.Engine.Type,
		ActorUserID: hook.ActorUserID,
		Input: map[string]any{
			"webhook_id": hook.ID,
			"request":    incomingRequestMap(input),
		},
		Limits: map[string]any{
			"timeout_seconds": 10,
		},
	})
	if err != nil {
		return IncomingResult{}, err
	}

	output, logs, execErr := map[string]any{}, []string(nil), s.requireActiveActor(ctx, hook.ActorUserID)
	if execErr == nil {
		output, logs, execErr = s.executeIncoming(ctx, hook, input)
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
		return IncomingResult{}, finishErr
	}
	if err := s.recordRunResult(ctx, hook.ID, finished.Status, finish.Error); err != nil {
		return IncomingResult{}, err
	}
	hook, getErr := s.get(ctx, hook.ID)
	if getErr != nil {
		return IncomingResult{}, getErr
	}
	result := IncomingResult{Webhook: hook, Run: finished}
	if execErr != nil {
		return result, fmt.Errorf("%w: incoming webhook script failed: %v", ErrValidation, execErr)
	}
	return result, nil
}

func (s *Service) ListRuns(ctx context.Context, principal authz.Principal, webhookID string, limit int, offset int) ([]automation.Run, error) {
	hook, err := s.Get(ctx, principal, webhookID)
	if err != nil {
		return nil, err
	}
	if s.runs == nil {
		return nil, errors.New("webhooks: run store is required")
	}
	return s.runs.List(ctx, automation.ListInput{
		TriggerType: "incoming_webhook",
		TriggerRef:  hook.ID,
		ProjectID:   hook.ProjectID,
		Limit:       limit,
		Offset:      offset,
	})
}

func (s *Service) executeIncoming(ctx context.Context, hook Webhook, input IncomingInput) (map[string]any, []string, error) {
	switch hook.Engine.Type {
	case EngineTypeLua:
		return s.executeIncomingLua(ctx, hook, input)
	case EngineTypeAI:
		return s.executeIncomingAI(ctx, hook, input)
	default:
		return map[string]any{}, nil, fmt.Errorf("%w: unsupported engine", ErrValidation)
	}
}

func (s *Service) executeIncomingLua(ctx context.Context, hook Webhook, input IncomingInput) (map[string]any, []string, error) {
	runCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	sandbox := luasandbox.New(luasandbox.DefaultJSONLimits())
	defer sandbox.Close()
	sandbox.L.SetContext(runCtx)

	requestValue, err := sandbox.JSON.FromGo(incomingRequestMap(input))
	if err != nil {
		return map[string]any{}, nil, err
	}
	sandbox.L.SetGlobal("request", requestValue)
	logs := []string{}
	s.registerIncomingLuaHelpers(runCtx, sandbox, hook, &logs)

	fn, err := sandbox.L.LoadString(hook.Engine.Script)
	if err != nil {
		return map[string]any{}, logs, err
	}
	top := sandbox.L.GetTop()
	sandbox.L.Push(fn)
	if err := sandbox.L.PCall(0, lua.MultRet, nil); err != nil {
		return map[string]any{}, logs, err
	}
	output := map[string]any{"ok": true}
	if sandbox.L.GetTop() > top {
		returned := sandbox.L.Get(-1)
		value, err := sandbox.JSON.ToGo(returned)
		if err != nil {
			return map[string]any{}, logs, err
		}
		if object, ok := value.(map[string]any); ok {
			output = object
		} else if value != nil {
			output = map[string]any{"result": value}
		}
	}
	if message := rejectMessage(output); message != "" {
		return output, logs, fmt.Errorf("%w: %s", ErrValidation, message)
	}
	return output, logs, nil
}

func (s *Service) executeIncomingAI(ctx context.Context, hook Webhook, input IncomingInput) (map[string]any, []string, error) {
	if s.openrouter == nil {
		return map[string]any{}, nil, fmt.Errorf("%w: OpenRouter service is not configured", ErrValidation)
	}
	prompt, err := incomingAIPrompt(hook, input)
	if err != nil {
		return map[string]any{}, nil, err
	}
	result, err := s.openrouter.CompleteJSON(ctx, openrouter.CompletionInput{
		ProviderID: hook.Engine.ProviderID,
		Prompt:     prompt,
	})
	if err != nil {
		return map[string]any{}, nil, err
	}
	output := result.Output
	if output == nil {
		output = map[string]any{}
	}
	if message := rejectMessage(output); message != "" {
		return output, nil, fmt.Errorf("%w: %s", ErrValidation, message)
	}
	actionResults, err := s.executeIncomingAIActions(ctx, hook, output["actions"])
	if err != nil {
		return output, nil, err
	}
	if len(actionResults) > 0 {
		output["action_results"] = actionResults
	}
	output["provider_id"] = result.ProviderID
	output["model"] = result.Model
	if result.ResponseID != "" {
		output["response_id"] = result.ResponseID
	}
	if len(result.Usage) > 0 {
		output["usage"] = result.Usage
	}
	return output, nil, nil
}

func incomingAIPrompt(hook Webhook, input IncomingInput) (string, error) {
	payload := map[string]any{
		"context": map[string]any{
			"direction":  hook.Direction,
			"project_id": hook.ProjectID,
			"webhook_id": hook.ID,
			"user_id":    hook.ActorUserID,
		},
		"request": incomingRequestMap(input),
		"instructions": []string{
			"Return only a JSON object.",
			"Return {\"reject\":{\"message\":\"...\"}} to reject the webhook.",
			"Return {\"actions\":[{\"type\":\"create_ticket\",\"input\":{...}}]} to perform allowed Rayboard actions.",
			"Allowed action types are search, get_ticket, create_ticket, update_ticket, and comment.",
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode incoming webhook AI input: %w", err)
	}
	return strings.TrimSpace(hook.Engine.Prompt) + "\n\nRayboard incoming webhook input:\n" + string(data), nil
}

func (s *Service) executeIncomingAIActions(ctx context.Context, hook Webhook, value any) ([]map[string]any, error) {
	if value == nil {
		return nil, nil
	}
	actions, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: actions must be an array", ErrValidation)
	}
	if len(actions) > 20 {
		return nil, fmt.Errorf("%w: actions must contain 20 or fewer items", ErrValidation)
	}
	results := make([]map[string]any, 0, len(actions))
	for index, item := range actions {
		action, ok := item.(map[string]any)
		if !ok {
			return results, fmt.Errorf("%w: action %d must be an object", ErrValidation, index+1)
		}
		actionType := stringValue(action, "type")
		input, _ := action["input"].(map[string]any)
		if input == nil {
			input = map[string]any{}
		}
		result, err := s.executeIncomingAIAction(ctx, hook, actionType, input)
		entry := map[string]any{"type": actionType}
		if err != nil {
			entry["error"] = err.Error()
			results = append(results, entry)
			return results, err
		}
		entry["result"] = result
		results = append(results, entry)
	}
	return results, nil
}

func (s *Service) executeIncomingAIAction(ctx context.Context, hook Webhook, actionType string, input map[string]any) (any, error) {
	switch actionType {
	case "search":
		if s.search == nil {
			return nil, fmt.Errorf("%w: rayboard.search is not configured", ErrValidation)
		}
		return s.search.SearchTickets(ctx, webhookPrincipal(hook), search.SearchTicketsInput{
			ProjectID: stringValue(input, "project_id"),
			Filter:    stringValue(input, "filter"),
			Text:      stringValue(input, "text"),
			Sort:      sortSpecs(input["sort"]),
			Limit:     intValue(input, "limit"),
			Cursor:    stringValue(input, "cursor"),
		})
	case "get_ticket":
		if s.tracker == nil {
			return nil, fmt.Errorf("%w: rayboard.get_ticket is not configured", ErrValidation)
		}
		return s.tracker.GetTicket(ctx, webhookPrincipal(hook), ticketIDValue(input))
	case "create_ticket":
		if s.tracker == nil {
			return nil, fmt.Errorf("%w: rayboard.create_ticket is not configured", ErrValidation)
		}
		customFields, ok := customFieldsValue(input)
		if !ok {
			return nil, fmt.Errorf("%w: custom_fields must be an object", ErrValidation)
		}
		labels, ok := stringSliceValue(input, "labels")
		if !ok {
			return nil, fmt.Errorf("%w: labels must be an array of strings", ErrValidation)
		}
		return s.tracker.CreateTicket(ctx, webhookPrincipal(hook), tracker.CreateTicketInput{
			ProjectID:      stringValue(input, "project_id"),
			Title:          stringValue(input, "title"),
			Description:    stringValue(input, "description"),
			Status:         stringValue(input, "status"),
			Priority:       stringValue(input, "priority"),
			Type:           stringValue(input, "type"),
			ReporterID:     stringValue(input, "reporter_id"),
			AssigneeID:     stringValue(input, "assignee_id"),
			ParentTicketID: stringValue(input, "parent_ticket_id"),
			SprintID:       stringValue(input, "sprint_id"),
			ComponentID:    stringValue(input, "component_id"),
			VersionID:      stringValue(input, "version_id"),
			Rank:           stringValue(input, "rank"),
			StartDate:      stringValue(input, "start_date"),
			DueDate:        stringValue(input, "due_date"),
			Labels:         labels,
			CustomFields:   customFields,
		})
	case "update_ticket":
		if s.tracker == nil {
			return nil, fmt.Errorf("%w: rayboard.update_ticket is not configured", ErrValidation)
		}
		customFields, hasCustomFields, ok := optionalCustomFieldsValue(input)
		if !ok {
			return nil, fmt.Errorf("%w: custom_fields must be an object", ErrValidation)
		}
		labels, hasLabels, ok := optionalStringSliceValue(input, "labels")
		if !ok {
			return nil, fmt.Errorf("%w: labels must be an array of strings", ErrValidation)
		}
		update := tracker.UpdateTicketInput{
			Title:          optionalString(input, "title"),
			Description:    optionalString(input, "description"),
			Status:         optionalString(input, "status"),
			Priority:       optionalString(input, "priority"),
			Type:           optionalString(input, "type"),
			AssigneeID:     optionalString(input, "assignee_id"),
			ParentTicketID: optionalString(input, "parent_ticket_id"),
			SprintID:       optionalString(input, "sprint_id"),
			ComponentID:    optionalString(input, "component_id"),
			VersionID:      optionalString(input, "version_id"),
			Rank:           optionalString(input, "rank"),
			StartDate:      optionalString(input, "start_date"),
			DueDate:        optionalString(input, "due_date"),
		}
		if hasCustomFields {
			update.CustomFields = &customFields
		}
		if hasLabels {
			update.Labels = &labels
		}
		return s.tracker.UpdateTicket(ctx, webhookPrincipal(hook), ticketIDValue(input), update)
	case "comment":
		if s.comments == nil {
			return nil, fmt.Errorf("%w: rayboard.comment is not configured", ErrValidation)
		}
		return s.comments.Create(ctx, webhookPrincipal(hook), comments.CreateInput{
			TicketID: stringValue(input, "ticket_id"),
			Body:     stringValue(input, "body"),
		})
	default:
		return nil, fmt.Errorf("%w: unsupported AI action %q", ErrValidation, actionType)
	}
}

func incomingRequestMap(input IncomingInput) map[string]any {
	return map[string]any{
		"headers": stringAnyMap(input.Headers),
		"query":   stringAnyMap(input.Query),
		"payload": nonNilMap(input.Payload),
	}
}

func (s *Service) recordRunResult(ctx context.Context, webhookID string, status string, lastError string) error {
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE webhooks
		SET last_error = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, nullableString(lastError), formatTime(now), webhookID)
	if err != nil {
		return fmt.Errorf("record webhook run result: %w", err)
	}
	_ = status
	return requireRowsAffected(result, "webhook run result")
}

func (s *Service) get(ctx context.Context, webhookID string) (Webhook, error) {
	hook, err := scanWebhook(s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, direction, enabled, actor_user_id, engine_type,
			COALESCE(lua_script, ''), COALESCE(ai_prompt, ''), COALESCE(ai_provider_id, ''),
			event_types_json,
			token_hash, token_rotated_at, COALESCE(last_error, ''), created_at, updated_at
		FROM webhooks
		WHERE id = ? AND deleted_at IS NULL
	`, webhookID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Webhook{}, ErrNotFound
		}
		return Webhook{}, fmt.Errorf("get webhook: %w", err)
	}
	return hook, nil
}

func (s *Service) validateWebhook(ctx context.Context, hook Webhook, _ bool) error {
	fields := map[string]string{}
	if hook.ProjectID == "" {
		fields["project_id"] = "Required"
	} else if err := s.requireProject(ctx, hook.ProjectID); err != nil {
		return err
	}
	if hook.Name == "" {
		fields["name"] = "Required"
	}
	if len(hook.Name) > 80 {
		fields["name"] = "Must be 80 characters or fewer"
	}
	if !validDirection(hook.Direction) {
		fields["direction"] = "Must be incoming or outgoing"
	}
	if hook.ActorUserID == "" {
		fields["actor_user_id"] = "Required"
	} else if err := s.requireUser(ctx, hook.ActorUserID); err != nil {
		return err
	}
	if err := s.validateEngine(ctx, hook.Engine); err != nil {
		fields["engine"] = err.Error()
	}
	if hook.Direction == DirectionOutgoing && len(hook.EventTypes) == 0 {
		fields["event_types"] = "Outgoing webhooks require at least one event type"
	}
	if len(fields) > 0 {
		return fmt.Errorf("%w: invalid webhook", ErrValidation)
	}
	return nil
}

func (s *Service) validateEngine(ctx context.Context, engine EngineSpec) error {
	switch engine.Type {
	case EngineTypeLua:
		if strings.TrimSpace(engine.Script) == "" {
			return errors.New("lua engine requires script")
		}
	case EngineTypeAI:
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
	if provider.DefaultTimeoutSeconds <= 0 {
		return errors.New("OpenRouter provider timeout must be greater than zero")
	}
	if provider.MaxOutputTokens <= 0 {
		return errors.New("OpenRouter provider max output tokens must be greater than zero")
	}
	return nil
}

func (s *Service) require(principal authz.Principal, projectID string) error {
	if s == nil || s.authorizer == nil {
		return errors.New("webhooks: authorization evaluator is required")
	}
	return s.authorizer.Require(principal, authz.PermissionWebhooksManage, authz.ProjectScope(projectID))
}

func (s *Service) requireProject(ctx context.Context, projectID string) error {
	var exists bool
	if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE id = ? AND deleted_at IS NULL)", projectID).Scan(&exists); err != nil {
		return fmt.Errorf("check webhook project: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func (s *Service) requireUser(ctx context.Context, userID string) error {
	var exists bool
	if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL AND is_disabled = 0)", userID).Scan(&exists); err != nil {
		return fmt.Errorf("check webhook actor user: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func (s *Service) requireActiveActor(ctx context.Context, userID string) error {
	var exists bool
	if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL AND is_disabled = 0)", userID).Scan(&exists); err != nil {
		return fmt.Errorf("check webhook actor user: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: webhook actor user is disabled or deleted", ErrValidation)
	}
	return nil
}

func scanWebhook(scanner interface{ Scan(...any) error }) (Webhook, error) {
	var hook Webhook
	var tokenHash sql.NullString
	var tokenRotatedAt sql.NullString
	var eventTypesJSON string
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&hook.ID,
		&hook.ProjectID,
		&hook.Name,
		&hook.Direction,
		&hook.Enabled,
		&hook.ActorUserID,
		&hook.Engine.Type,
		&hook.Engine.Script,
		&hook.Engine.Prompt,
		&hook.Engine.ProviderID,
		&eventTypesJSON,
		&tokenHash,
		&tokenRotatedAt,
		&hook.LastError,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Webhook{}, err
	}
	hook.TokenSet = tokenHash.Valid && tokenHash.String != ""
	hook.TokenRotatedAt = parseNullableTime(tokenRotatedAt)
	eventTypes, err := unmarshalEventTypes(eventTypesJSON)
	if err != nil {
		return Webhook{}, err
	}
	hook.EventTypes = eventTypes
	created, err := parseTime(createdAt)
	if err != nil {
		return Webhook{}, err
	}
	updated, err := parseTime(updatedAt)
	if err != nil {
		return Webhook{}, err
	}
	hook.CreatedAt = created
	hook.UpdatedAt = updated
	return hook, nil
}

func normalizeEngine(engine EngineSpec) EngineSpec {
	engine.Type = strings.ToLower(strings.TrimSpace(engine.Type))
	engine.Script = strings.TrimSpace(engine.Script)
	engine.Prompt = strings.TrimSpace(engine.Prompt)
	engine.ProviderID = strings.TrimSpace(engine.ProviderID)
	return engine
}

func normalizeEventTypes(eventTypes []string) []string {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		eventType = strings.TrimSpace(eventType)
		if eventType == "" {
			continue
		}
		if _, ok := seen[eventType]; ok {
			continue
		}
		seen[eventType] = struct{}{}
		normalized = append(normalized, eventType)
	}
	return normalized
}

func mustEventTypesJSON(eventTypes []string) string {
	data, err := json.Marshal(normalizeEventTypes(eventTypes))
	if err != nil {
		return "[]"
	}
	return string(data)
}

func unmarshalEventTypes(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var eventTypes []string
	if err := json.Unmarshal([]byte(value), &eventTypes); err != nil {
		return nil, fmt.Errorf("decode webhook event types: %w", err)
	}
	return normalizeEventTypes(eventTypes), nil
}

func validDirection(direction string) bool {
	return direction == DirectionIncoming || direction == DirectionOutgoing
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func normalizeList(limit int, offset int) (int, int, error) {
	if limit < 0 || offset < 0 {
		return 0, 0, fmt.Errorf("%w: limit and offset must be non-negative", ErrValidation)
	}
	if limit == 0 {
		limit = 50
	}
	if limit > 200 {
		return 0, 0, fmt.Errorf("%w: limit must be 200 or fewer", ErrValidation)
	}
	return limit, offset, nil
}

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func stringAnyMap(value map[string]string) map[string]any {
	result := map[string]any{}
	for key, item := range value {
		result[key] = item
	}
	return result
}

func rejectMessage(output map[string]any) string {
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

func joinAnd(parts []string) string {
	return strings.Join(parts, " AND ")
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse webhook time: %w", err)
	}
	return parsed, nil
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}
	parsed, err := parseTime(value.String)
	if err != nil {
		return nil
	}
	return &parsed
}

func randomSecret() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate webhook token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func newID(prefix string) string {
	secret, err := randomSecret()
	if err != nil {
		return prefix + "_fallback"
	}
	return prefix + "_" + secret[:22]
}

func isUniqueConstraint(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "unique constraint") || strings.Contains(text, "constraint failed")
}

func requireRowsAffected(result sql.Result, action string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check %s: %w", action, err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
