package comments

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
	"github.com/timo-42/rayboard/internal/backend/events"
)

const (
	activityCommentCreated = "comment.created"
	activityCommentDeleted = "comment.deleted"
)

var (
	ErrNotFound   = errors.New("comments: not found")
	ErrValidation = errors.New("comments: validation failed")
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

type Comment struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	AuthorID  string    `json:"author_id,omitempty"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	TicketID string
	Body     string
}

type Service struct {
	db         *sql.DB
	authorizer authz.Evaluator
	now        func() time.Time
	eventBus   *events.Bus
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

func WithEventBus(bus *events.Bus) Option {
	return func(service *Service) {
		service.eventBus = bus
	}
}

func (s *Service) Create(ctx context.Context, principal authz.Principal, input CreateInput) (Comment, error) {
	input.TicketID = strings.TrimSpace(input.TicketID)
	body := strings.TrimSpace(input.Body)
	fields := map[string]string{}
	if input.TicketID == "" {
		fields["ticket_id"] = "Required"
	}
	if body == "" {
		fields["body"] = "Required"
	}
	if len(body) > 20000 {
		fields["body"] = "Must be 20000 characters or fewer"
	}
	if len(fields) > 0 {
		return Comment{}, &ValidationError{Message: "Invalid comment", Fields: fields}
	}

	projectID, err := s.ticketProject(ctx, input.TicketID)
	if err != nil {
		return Comment{}, err
	}
	if err := s.require(principal, authz.PermissionCommentsWrite, authz.ProjectScope(projectID)); err != nil {
		return Comment{}, err
	}

	id, err := newID("comment")
	if err != nil {
		return Comment{}, err
	}
	now := s.now().UTC()
	comment := Comment{
		ID:        id,
		TicketID:  input.TicketID,
		AuthorID:  actorID(principal),
		Body:      body,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Comment{}, fmt.Errorf("begin create comment: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO ticket_comments (id, ticket_id, author_id, body, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, comment.ID, comment.TicketID, nullableString(comment.AuthorID), comment.Body, formatTime(comment.CreatedAt), formatTime(comment.UpdatedAt)); err != nil {
		return Comment{}, fmt.Errorf("insert comment: %w", err)
	}
	if err := insertActivity(ctx, tx, comment.TicketID, comment.AuthorID, activityCommentCreated, map[string]any{
		"comment_id": comment.ID,
	}, comment.CreatedAt); err != nil {
		return Comment{}, err
	}
	if err := tx.Commit(); err != nil {
		return Comment{}, fmt.Errorf("commit create comment: %w", err)
	}

	s.publish(ctx, events.Event{
		Type:      activityCommentCreated,
		ActorID:   comment.AuthorID,
		ProjectID: projectID,
		ObjectID:  comment.ID,
		At:        comment.CreatedAt,
		Data: map[string]any{
			"ticket_id": comment.TicketID,
		},
	})
	return comment, nil
}

func (s *Service) List(ctx context.Context, principal authz.Principal, ticketID string) ([]Comment, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return nil, &ValidationError{Message: "Invalid comment list", Fields: map[string]string{"ticket_id": "Required"}}
	}
	projectID, err := s.ticketProject(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, ticket_id, author_id, body, created_at, updated_at
		FROM ticket_comments
		WHERE ticket_id = ? AND deleted_at IS NULL
		ORDER BY created_at ASC, id ASC
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		comment, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}
	return comments, nil
}

func (s *Service) Delete(ctx context.Context, principal authz.Principal, commentID string) error {
	comment, projectID, err := s.getComment(ctx, commentID)
	if err != nil {
		return err
	}
	if err := s.require(principal, authz.PermissionCommentsWrite, authz.ProjectScope(projectID)); err != nil {
		return err
	}

	deletedAt := s.now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete comment: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE ticket_comments
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatTime(deletedAt), formatTime(deletedAt), comment.ID)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted comment: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	if err := insertActivity(ctx, tx, comment.TicketID, actorID(principal), activityCommentDeleted, map[string]any{
		"comment_id": comment.ID,
	}, deletedAt); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete comment: %w", err)
	}

	s.publish(ctx, events.Event{
		Type:      activityCommentDeleted,
		ActorID:   actorID(principal),
		ProjectID: projectID,
		ObjectID:  comment.ID,
		At:        deletedAt,
		Data: map[string]any{
			"ticket_id": comment.TicketID,
		},
	})
	return nil
}

func (s *Service) getComment(ctx context.Context, commentID string) (Comment, string, error) {
	commentID = strings.TrimSpace(commentID)
	if commentID == "" {
		return Comment{}, "", &ValidationError{Message: "Invalid comment", Fields: map[string]string{"comment_id": "Required"}}
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT c.id, c.ticket_id, c.author_id, c.body, c.created_at, c.updated_at, t.project_id
		FROM ticket_comments c
		JOIN tickets t ON t.id = c.ticket_id
		WHERE c.id = ? AND c.deleted_at IS NULL AND t.deleted_at IS NULL
	`, commentID)
	var projectID string
	comment, err := scanCommentWithProject(row, &projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Comment{}, "", ErrNotFound
		}
		return Comment{}, "", fmt.Errorf("get comment: %w", err)
	}
	return comment, projectID, nil
}

func (s *Service) ticketProject(ctx context.Context, ticketID string) (string, error) {
	var projectID string
	if err := s.db.QueryRowContext(ctx, `
		SELECT project_id
		FROM tickets
		WHERE id = ? AND deleted_at IS NULL
	`, ticketID).Scan(&projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("query comment ticket project: %w", err)
	}
	return projectID, nil
}

func (s *Service) require(principal authz.Principal, permission authz.Permission, scope authz.Scope) error {
	if s == nil || s.authorizer == nil {
		return errors.New("comments: authorization evaluator is required")
	}
	return s.authorizer.Require(principal, permission, scope)
}

func (s *Service) publish(ctx context.Context, event events.Event) {
	if s == nil || s.eventBus == nil {
		return
	}
	_ = s.eventBus.Publish(ctx, event)
}

func scanComment(scanner interface{ Scan(...any) error }) (Comment, error) {
	return scanCommentWithProject(scanner, nil)
}

func scanCommentWithProject(scanner interface{ Scan(...any) error }, projectID *string) (Comment, error) {
	var comment Comment
	var authorID sql.NullString
	var createdAt string
	var updatedAt string
	dest := []any{&comment.ID, &comment.TicketID, &authorID, &comment.Body, &createdAt, &updatedAt}
	if projectID != nil {
		dest = append(dest, projectID)
	}
	if err := scanner.Scan(dest...); err != nil {
		return Comment{}, err
	}
	comment.AuthorID = nullString(authorID)
	var err error
	comment.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return Comment{}, err
	}
	comment.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return Comment{}, err
	}
	return comment, nil
}

func insertActivity(ctx context.Context, tx *sql.Tx, ticketID string, actorID string, activityType string, data map[string]any, at time.Time) error {
	id, err := newID("activity")
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode comment activity: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO ticket_activity (id, ticket_id, actor_id, activity_type, data_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, ticketID, nullableString(actorID), activityType, string(encoded), formatTime(at)); err != nil {
		return fmt.Errorf("insert comment activity: %w", err)
	}
	return nil
}

func newID(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func actorID(principal authz.Principal) string {
	if principal.ActorUserID != "" {
		return principal.ActorUserID
	}
	return principal.UserID
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
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

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse comment time: %w", err)
	}
	return parsed, nil
}
