package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/attachments"
)

const GlobalSettingsKey = "global"

const (
	defaultAttachmentMaxSizeBytes = 10 << 20
	maxAttachmentMaxSizeBytes     = 100 << 20
)

var (
	ErrValidation = errors.New("settings: validation failed")
	ErrNotFound   = errors.New("settings: not found")
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

type GlobalSettings struct {
	AttachmentMaxSizeBytes        int64     `json:"attachment_max_size_bytes"`
	AttachmentAllowedContentTypes []string  `json:"attachment_allowed_content_types"`
	WebhookAllowedBaseURLs        []string  `json:"webhook_allowed_base_urls"`
	DemoWarningEnabled            bool      `json:"demo_warning_enabled"`
	BackupEnabled                 bool      `json:"backup_enabled"`
	SystemHealthNote              string    `json:"system_health_note,omitempty"`
	UpdatedBy                     string    `json:"updated_by,omitempty"`
	UpdatedAt                     time.Time `json:"updated_at"`
}

type UpdateGlobalInput struct {
	AttachmentMaxSizeBytes        *int64
	AttachmentAllowedContentTypes *[]string
	WebhookAllowedBaseURLs        *[]string
	DemoWarningEnabled            *bool
	BackupEnabled                 *bool
	SystemHealthNote              *string
	UpdatedBy                     string
}

type Service struct {
	db  *sql.DB
	now func() time.Time
}

type Option func(*Service)

func NewService(db *sql.DB, options ...Option) *Service {
	service := &Service{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
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

func (s *Service) GetGlobal(ctx context.Context) (GlobalSettings, error) {
	settings, err := s.getGlobal(ctx)
	if err != nil {
		return GlobalSettings{}, err
	}
	return settings, nil
}

func (s *Service) UpdateGlobal(ctx context.Context, input UpdateGlobalInput) (GlobalSettings, error) {
	current, err := s.getGlobal(ctx)
	if err != nil {
		return GlobalSettings{}, err
	}
	next := current
	if input.AttachmentMaxSizeBytes != nil {
		next.AttachmentMaxSizeBytes = *input.AttachmentMaxSizeBytes
	}
	if input.AttachmentAllowedContentTypes != nil {
		next.AttachmentAllowedContentTypes = cloneStrings(*input.AttachmentAllowedContentTypes)
	}
	if input.WebhookAllowedBaseURLs != nil {
		next.WebhookAllowedBaseURLs = cloneStrings(*input.WebhookAllowedBaseURLs)
	}
	if input.DemoWarningEnabled != nil {
		next.DemoWarningEnabled = *input.DemoWarningEnabled
	}
	if input.BackupEnabled != nil {
		next.BackupEnabled = *input.BackupEnabled
	}
	if input.SystemHealthNote != nil {
		next.SystemHealthNote = strings.TrimSpace(*input.SystemHealthNote)
	}
	next.UpdatedBy = strings.TrimSpace(input.UpdatedBy)
	next.UpdatedAt = s.now().UTC()
	if err := validateGlobal(next); err != nil {
		return GlobalSettings{}, err
	}
	next.AttachmentAllowedContentTypes = normalizeContentTypes(next.AttachmentAllowedContentTypes)
	next.WebhookAllowedBaseURLs = normalizeBaseURLs(next.WebhookAllowedBaseURLs)
	if err := s.writeGlobal(ctx, next); err != nil {
		return GlobalSettings{}, err
	}
	return next, nil
}

func (s *Service) AttachmentPolicy(ctx context.Context) (attachments.AttachmentPolicy, error) {
	global, err := s.GetGlobal(ctx)
	if err != nil {
		return attachments.AttachmentPolicy{}, err
	}
	return attachments.AttachmentPolicy{
		MaxSizeBytes:        global.AttachmentMaxSizeBytes,
		AllowedContentTypes: cloneStrings(global.AttachmentAllowedContentTypes),
	}, nil
}

func (s *Service) OutgoingWebhookBaseURLs(ctx context.Context) ([]string, error) {
	global, err := s.GetGlobal(ctx)
	if err != nil {
		return nil, err
	}
	return cloneStrings(global.WebhookAllowedBaseURLs), nil
}

func (s *Service) getGlobal(ctx context.Context) (GlobalSettings, error) {
	var valueJSON string
	var updatedBy sql.NullString
	var updatedAt string
	err := s.db.QueryRowContext(ctx, `
		SELECT value_json, updated_by, updated_at
		FROM system_settings
		WHERE key = ?
	`, GlobalSettingsKey).Scan(&valueJSON, &updatedBy, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			settings := defaultGlobalSettings(s.now().UTC())
			if seedErr := s.writeGlobal(ctx, settings); seedErr != nil {
				return GlobalSettings{}, seedErr
			}
			return s.getGlobal(ctx)
		}
		return GlobalSettings{}, fmt.Errorf("get global settings: %w", err)
	}
	settings, err := decodeGlobal(valueJSON)
	if err != nil {
		return GlobalSettings{}, err
	}
	settings.UpdatedBy = nullString(updatedBy)
	settings.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return GlobalSettings{}, fmt.Errorf("parse global settings updated_at: %w", err)
	}
	return settings, nil
}

func (s *Service) writeGlobal(ctx context.Context, settings GlobalSettings) error {
	valueJSON, err := encodeGlobal(settings)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO system_settings (key, value_json, updated_by, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value_json = excluded.value_json,
			updated_by = excluded.updated_by,
			updated_at = excluded.updated_at
	`, GlobalSettingsKey, valueJSON, nullableString(settings.UpdatedBy), formatTime(settings.UpdatedAt)); err != nil {
		return fmt.Errorf("write global settings: %w", err)
	}
	return nil
}

func defaultGlobalSettings(updatedAt time.Time) GlobalSettings {
	return GlobalSettings{
		AttachmentMaxSizeBytes:        defaultAttachmentMaxSizeBytes,
		AttachmentAllowedContentTypes: []string{},
		WebhookAllowedBaseURLs:        []string{},
		DemoWarningEnabled:            true,
		BackupEnabled:                 false,
		SystemHealthNote:              "",
		UpdatedAt:                     updatedAt,
	}
}

func validateGlobal(input GlobalSettings) error {
	fields := map[string]string{}
	if input.AttachmentMaxSizeBytes <= 0 {
		fields["attachment_max_size_bytes"] = "Must be greater than zero"
	} else if input.AttachmentMaxSizeBytes > maxAttachmentMaxSizeBytes {
		fields["attachment_max_size_bytes"] = "Must be 100 MiB or less"
	}
	for _, contentType := range input.AttachmentAllowedContentTypes {
		contentType = strings.ToLower(strings.TrimSpace(contentType))
		if contentType == "" {
			continue
		}
		if !validContentType(contentType) {
			fields["attachment_allowed_content_types"] = "Must contain MIME content types such as text/plain"
			break
		}
	}
	for _, rawURL := range input.WebhookAllowedBaseURLs {
		rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
		if rawURL == "" {
			continue
		}
		parsed, err := url.Parse(rawURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			fields["webhook_allowed_base_urls"] = "Must contain absolute http or https URLs without credentials, query, or fragment"
			break
		}
	}
	if len(input.SystemHealthNote) > 2000 {
		fields["system_health_note"] = "Must be 2000 characters or fewer"
	}
	if len(fields) > 0 {
		return &ValidationError{Message: "Invalid global settings", Fields: fields}
	}
	return nil
}

func encodeGlobal(settings GlobalSettings) (string, error) {
	settings.AttachmentAllowedContentTypes = normalizeContentTypes(settings.AttachmentAllowedContentTypes)
	settings.WebhookAllowedBaseURLs = normalizeBaseURLs(settings.WebhookAllowedBaseURLs)
	payload := map[string]any{
		"attachment_max_size_bytes":        settings.AttachmentMaxSizeBytes,
		"attachment_allowed_content_types": settings.AttachmentAllowedContentTypes,
		"webhook_allowed_base_urls":        settings.WebhookAllowedBaseURLs,
		"demo_warning_enabled":             settings.DemoWarningEnabled,
		"backup_enabled":                   settings.BackupEnabled,
		"system_health_note":               settings.SystemHealthNote,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode global settings: %w", err)
	}
	return string(encoded), nil
}

func decodeGlobal(valueJSON string) (GlobalSettings, error) {
	settings := defaultGlobalSettings(time.Time{})
	if strings.TrimSpace(valueJSON) == "" {
		return settings, nil
	}
	var payload struct {
		AttachmentMaxSizeBytes        int64    `json:"attachment_max_size_bytes"`
		AttachmentAllowedContentTypes []string `json:"attachment_allowed_content_types"`
		WebhookAllowedBaseURLs        []string `json:"webhook_allowed_base_urls"`
		DemoWarningEnabled            *bool    `json:"demo_warning_enabled"`
		BackupEnabled                 *bool    `json:"backup_enabled"`
		SystemHealthNote              string   `json:"system_health_note"`
	}
	if err := json.Unmarshal([]byte(valueJSON), &payload); err != nil {
		return GlobalSettings{}, fmt.Errorf("decode global settings: %w", err)
	}
	if payload.AttachmentMaxSizeBytes > 0 {
		settings.AttachmentMaxSizeBytes = payload.AttachmentMaxSizeBytes
	}
	settings.AttachmentAllowedContentTypes = normalizeContentTypes(payload.AttachmentAllowedContentTypes)
	settings.WebhookAllowedBaseURLs = normalizeBaseURLs(payload.WebhookAllowedBaseURLs)
	if payload.DemoWarningEnabled != nil {
		settings.DemoWarningEnabled = *payload.DemoWarningEnabled
	}
	if payload.BackupEnabled != nil {
		settings.BackupEnabled = *payload.BackupEnabled
	}
	settings.SystemHealthNote = strings.TrimSpace(payload.SystemHealthNote)
	return settings, nil
}

func normalizeContentTypes(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" || slices.Contains(normalized, value) {
			continue
		}
		normalized = append(normalized, value)
	}
	slices.Sort(normalized)
	return normalized
}

func normalizeBaseURLs(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimRight(strings.TrimSpace(value), "/")
		if value == "" || slices.Contains(normalized, value) {
			continue
		}
		normalized = append(normalized, value)
	}
	slices.Sort(normalized)
	return normalized
}

func validContentType(value string) bool {
	left, right, ok := strings.Cut(value, "/")
	return ok && left != "" && right != "" && !strings.ContainsAny(value, " \t\r\n")
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
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
