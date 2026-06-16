package attachments

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
	MaxAttachmentSizeBytes = 10 << 20

	activityAttachmentUploaded = "attachment.uploaded"
	activityAttachmentDeleted  = "attachment.deleted"
)

var (
	ErrNotFound   = errors.New("attachments: not found")
	ErrValidation = errors.New("attachments: validation failed")
	ErrTooLarge   = errors.New("attachments: too large")
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

type Metadata struct {
	ID          string    `json:"id"`
	TicketID    string    `json:"ticket_id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	UploaderID  string    `json:"uploader_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type File struct {
	Metadata
	Data []byte
}

type UploadInput struct {
	TicketID    string
	FileName    string
	ContentType string
	Data        []byte
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

func (s *Service) Upload(ctx context.Context, principal authz.Principal, input UploadInput) (Metadata, error) {
	fields := map[string]string{}
	input.TicketID = strings.TrimSpace(input.TicketID)
	if input.TicketID == "" {
		fields["ticket_id"] = "Required"
	}
	input.FileName = strings.TrimSpace(input.FileName)
	if input.FileName == "" {
		fields["file_name"] = "Required"
	}
	if len(input.FileName) > 240 {
		fields["file_name"] = "Must be 240 characters or fewer"
	}
	if len(input.Data) > MaxAttachmentSizeBytes {
		return Metadata{}, ErrTooLarge
	}
	if len(fields) > 0 {
		return Metadata{}, &ValidationError{Message: "Invalid attachment", Fields: fields}
	}
	if input.ContentType = strings.TrimSpace(input.ContentType); input.ContentType == "" {
		input.ContentType = "application/octet-stream"
	}

	projectID, err := s.ticketProject(ctx, input.TicketID)
	if err != nil {
		return Metadata{}, err
	}
	if err := s.require(principal, authz.PermissionAttachmentsWrite, authz.ProjectScope(projectID)); err != nil {
		return Metadata{}, err
	}

	id, err := newID("attachment")
	if err != nil {
		return Metadata{}, err
	}
	createdAt := s.now().UTC()
	meta := Metadata{
		ID:          id,
		TicketID:    input.TicketID,
		FileName:    input.FileName,
		ContentType: input.ContentType,
		SizeBytes:   int64(len(input.Data)),
		UploaderID:  actorID(principal),
		CreatedAt:   createdAt,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Metadata{}, fmt.Errorf("begin upload attachment: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO ticket_attachments (
			id, ticket_id, uploader_id, file_name, content_type, size_bytes, data, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, meta.ID, meta.TicketID, nullableString(meta.UploaderID), meta.FileName, meta.ContentType, meta.SizeBytes, input.Data, formatTime(meta.CreatedAt)); err != nil {
		return Metadata{}, fmt.Errorf("insert attachment: %w", err)
	}
	if err := insertActivity(ctx, tx, meta.TicketID, meta.UploaderID, activityAttachmentUploaded, map[string]any{
		"attachment_id": meta.ID,
		"file_name":     meta.FileName,
		"size_bytes":    meta.SizeBytes,
	}, meta.CreatedAt); err != nil {
		return Metadata{}, err
	}
	if err := tx.Commit(); err != nil {
		return Metadata{}, fmt.Errorf("commit upload attachment: %w", err)
	}

	s.publish(ctx, events.Event{
		Type:      activityAttachmentUploaded,
		ActorID:   meta.UploaderID,
		ProjectID: projectID,
		ObjectID:  meta.ID,
		At:        meta.CreatedAt,
		Data: map[string]any{
			"ticket_id":  meta.TicketID,
			"file_name":  meta.FileName,
			"size_bytes": meta.SizeBytes,
		},
	})
	return meta, nil
}

func (s *Service) List(ctx context.Context, principal authz.Principal, ticketID string) ([]Metadata, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return nil, &ValidationError{Message: "Invalid attachment list", Fields: map[string]string{"ticket_id": "Required"}}
	}
	projectID, err := s.ticketProject(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, ticket_id, file_name, content_type, size_bytes, uploader_id, created_at
		FROM ticket_attachments
		WHERE ticket_id = ? AND deleted_at IS NULL
		ORDER BY created_at ASC, id ASC
	`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	defer rows.Close()

	var attachments []Metadata
	for rows.Next() {
		meta, err := scanMetadata(rows)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, meta)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attachments: %w", err)
	}
	return attachments, nil
}

func (s *Service) Download(ctx context.Context, principal authz.Principal, attachmentID string) (File, error) {
	file, err := s.file(ctx, attachmentID)
	if err != nil {
		return File{}, err
	}
	projectID, err := s.ticketProject(ctx, file.TicketID)
	if err != nil {
		return File{}, err
	}
	if err := s.require(principal, authz.PermissionTicketsRead, authz.ProjectScope(projectID)); err != nil {
		return File{}, err
	}
	return file, nil
}

func (s *Service) Delete(ctx context.Context, principal authz.Principal, attachmentID string) error {
	file, err := s.file(ctx, attachmentID)
	if err != nil {
		return err
	}
	projectID, err := s.ticketProject(ctx, file.TicketID)
	if err != nil {
		return err
	}
	if err := s.require(principal, authz.PermissionAttachmentsWrite, authz.ProjectScope(projectID)); err != nil {
		return err
	}

	deletedAt := s.now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete attachment: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE ticket_attachments
		SET deleted_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatTime(deletedAt), attachmentID)
	if err != nil {
		return fmt.Errorf("delete attachment: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted attachment: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	if err := insertActivity(ctx, tx, file.TicketID, actorID(principal), activityAttachmentDeleted, map[string]any{
		"attachment_id": file.ID,
		"file_name":     file.FileName,
	}, deletedAt); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete attachment: %w", err)
	}

	s.publish(ctx, events.Event{
		Type:      activityAttachmentDeleted,
		ActorID:   actorID(principal),
		ProjectID: projectID,
		ObjectID:  file.ID,
		At:        deletedAt,
		Data: map[string]any{
			"ticket_id": file.TicketID,
			"file_name": file.FileName,
		},
	})
	return nil
}

func (s *Service) file(ctx context.Context, attachmentID string) (File, error) {
	attachmentID = strings.TrimSpace(attachmentID)
	if attachmentID == "" {
		return File{}, &ValidationError{Message: "Invalid attachment", Fields: map[string]string{"attachment_id": "Required"}}
	}
	var file File
	var uploaderID sql.NullString
	var createdAt string
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, ticket_id, file_name, content_type, size_bytes, uploader_id, created_at, data
		FROM ticket_attachments
		WHERE id = ? AND deleted_at IS NULL
	`, attachmentID).Scan(&file.ID, &file.TicketID, &file.FileName, &file.ContentType, &file.SizeBytes, &uploaderID, &createdAt, &file.Data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return File{}, ErrNotFound
		}
		return File{}, fmt.Errorf("get attachment: %w", err)
	}
	file.UploaderID = nullString(uploaderID)
	parsed, err := parseTime(createdAt)
	if err != nil {
		return File{}, err
	}
	file.CreatedAt = parsed
	return file, nil
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
		return "", fmt.Errorf("query attachment ticket project: %w", err)
	}
	return projectID, nil
}

