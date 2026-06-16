package automation

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"
)

var (
	ErrNotFound   = errors.New("automation: run not found")
	ErrValidation = errors.New("automation: validation failed")
)

type Run struct {
	ID          string         `json:"id"`
	TriggerType string         `json:"trigger_type"`
	TriggerRef  string         `json:"trigger_ref,omitempty"`
	ProjectID   string         `json:"project_id,omitempty"`
	TicketID    string         `json:"ticket_id,omitempty"`
	Status      string         `json:"status"`
	Input       map[string]any `json:"input"`
	Output      map[string]any `json:"output"`
	Error       string         `json:"error,omitempty"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	FinishedAt  *time.Time     `json:"finished_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type StartInput struct {
	TriggerType string
	TriggerRef  string
	ProjectID   string
	TicketID    string
	Engine      string
	ActorUserID string
	Input       map[string]any
	Limits      map[string]any
}

type FinishInput struct {
	Status string
	Output map[string]any
	Error  string
	Logs   []string
}

type ListInput struct {
	TriggerType string
	TriggerRef  string
	ProjectID   string
	TicketID    string
	Status      string
	Limit       int
	Offset      int
}

type RunStore struct {
	db  *sql.DB
	now func() time.Time
}

type Option func(*RunStore)

func NewRunStore(db *sql.DB, options ...Option) *RunStore {
	store := &RunStore{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		option(store)
	}
	return store
}

func WithNow(now func() time.Time) Option {
	return func(store *RunStore) {
		if now != nil {
			store.now = now
		}
	}
}

