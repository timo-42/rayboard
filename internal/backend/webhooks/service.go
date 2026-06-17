package webhooks

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/luasandbox"
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
)

type Service struct {
	db         *sql.DB
	authorizer authz.Evaluator
	runs       *automation.RunStore
	tracker    *tracker.Service
	search     *search.Service
	comments   *comments.Service
	now        func() time.Time
}

type Option func(*Service)

func NewService(db *sql.DB, authorizer authz.Evaluator, options ...Option) *Service {
	service := &Service{
		db:         db,
		authorizer: authorizer,
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
	Engine      EngineSpec
}

type UpdateInput struct {
	Name        *string
	Enabled     *bool
	ActorUserID *string
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
			lua_script, ai_prompt, ai_provider_id, token_hash, token_rotated_at,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, hook.ID, hook.ProjectID, hook.Name, hook.Direction, hook.Enabled, hook.ActorUserID,
		hook.Engine.Type, nullableString(hook.Engine.Script), nullableString(hook.Engine.Prompt),
		nullableString(hook.Engine.ProviderID), tokenHash, tokenRotatedAt,
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
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	if input.ActorUserID != nil {
		current.ActorUserID = strings.TrimSpace(*input.ActorUserID)
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
		SET name = ?, enabled = ?, actor_user_id = ?, engine_type = ?,
			lua_script = ?, ai_prompt = ?, ai_provider_id = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, current.Name, current.Enabled, current.ActorUserID, current.Engine.Type,
		nullableString(current.Engine.Script), nullableString(current.Engine.Prompt),
		nullableString(current.Engine.ProviderID), formatTime(current.UpdatedAt), webhookID)
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
		return map[string]any{}, nil, fmt.Errorf("%w: AI incoming webhooks are not implemented yet", ErrValidation)
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

func (s *Service) validateWebhook(ctx context.Context, hook Webhook, creating bool) error {
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
	if err := validateEngine(hook.Engine); err != nil {
		fields["engine"] = err.Error()
	}
	if hook.Direction == DirectionOutgoing && creating {
		fields["direction"] = "Outgoing webhooks are not implemented yet"
	}
	if len(fields) > 0 {
		return fmt.Errorf("%w: invalid webhook", ErrValidation)
	}
	return nil
}

func validateEngine(engine EngineSpec) error {
	switch engine.Type {
	case EngineTypeLua:
		if strings.TrimSpace(engine.Script) == "" {
			return errors.New("lua engine requires script")
		}
	case EngineTypeAI:
		if strings.TrimSpace(engine.Prompt) == "" || strings.TrimSpace(engine.ProviderID) == "" {
			return errors.New("ai engine requires prompt and provider_id")
		}
	default:
		return errors.New("engine type must be lua or ai")
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