func (s *Service) require(principal authz.Principal, permission authz.Permission, scope authz.Scope) error {
	if s == nil || s.authorizer == nil {
		return errors.New("attachments: authorization evaluator is required")
	}
	return s.authorizer.Require(principal, permission, scope)
}

func (s *Service) publish(ctx context.Context, event events.Event) {
	if s == nil || s.eventBus == nil {
		return
	}
	_ = s.eventBus.Publish(ctx, event)
}

func scanMetadata(scanner interface{ Scan(...any) error }) (Metadata, error) {
	var meta Metadata
	var uploaderID sql.NullString
	var createdAt string
	if err := scanner.Scan(&meta.ID, &meta.TicketID, &meta.FileName, &meta.ContentType, &meta.SizeBytes, &uploaderID, &createdAt); err != nil {
		return Metadata{}, fmt.Errorf("scan attachment metadata: %w", err)
	}
	meta.UploaderID = nullString(uploaderID)
	parsed, err := parseTime(createdAt)
	if err != nil {
		return Metadata{}, err
	}
	meta.CreatedAt = parsed
	return meta, nil
}

func insertActivity(ctx context.Context, tx *sql.Tx, ticketID string, actorID string, activityType string, data map[string]any, at time.Time) error {
	id, err := newID("activity")
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode attachment activity: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO ticket_activity (id, ticket_id, actor_id, activity_type, data_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, ticketID, nullableString(actorID), activityType, string(encoded), formatTime(at)); err != nil {
		return fmt.Errorf("insert attachment activity: %w", err)
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
		return time.Time{}, fmt.Errorf("parse attachment time: %w", err)
	}
	return parsed, nil
}
