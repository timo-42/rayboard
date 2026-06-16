package cronjobs

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
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
	Engine        string     `json:"engine"`
	LuaSource     string     `json:"lua_source,omitempty"`
	AIPrompt      string     `json:"ai_prompt,omitempty"`
	LastRunStatus string     `json:"last_run_status,omitempty"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	NextRunAt     *time.Time `json:"next_run_at,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateInput struct {
	OwnerUserID string `json:"owner_user_id"`
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Schedule    string `json:"schedule"`
	Timezone    string `json:"timezone"`
	Enabled     bool   `json:"enabled"`
	Engine      string `json:"engine"`
	LuaSource   string `json:"lua_source"`
	AIPrompt    string `json:"ai_prompt"`
}

type UpdateInput struct {
	OwnerUserID *string `json:"owner_user_id"`
	ProjectID   *string `json:"project_id"`
	Name        *string `json:"name"`
	Schedule    *string `json:"schedule"`
	Timezone    *string `json:"timezone"`
	Enabled     *bool   `json:"enabled"`
	Engine      *string `json:"engine"`
	LuaSource   *string `json:"lua_source"`
	AIPrompt    *string `json:"ai_prompt"`
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
			lua_source, ai_prompt, last_run_status, last_run_at, next_run_at, last_error,
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
			lua_source, ai_prompt, next_run_at, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.OwnerUserID, nullableString(job.ProjectID), job.Name, job.Schedule, job.Timezone, job.Enabled, job.Engine, job.LuaSource, job.AIPrompt, nullableTime(job.NextRunAt), formatTime(job.CreatedAt), formatTime(job.UpdatedAt)); err != nil {
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
			enabled = ?, engine = ?, lua_source = ?, ai_prompt = ?, next_run_at = ?, updated_at = ?
		WHERE id = ?
	`, updated.OwnerUserID, nullableString(updated.ProjectID), updated.Name, updated.Schedule, updated.Timezone, updated.Enabled, updated.Engine, updated.LuaSource, updated.AIPrompt, nullableTime(updated.NextRunAt), formatTime(updated.UpdatedAt), updated.ID); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		Engine:      job.Engine,
		ActorUserID: job.OwnerUserID,
		Input: map[string]any{
			"mode":     mode,
			"job_id":   job.ID,
			"job_name": job.Name,
		},
		Limits: map[string]any{
			"timeout_seconds": 30,
		},
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
	switch job.Engine {
	case EngineLua:
		return s.executeLua(ctx, job)
	case EngineAI:
		return nil, nil, fmt.Errorf("%w: AI cron jobs are not implemented yet", ErrValidation)
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

	if err := sandbox.L.DoString(job.LuaSource); err != nil {
		return map[string]any{}, logs, err
	}
	return map[string]any{"ok": true}, logs, nil
}

func (s *Service) validateCreate(ctx context.Context, input CreateInput) (Job, error) {
	job := Job{
		OwnerUserID: input.OwnerUserID,
		ProjectID:   strings.TrimSpace(input.ProjectID),
		Name:        strings.TrimSpace(input.Name),
		Schedule:    strings.TrimSpace(input.Schedule),
		Timezone:    strings.TrimSpace(input.Timezone),
		Enabled:     input.Enabled,
		Engine:      strings.TrimSpace(input.Engine),
		LuaSource:   input.LuaSource,
		AIPrompt:    input.AIPrompt,
	}
	if job.Timezone == "" {
		job.Timezone = "UTC"
	}
	if job.Engine == "" {
		job.Engine = EngineLua
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
	switch job.Engine {
	case EngineLua:
		if strings.TrimSpace(job.LuaSource) == "" {
			fields["lua_source"] = "Required for lua engine"
		}
	case EngineAI:
		fields["engine"] = "AI cron jobs are planned but not implemented"
	default:
		fields["engine"] = "Must be lua or ai"
	}
	if len(fields) > 0 {
		return &ValidationError{Message: "Invalid cron job", Fields: fields}
	}
	return nil
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
			lua_source, ai_prompt, last_run_status, last_run_at, next_run_at, last_error,
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
			lua_source, ai_prompt, last_run_status, last_run_at, next_run_at, last_error,
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
		&job.Engine,
		&job.LuaSource,
		&job.AIPrompt,
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
		job.Engine = strings.TrimSpace(*input.Engine)
	}
	if input.LuaSource != nil {
		job.LuaSource = *input.LuaSource
	}
	if input.AIPrompt != nil {
		job.AIPrompt = *input.AIPrompt
	}
	return job
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