func (s *RunStore) Start(ctx context.Context, input StartInput) (Run, error) {
	if input.TriggerType == "" {
		return Run{}, fmt.Errorf("%w: trigger type is required", ErrValidation)
	}
	id, err := newID("run")
	if err != nil {
		return Run{}, err
	}
	now := s.now().UTC()
	inputEnvelope := map[string]any{
		"engine":        input.Engine,
		"actor_user_id": input.ActorUserID,
		"input":         nonNilMap(input.Input),
		"limits":        nonNilMap(input.Limits),
	}
	inputJSON, err := marshalObject(inputEnvelope)
	if err != nil {
		return Run{}, err
	}
	outputJSON, err := marshalObject(map[string]any{})
	if err != nil {
		return Run{}, err
	}

	run := Run{
		ID:          id,
		TriggerType: input.TriggerType,
		TriggerRef:  input.TriggerRef,
		ProjectID:   input.ProjectID,
		TicketID:    input.TicketID,
		Status:      StatusRunning,
		Input:       inputEnvelope,
		Output:      map[string]any{},
		StartedAt:   &now,
		CreatedAt:   now,
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO automation_runs (
			id, trigger_type, trigger_ref, project_id, ticket_id, status,
			input_json, output_json, started_at, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, run.ID, run.TriggerType, nullableString(run.TriggerRef), nullableString(run.ProjectID), nullableString(run.TicketID), run.Status, inputJSON, outputJSON, formatTime(now), formatTime(now)); err != nil {
		return Run{}, fmt.Errorf("insert automation run: %w", err)
	}
	return run, nil
}

func (s *RunStore) Finish(ctx context.Context, runID string, input FinishInput) (Run, error) {
	if runID == "" {
		return Run{}, fmt.Errorf("%w: run id is required", ErrValidation)
	}
	if !validFinalStatus(input.Status) {
		return Run{}, fmt.Errorf("%w: invalid final status", ErrValidation)
	}
	outputEnvelope := map[string]any{
		"output": nonNilMap(input.Output),
		"logs":   input.Logs,
	}
	outputJSON, err := marshalObject(outputEnvelope)
	if err != nil {
		return Run{}, err
	}
	finishedAt := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE automation_runs
		SET status = ?, output_json = ?, error = ?, finished_at = ?
		WHERE id = ?
	`, input.Status, outputJSON, nullableString(input.Error), formatTime(finishedAt), runID)
	if err != nil {
		return Run{}, fmt.Errorf("finish automation run: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Run{}, fmt.Errorf("check automation run finish: %w", err)
	}
	if affected == 0 {
		return Run{}, ErrNotFound
	}
	return s.Get(ctx, runID)
}

func (s *RunStore) Get(ctx context.Context, runID string) (Run, error) {
	run, err := scanRun(s.db.QueryRowContext(ctx, `
		SELECT id, trigger_type, trigger_ref, project_id, ticket_id, status,
			input_json, output_json, COALESCE(error, ''), started_at, finished_at, created_at
		FROM automation_runs
		WHERE id = ?
	`, runID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Run{}, ErrNotFound
		}
		return Run{}, fmt.Errorf("get automation run: %w", err)
	}
	return run, nil
}

func (s *RunStore) List(ctx context.Context, input ListInput) ([]Run, error) {
	limit, offset, err := normalizeList(input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}
	where := []string{"1 = 1"}
	args := []any{}
	if input.TriggerType != "" {
		where = append(where, "trigger_type = ?")
		args = append(args, input.TriggerType)
	}
	if input.TriggerRef != "" {
		where = append(where, "trigger_ref = ?")
		args = append(args, input.TriggerRef)
	}
	if input.ProjectID != "" {
		where = append(where, "project_id = ?")
		args = append(args, input.ProjectID)
	}
	if input.TicketID != "" {
		where = append(where, "ticket_id = ?")
		args = append(args, input.TicketID)
	}
	if input.Status != "" {
		where = append(where, "status = ?")
		args = append(args, input.Status)
	}
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, trigger_type, trigger_ref, project_id, ticket_id, status,
			input_json, output_json, COALESCE(error, ''), started_at, finished_at, created_at
		FROM automation_runs
		WHERE `+joinAnd(where)+`
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("list automation runs: %w", err)
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate automation runs: %w", err)
	}
	return runs, nil
}

func scanRun(scanner interface{ Scan(...any) error }) (Run, error) {
	var run Run
	var triggerRef sql.NullString
	var projectID sql.NullString
	var ticketID sql.NullString
	var inputJSON string
	var outputJSON string
	var startedAt sql.NullString
	var finishedAt sql.NullString
	var createdAt string
	if err := scanner.Scan(
		&run.ID,
		&run.TriggerType,
		&triggerRef,
		&projectID,
		&ticketID,
		&run.Status,
		&inputJSON,
		&outputJSON,
		&run.Error,
		&startedAt,
		&finishedAt,
		&createdAt,
	); err != nil {
		return Run{}, err
	}
	run.TriggerRef = nullString(triggerRef)
	run.ProjectID = nullString(projectID)
	run.TicketID = nullString(ticketID)
	if err := json.Unmarshal([]byte(defaultJSON(inputJSON)), &run.Input); err != nil {
		return Run{}, fmt.Errorf("decode automation run input: %w", err)
	}
	if err := json.Unmarshal([]byte(defaultJSON(outputJSON)), &run.Output); err != nil {
		return Run{}, fmt.Errorf("decode automation run output: %w", err)
	}
	run.StartedAt = parseNullableTime(startedAt)
	run.FinishedAt = parseNullableTime(finishedAt)
	created, err := parseTime(createdAt)
	if err != nil {
		return Run{}, err
	}
	run.CreatedAt = created
	return run, nil
}

func marshalObject(value map[string]any) (string, error) {
	encoded, err := json.Marshal(nonNilMap(value))
	if err != nil {
		return "", fmt.Errorf("encode automation run JSON: %w", err)
	}
	return string(encoded), nil
}

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func validFinalStatus(status string) bool {
	switch status {
	case StatusSucceeded, StatusFailed, StatusCanceled:
		return true
	default:
		return false
	}
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

func newID(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate automation run id: %w", err)
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(raw[:]), nil
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

func defaultJSON(value string) string {
	if value == "" {
		return "{}"
	}
	return value
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse automation run time: %w", err)
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
