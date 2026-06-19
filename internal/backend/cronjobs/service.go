package cronjobs

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/luasandbox"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

const (
	EngineLua = "lua"
	EngineAI  = "ai"
)

var (
	ErrNotFound   = errors.New("cronjobs: job not found")
	ErrValidation = errors.New("cronjobs: validation failed")
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

type Job struct {
	ID            string     `json:"id"`
	OwnerUserID   string     `json:"owner_user_id"`
	ProjectID     string     `json:"project_id,omitempty"`
	Name          string     `json:"name"`
	Schedule      string     `json:"schedule"`
	Timezone      string     `json:"timezone"`
	Enabled       bool       `json:"enabled"`
	Engine        EngineSpec `json:"engine"`
	LastRunStatus string     `json:"last_run_status,omitempty"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	NextRunAt     *time.Time `json:"next_run_at,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type EngineSpec struct {
	Type       string `json:"type" enum:"lua,ai" doc:"Execution engine discriminator."`
	Script     string `json:"script,omitempty" doc:"Lua script source. Required when type is lua."`
	Prompt     string `json:"prompt,omitempty" doc:"AI prompt. Required when type is ai."`
	ProviderID string `json:"provider_id,omitempty" doc:"AI provider configuration ID. Required when type is ai."`
}

type CreateInput struct {
	OwnerUserID string     `json:"owner_user_id,omitempty"`
	ProjectID   string     `json:"project_id,omitempty"`
	Name        string     `json:"name,omitempty"`
	Schedule    string     `json:"schedule,omitempty"`
	Timezone    string     `json:"timezone,omitempty"`
	Enabled     bool       `json:"enabled,omitempty"`
	Engine      EngineSpec `json:"engine"`
}

type UpdateInput struct {
	OwnerUserID *string     `json:"owner_user_id,omitempty"`
	ProjectID   *string     `json:"project_id,omitempty"`
	Name        *string     `json:"name,omitempty"`
	Schedule    *string     `json:"schedule,omitempty"`
	Timezone    *string     `json:"timezone,omitempty"`
	Enabled     *bool       `json:"enabled,omitempty"`
	Engine      *EngineSpec `json:"engine,omitempty"`
}

type ListInput struct {
	ProjectID string
	Limit     int
	Offset    int
}

type Service struct {
	db         *sql.DB
	authorizer authz.Evaluator
	runs       *automation.RunStore
	tracker    *tracker.Service
	search     *search.Service
	comments   *comments.Service
	openrouter *openrouter.Service
	now        func() time.Time

	mu      sync.Mutex
	started bool
	cron    *cron.Cron
	entries map[string]cron.EntryID
}

type Option func(*Service)

func NewService(db *sql.DB, authorizer authz.Evaluator, runStore *automation.RunStore, options ...Option) *Service {
	service := &Service{
		db:         db,
		authorizer: authorizer,
		runs:       runStore,
		now:        func() time.Time { return time.Now().UTC() },
		entries:    map[string]cron.EntryID{},
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

func (s *Service) StartScheduler(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	s.cron = cron.New(
		cron.WithChain(cron.SkipIfStillRunning(cron.PrintfLogger(log.New(io.Discard, "", 0)))),
	)
	s.started = true
	jobs, err := s.listEnabled(ctx)
	if err != nil {
		s.started = false
		s.cron = nil
		return err
	}
	for _, job := range jobs {
		if err := s.scheduleLocked(job); err != nil {
			s.started = false
			s.cron = nil
			return err
		}
	}
	s.cron.Start()
	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	cronRunner := s.cron
	s.started = false
	s.cron = nil
	s.entries = map[string]cron.EntryID{}
	s.mu.Unlock()
	if cronRunner == nil {
		return nil
	}
	stopCtx := cronRunner.Stop()
	select {
	case <-stopCtx.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) List(ctx context.Context, principal authz.Principal, input ListInput) ([]Job, error) {
	limit, offset, err := normalizeList(input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}
	if input.ProjectID != "" {
		if err := s.requireManage(principal, input.ProjectID); err != nil {
			return nil, err
		}
	} else if err := s.authorizer.Require(principal, authz.PermissionAutomationsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}

	where := []string{"1 = 1"}
	args := []any{}
	if input.ProjectID != "" {
		where = append(where, "project_id = ?")
		args = append(args, input.ProjectID)
	}
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, owner_user_id, project_id, name, schedule, timezone, enabled, engine,
			lua_source, ai_prompt, ai_provider_id, last_run_status, last_run_at, next_run_at, last_error,
			created_at, updated_at
		FROM cron_jobs
		WHERE `+joinAnd(where)+`
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list cron jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cron jobs: %w", err)
	}
	return jobs, nil
}

func (s *Service) Create(ctx context.Context, principal authz.Principal, input CreateInput) (Job, error) {
	input.OwnerUserID = strings.TrimSpace(input.OwnerUserID)
	if input.OwnerUserID == "" {
		input.OwnerUserID = principal.UserID
	}
	job, err := s.validateCreate(ctx, input)
	if err != nil {
		return Job{}, err
	}
	if err := s.requireManage(principal, job.ProjectID); err != nil {
		return Job{}, err
	}
	now := s.now().UTC()
	job.ID = newID("cron")
	job.CreatedAt = now
	job.UpdatedAt = now
	job.NextRunAt = nextRun(job.Schedule, job.Timezone, now)
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO cron_jobs (
			id, owner_user_id, project_id, name, schedule, timezone, enabled, engine,
			lua_source, ai_prompt, ai_provider_id, next_run_at, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.OwnerUserID, nullableString(job.ProjectID), job.Name, job.Schedule, job.Timezone, job.Enabled, job.Engine.Type, job.Engine.Script, job.Engine.Prompt, job.Engine.ProviderID, nullableTime(job.NextRunAt), formatTime(job.CreatedAt), formatTime(job.UpdatedAt)); err != nil {
		return Job{}, fmt.Errorf("insert cron job: %w", err)
	}
	s.refreshSchedule(job)
	return job, nil
}

func (s *Service) Get(ctx context.Context, principal authz.Principal, jobID string) (Job, error) {
	job, err := s.get(ctx, jobID)
	if err != nil {
		return Job{}, err
	}
	if err := s.requireManage(principal, job.ProjectID); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (s *Service) Update(ctx context.Context, principal authz.Principal, jobID string, input UpdateInput) (Job, error) {
	current, err := s.get(ctx, jobID)
	if err != nil {
		return Job{}, err
	}
	if err := s.requireManage(principal, current.ProjectID); err != nil {
		return Job{}, err
	}
	updated := applyUpdate(current, input)
	if err := s.validateJob(ctx, updated); err != nil {
		return Job{}, err
	}
	if err := s.requireManage(principal, updated.ProjectID); err != nil {
		return Job{}, err
	}
	updated.UpdatedAt = s.now().UTC()
	updated.NextRunAt = nextRun(updated.Schedule, updated.Timezone, updated.UpdatedAt)
	if _, err := s.db.ExecContext(ctx, `
		UPDATE cron_jobs
		SET owner_user_id = ?, project_id = ?, name = ?, schedule = ?, timezone = ?,
			enabled = ?, engine = ?, lua_source = ?, ai_prompt = ?, ai_provider_id = ?, next_run_at = ?, updated_at = ?
		WHERE id = ?
	`, updated.OwnerUserID, nullableString(updated.ProjectID), updated.Name, updated.Schedule, updated.Timezone, updated.Enabled, updated.Engine.Type, updated.Engine.Script, updated.Engine.Prompt, updated.Engine.ProviderID, nullableTime(updated.NextRunAt), formatTime(updated.UpdatedAt), updated.ID); err != nil {
		return Job{}, fmt.Errorf("update cron job: %w", err)
	}
	s.refreshSchedule(updated)
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, principal authz.Principal, jobID string) error {
	job, err := s.get(ctx, jobID)
	if err != nil {
		return err
	}
	if err := s.requireManage(principal, job.ProjectID); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM cron_jobs WHERE id = ?", jobID)
	if err != nil {
		return fmt.Errorf("delete cron job: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check cron job delete: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	s.unschedule(jobID)
	return nil
}

func (s *Service) RunNow(ctx context.Context, principal authz.Principal, jobID string) (automation.Run, error) {
	job, err := s.Get(ctx, principal, jobID)
	if err != nil {
		return automation.Run{}, err
	}
	return s.runJob(ctx, job, "manual")
}

func (s *Service) ListRuns(ctx context.Context, principal authz.Principal, jobID string, limit int, offset int) ([]automation.Run, error) {
	job, err := s.Get(ctx, principal, jobID)
	if err != nil {
		return nil, err
	}
	return s.runs.List(ctx, automation.ListInput{
		TriggerType: "cron",
		TriggerRef:  job.ID,
		ProjectID:   job.ProjectID,
		Status:      "",
		Limit:       limit,
		Offset:      offset,
	})
}

func (s *Service) runScheduled(jobID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	job, err := s.get(ctx, jobID)
	if err != nil || !job.Enabled {
		return
	}
	_, _ = s.runJob(ctx, job, "schedule")
}

func (s *Service) runJob(ctx context.Context, job Job, mode string) (automation.Run, error) {
	if err := s.ownerCanRun(ctx, job.OwnerUserID); err != nil {
		return automation.Run{}, err
	}
	run, err := s.runs.Start(ctx, automation.StartInput{
		TriggerType: "cron",
		TriggerRef:  job.ID,
		ProjectID:   job.ProjectID,
		Engine:      job.Engine.Type,
		ActorUserID: job.OwnerUserID,
		Input: map[string]any{
			"mode":     mode,
			"job_id":   job.ID,
			"job_name": job.Name,
		},
		Limits: s.runLimits(ctx, job),
	})
	if err != nil {
		return automation.Run{}, err
	}

	output, logs, execErr := s.execute(ctx, job)
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
	_ = s.recordRunResult(ctx, job.ID, finished.Status, finish.Error)
	if execErr != nil {
		return finished, execErr
	}
	return finished, nil
}

func (s *Service) execute(ctx context.Context, job Job) (map[string]any, []string, error) {
	switch job.Engine.Type {
	case EngineLua:
		return s.executeLua(ctx, job)
	case EngineAI:
		return s.executeAI(ctx, job)
	default:
		return nil, nil, fmt.Errorf("%w: unsupported engine", ErrValidation)
	}
}

func (s *Service) executeLua(ctx context.Context, job Job) (map[string]any, []string, error) {
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	sandbox := luasandbox.New(luasandbox.DefaultJSONLimits())
	defer sandbox.Close()
	sandbox.L.SetContext(runCtx)

	logs := []string{}
	s.registerLuaHelpers(runCtx, sandbox, job, &logs)

	if err := sandbox.L.DoString(job.Engine.Script); err != nil {
		return map[string]any{}, logs, err
	}
	return map[string]any{"ok": true}, logs, nil
}

func (s *Service) executeAI(ctx context.Context, job Job) (map[string]any, []string, error) {
	if s.openrouter == nil {
		return map[string]any{}, nil, fmt.Errorf("%w: OpenRouter service is not configured", ErrValidation)
	}
	prompt, err := cronAIPrompt(job)
	if err != nil {
		return map[string]any{}, nil, err
	}
	result, err := s.openrouter.CompleteJSON(ctx, openrouter.CompletionInput{
		ProviderID: job.Engine.ProviderID,
		Prompt:     prompt,
	})
	if err != nil {
		return map[string]any{}, nil, err
	}
	aiOutput := result.Output
	if aiOutput == nil {
		aiOutput = map[string]any{}
	}
	actionResults, actionErr := s.executeAIActions(ctx, job, aiOutput["actions"])
	if len(actionResults) > 0 {
		aiOutput["action_results"] = actionResults
	}
	output := map[string]any{
		"ok":          true,
		"provider_id": result.ProviderID,
		"model":       result.Model,
		"output":      aiOutput,
	}
	if result.ResponseID != "" {
		output["response_id"] = result.ResponseID
	}
	if len(result.Usage) > 0 {
		output["usage"] = result.Usage
	}
	if actionErr != nil {
		return output, nil, actionErr
	}
	return output, nil, nil
}

func cronAIPrompt(job Job) (string, error) {
	payload := map[string]any{
		"context": map[string]any{
			"surface":    "cron",
			"job_id":     job.ID,
			"job_name":   job.Name,
			"project_id": job.ProjectID,
			"user_id":    job.OwnerUserID,
		},
		"instructions": []string{
			"Return only a JSON object.",
			"Return {\"actions\":[{\"type\":\"create_ticket\",\"input\":{...}}]} to perform allowed Rayboard actions.",
			"Allowed action types are search, get_ticket, create_ticket, update_ticket, and comment.",
			"Actions run as the cron job owner and must pass normal Rayboard validation and RBAC.",
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode cron AI input: %w", err)
	}
	return strings.TrimSpace(job.Engine.Prompt) + "\n\nRayboard cron input:\n" + string(data), nil
}

func (s *Service) executeAIActions(ctx context.Context, job Job, value any) ([]map[string]any, error) {
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
		result, err := s.executeAIAction(ctx, job, actionType, input)
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

func (s *Service) executeAIAction(ctx context.Context, job Job, actionType string, input map[string]any) (any, error) {
	switch actionType {
	case "search":
		if s.search == nil {
			return nil, fmt.Errorf("%w: rayboard.search is not configured", ErrValidation)
		}
		return s.search.SearchTickets(ctx, cronPrincipal(job), search.SearchTicketsInput{
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
		return s.tracker.GetTicket(ctx, cronPrincipal(job), ticketIDValue(input))
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
		return s.tracker.CreateTicket(ctx, cronPrincipal(job), tracker.CreateTicketInput{
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
			StoryPoints:    storyPointsValue(input, "story_points"),
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
			StoryPoints:    storyPointsValue(input, "story_points"),
			StoryPointsSet: hasKey(input, "story_points"),
		}
		if hasCustomFields {
			update.CustomFields = &customFields
		}
		if hasLabels {
			update.Labels = &labels
		}
		return s.tracker.UpdateTicket(ctx, cronPrincipal(job), ticketIDValue(input), update)
	case "comment":
		if s.comments == nil {
			return nil, fmt.Errorf("%w: rayboard.comment is not configured", ErrValidation)
		}
		return s.comments.Create(ctx, cronPrincipal(job), comments.CreateInput{
			TicketID: stringValue(input, "ticket_id"),
			Body:     stringValue(input, "body"),
		})
	default:
		return nil, fmt.Errorf("%w: unsupported AI action %q", ErrValidation, actionType)
	}
}

func (s *Service) validateCreate(ctx context.Context, input CreateInput) (Job, error) {
	job := Job{
		OwnerUserID: input.OwnerUserID,
		ProjectID:   strings.TrimSpace(input.ProjectID),
		Name:        strings.TrimSpace(input.Name),
		Schedule:    strings.TrimSpace(input.Schedule),
		Timezone:    strings.TrimSpace(input.Timezone),
		Enabled:     input.Enabled,
		Engine:      normalizeEngine(input.Engine),
	}
	if job.Timezone == "" {
		job.Timezone = "UTC"
	}
	if err := s.validateJob(ctx, job); err != nil {
		return Job{}, err
	}
	return job, nil
}

func (s *Service) validateJob(ctx context.Context, job Job) error {
	fields := map[string]string{}
	if job.OwnerUserID == "" {
		fields["owner_user_id"] = "Required"
	} else if err := s.ownerCanRun(ctx, job.OwnerUserID); err != nil {
		fields["owner_user_id"] = "Owner must exist and be enabled"
	}
	if job.Name == "" {
		fields["name"] = "Required"
	}
	if job.Schedule == "" {
		fields["schedule"] = "Required"
	} else if _, err := parseSchedule(job.Schedule, job.Timezone); err != nil {
		fields["schedule"] = "Invalid cron schedule"
	}
	if job.Timezone == "" {
		fields["timezone"] = "Required"
	} else if _, err := time.LoadLocation(job.Timezone); err != nil {
		fields["timezone"] = "Invalid timezone"
	}
	switch job.Engine.Type {
	case EngineLua:
		if strings.TrimSpace(job.Engine.Script) == "" {
			fields["engine.script"] = "Required for lua engine"
		}
	case EngineAI:
		if strings.TrimSpace(job.Engine.Prompt) == "" {
			fields["engine.prompt"] = "Required for ai engine"
		}
		if strings.TrimSpace(job.Engine.ProviderID) == "" {
			fields["engine.provider_id"] = "Required for ai engine"
		}
		if _, ok := fields["engine.provider_id"]; !ok {
			if err := s.validateAIProvider(ctx, job.Engine.ProviderID); err != nil {
				fields["engine.provider_id"] = err.Error()
			}
		}
	default:
		fields["engine.type"] = "Must be lua or ai"
	}
	if len(fields) > 0 {
		return &ValidationError{Message: "Invalid cron job", Fields: fields}
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

func (s *Service) runLimits(ctx context.Context, job Job) map[string]any {
	limits := map[string]any{"timeout_seconds": 30}
	if job.Engine.Type != EngineAI || s.openrouter == nil {
		return limits
	}
	provider, err := s.openrouter.GetExecutionProvider(ctx, job.Engine.ProviderID)
	if err != nil {
		return limits
	}
	limits["timeout_seconds"] = provider.DefaultTimeoutSeconds
	limits["max_output_tokens"] = provider.MaxOutputTokens
	limits["provider_id"] = provider.ID
	limits["model"] = provider.DefaultModel
	return limits
}

func (s *Service) ownerCanRun(ctx context.Context, userID string) error {
	var disabled bool
	if err := s.db.QueryRowContext(ctx, `
		SELECT is_disabled
		FROM users
		WHERE id = ? AND deleted_at IS NULL
	`, userID).Scan(&disabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("get cron owner: %w", err)
	}
	if disabled {
		return fmt.Errorf("%w: owner is disabled", ErrValidation)
	}
	return nil
}

func (s *Service) requireManage(principal authz.Principal, projectID string) error {
	if projectID != "" {
		return s.authorizer.Require(principal, authz.PermissionAutomationsManage, authz.ProjectScope(projectID))
	}
	return s.authorizer.Require(principal, authz.PermissionAutomationsManage, authz.GlobalScope())
}

func (s *Service) get(ctx context.Context, jobID string) (Job, error) {
	job, err := scanJob(s.db.QueryRowContext(ctx, `
		SELECT id, owner_user_id, project_id, name, schedule, timezone, enabled, engine,
			lua_source, ai_prompt, ai_provider_id, last_run_status, last_run_at, next_run_at, last_error,
			created_at, updated_at
		FROM cron_jobs
		WHERE id = ?
	`, jobID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Job{}, ErrNotFound
		}
		return Job{}, fmt.Errorf("get cron job: %w", err)
	}
	return job, nil
}

func (s *Service) listEnabled(ctx context.Context) ([]Job, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, owner_user_id, project_id, name, schedule, timezone, enabled, engine,
			lua_source, ai_prompt, ai_provider_id, last_run_status, last_run_at, next_run_at, last_error,
			created_at, updated_at
		FROM cron_jobs
		WHERE enabled = 1
	`)
	if err != nil {
		return nil, fmt.Errorf("list enabled cron jobs: %w", err)
	}
	defer rows.Close()
	var jobs []Job
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate enabled cron jobs: %w", err)
	}
	return jobs, nil
}

func scanJob(scanner interface{ Scan(...any) error }) (Job, error) {
	var job Job
	var projectID sql.NullString
	var lastStatus sql.NullString
	var lastRunAt sql.NullString
	var nextRunAt sql.NullString
	var lastError sql.NullString
	var engineType string
	var script string
	var prompt string
	var providerID string
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&job.ID,
		&job.OwnerUserID,
		&projectID,
		&job.Name,
		&job.Schedule,
		&job.Timezone,
		&job.Enabled,
		&engineType,
		&script,
		&prompt,
		&providerID,
		&lastStatus,
		&lastRunAt,
		&nextRunAt,
		&lastError,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Job{}, err
	}
	job.ProjectID = nullString(projectID)
	job.Engine = EngineSpec{
		Type:       engineType,
		Script:     script,
		Prompt:     prompt,
		ProviderID: providerID,
	}
	job.LastRunStatus = nullString(lastStatus)
	job.LastRunAt = parseNullableTime(lastRunAt)
	job.NextRunAt = parseNullableTime(nextRunAt)
	job.LastError = nullString(lastError)
	created, err := parseTime(createdAt)
	if err != nil {
		return Job{}, err
	}
	updated, err := parseTime(updatedAt)
	if err != nil {
		return Job{}, err
	}
	job.CreatedAt = created
	job.UpdatedAt = updated
	return job, nil
}

func applyUpdate(job Job, input UpdateInput) Job {
	if input.OwnerUserID != nil {
		job.OwnerUserID = strings.TrimSpace(*input.OwnerUserID)
	}
	if input.ProjectID != nil {
		job.ProjectID = strings.TrimSpace(*input.ProjectID)
	}
	if input.Name != nil {
		job.Name = strings.TrimSpace(*input.Name)
	}
	if input.Schedule != nil {
		job.Schedule = strings.TrimSpace(*input.Schedule)
	}
	if input.Timezone != nil {
		job.Timezone = strings.TrimSpace(*input.Timezone)
	}
	if input.Enabled != nil {
		job.Enabled = *input.Enabled
	}
	if input.Engine != nil {
		job.Engine = normalizeEngine(*input.Engine)
	}
	return job
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

func (s *Service) refreshSchedule(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started || s.cron == nil {
		return
	}
	if entryID, ok := s.entries[job.ID]; ok {
		s.cron.Remove(entryID)
		delete(s.entries, job.ID)
	}
	if !job.Enabled {
		return
	}
	_ = s.scheduleLocked(job)
}

func (s *Service) unschedule(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cron == nil {
		return
	}
	if entryID, ok := s.entries[jobID]; ok {
		s.cron.Remove(entryID)
		delete(s.entries, jobID)
	}
}

func (s *Service) scheduleLocked(job Job) error {
	spec := scheduleSpec(job.Schedule, job.Timezone)
	entryID, err := s.cron.AddFunc(spec, func() {
		s.runScheduled(job.ID)
	})
	if err != nil {
		return fmt.Errorf("schedule cron job %s: %w", job.ID, err)
	}
	s.entries[job.ID] = entryID
	return nil
}

func (s *Service) recordRunResult(ctx context.Context, jobID string, status string, message string) error {
	now := s.now().UTC()
	job, err := s.get(ctx, jobID)
	if err != nil {
		return err
	}
	next := nextRun(job.Schedule, job.Timezone, now)
	_, err = s.db.ExecContext(ctx, `
		UPDATE cron_jobs
		SET last_run_status = ?, last_run_at = ?, next_run_at = ?, last_error = ?, updated_at = ?
		WHERE id = ?
	`, status, formatTime(now), nullableTime(next), nullableString(message), formatTime(now), jobID)
	if err != nil {
		return fmt.Errorf("record cron run result: %w", err)
	}
	return nil
}

func parseSchedule(schedule string, timezone string) (cron.Schedule, error) {
	return cron.ParseStandard(scheduleSpec(schedule, timezone))
}

func scheduleSpec(schedule string, timezone string) string {
	if timezone == "" || timezone == "UTC" {
		return schedule
	}
	return "CRON_TZ=" + timezone + " " + schedule
}

func nextRun(schedule string, timezone string, from time.Time) *time.Time {
	parsed, err := parseSchedule(schedule, timezone)
	if err != nil {
		return nil
	}
	next := parsed.Next(from)
	return &next
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

func joinAnd(parts []string) string {
	result := ""
	for index, part := range parts {
		if index > 0 {
			result += " AND "
		}
		result += part
	}
	return result
}

func newID(prefix string) string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic(fmt.Sprintf("generate %s id: %v", prefix, err))
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(raw[:])
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatTime(*value)
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse cron job time: %w", err)
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
