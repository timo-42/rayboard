package openrouter

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound   = errors.New("openrouter: not found")
	ErrValidation = errors.New("openrouter: validation failed")
	ErrConflict   = errors.New("openrouter: conflict")
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

type Provider struct {
	ID                    string
	Name                  string
	DefaultModel          string
	APIKeySet             bool
	AllowedModels         []string
	DefaultTimeoutSeconds int
	MaxOutputTokens       int
	Enabled               bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type CreateProviderInput struct {
	Name                  string
	DefaultModel          string
	APIKey                string
	AllowedModels         []string
	DefaultTimeoutSeconds int
	MaxOutputTokens       int
	Enabled               bool
}

type UpdateProviderInput struct {
	Name                  *string
	DefaultModel          *string
	APIKey                *string
	AllowedModels         *[]string
	DefaultTimeoutSeconds *int
	MaxOutputTokens       *int
	Enabled               *bool
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

func (s *Service) CreateProvider(ctx context.Context, input CreateProviderInput) (Provider, error) {
	provider := Provider{
		ID:                    newID("ai_provider"),
		Name:                  normalizeName(input.Name),
		DefaultModel:          strings.TrimSpace(input.DefaultModel),
		AllowedModels:         normalizeList(input.AllowedModels),
		DefaultTimeoutSeconds: input.DefaultTimeoutSeconds,
		MaxOutputTokens:       input.MaxOutputTokens,
		Enabled:               input.Enabled,
		CreatedAt:             s.now().UTC(),
		UpdatedAt:             s.now().UTC(),
	}
	apiKey := strings.TrimSpace(input.APIKey)
	if provider.DefaultTimeoutSeconds == 0 {
		provider.DefaultTimeoutSeconds = 30
	}
	if provider.MaxOutputTokens == 0 {
		provider.MaxOutputTokens = 2048
	}
	if err := validateProvider(provider, apiKeyRequired(apiKey)); err != nil {
		return Provider{}, err
	}
	allowed, err := json.Marshal(provider.AllowedModels)
	if err != nil {
		return Provider{}, fmt.Errorf("encode allowed models: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO openrouter_providers (
			id, name, default_model, api_key_secret, allowed_models_json,
			default_timeout_seconds, max_output_tokens, enabled, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, provider.ID, provider.Name, provider.DefaultModel, apiKey, string(allowed),
		provider.DefaultTimeoutSeconds, provider.MaxOutputTokens, provider.Enabled,
		formatTime(provider.CreatedAt), formatTime(provider.UpdatedAt)); err != nil {
		if isUniqueConstraint(err) {
			return Provider{}, fmt.Errorf("%w: provider name already exists", ErrConflict)
		}
		return Provider{}, fmt.Errorf("insert OpenRouter provider: %w", err)
	}
	provider.APIKeySet = true
	return provider, nil
}

func (s *Service) ListProviders(ctx context.Context) ([]Provider, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, default_model, api_key_secret, allowed_models_json,
			default_timeout_seconds, max_output_tokens, enabled, created_at, updated_at
		FROM openrouter_providers
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list OpenRouter providers: %w", err)
	}
	defer rows.Close()

	var providers []Provider
	for rows.Next() {
		provider, err := scanProvider(rows)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate OpenRouter providers: %w", err)
	}
	return providers, nil
}

func (s *Service) GetProvider(ctx context.Context, providerID string) (Provider, error) {
	provider, err := s.getProvider(ctx, providerID)
	if err != nil {
		return Provider{}, err
	}
	return provider, nil
}

func (s *Service) UpdateProvider(ctx context.Context, providerID string, input UpdateProviderInput) (Provider, error) {
	current, err := s.getProvider(ctx, providerID)
	if err != nil {
		return Provider{}, err
	}
	apiKey := ""
	if input.APIKey != nil {
		apiKey = strings.TrimSpace(*input.APIKey)
	}
	if input.Name != nil {
		current.Name = normalizeName(*input.Name)
	}
	if input.DefaultModel != nil {
		current.DefaultModel = strings.TrimSpace(*input.DefaultModel)
	}
	if input.AllowedModels != nil {
		current.AllowedModels = normalizeList(*input.AllowedModels)
	}
	if input.DefaultTimeoutSeconds != nil {
		current.DefaultTimeoutSeconds = *input.DefaultTimeoutSeconds
	}
	if input.MaxOutputTokens != nil {
		current.MaxOutputTokens = *input.MaxOutputTokens
	}
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	current.UpdatedAt = s.now().UTC()
	if err := validateProvider(current, true); err != nil {
		return Provider{}, err
	}
	allowed, err := json.Marshal(current.AllowedModels)
	if err != nil {
		return Provider{}, fmt.Errorf("encode allowed models: %w", err)
	}
	query := `
		UPDATE openrouter_providers
		SET name = ?, default_model = ?, allowed_models_json = ?,
			default_timeout_seconds = ?, max_output_tokens = ?, enabled = ?,
			updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	args := []any{
		current.Name, current.DefaultModel, string(allowed),
		current.DefaultTimeoutSeconds, current.MaxOutputTokens, current.Enabled,
		formatTime(current.UpdatedAt), providerID,
	}
	if input.APIKey != nil {
		if apiKey == "" {
			return Provider{}, &ValidationError{Message: "Invalid OpenRouter provider", Fields: map[string]string{"api_key": "Required"}}
		}
		query = `
			UPDATE openrouter_providers
			SET name = ?, default_model = ?, api_key_secret = ?, allowed_models_json = ?,
				default_timeout_seconds = ?, max_output_tokens = ?, enabled = ?,
				updated_at = ?
			WHERE id = ? AND deleted_at IS NULL
		`
		args = []any{
			current.Name, current.DefaultModel, apiKey, string(allowed),
			current.DefaultTimeoutSeconds, current.MaxOutputTokens, current.Enabled,
			formatTime(current.UpdatedAt), providerID,
		}
	}
	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		if isUniqueConstraint(err) {
			return Provider{}, fmt.Errorf("%w: provider name already exists", ErrConflict)
		}
		return Provider{}, fmt.Errorf("update OpenRouter provider: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Provider{}, fmt.Errorf("check OpenRouter provider update: %w", err)
	}
	if affected == 0 {
		return Provider{}, ErrNotFound
	}
	updated, err := s.getProvider(ctx, providerID)
	if err != nil {
		return Provider{}, err
	}
	return updated, nil
}

func (s *Service) DeleteProvider(ctx context.Context, providerID string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE openrouter_providers
		SET deleted_at = ?, enabled = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatTime(s.now().UTC()), formatTime(s.now().UTC()), strings.TrimSpace(providerID))
	if err != nil {
		return fmt.Errorf("delete OpenRouter provider: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check OpenRouter provider delete: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) getProvider(ctx context.Context, providerID string) (Provider, error) {
	var provider Provider
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, default_model, api_key_secret, allowed_models_json,
			default_timeout_seconds, max_output_tokens, enabled, created_at, updated_at
		FROM openrouter_providers
		WHERE id = ? AND deleted_at IS NULL
	`, strings.TrimSpace(providerID))
	provider, err := scanProvider(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Provider{}, ErrNotFound
		}
		return Provider{}, err
	}
	return provider, nil
}

func validateProvider(provider Provider, hasAPIKey bool) error {
	fields := map[string]string{}
	if provider.Name == "" {
		fields["name"] = "Required"
	}
	if provider.DefaultModel == "" {
		fields["default_model"] = "Required"
	}
	if !hasAPIKey {
		fields["api_key"] = "Required"
	}
	if provider.DefaultTimeoutSeconds <= 0 {
		fields["default_timeout_seconds"] = "Must be greater than zero"
	}
	if provider.MaxOutputTokens <= 0 {
		fields["max_output_tokens"] = "Must be greater than zero"
	}
	if len(fields) > 0 {
		return &ValidationError{Message: "Invalid OpenRouter provider", Fields: fields}
	}
	return nil
}

func scanProvider(row interface {
	Scan(dest ...any) error
}) (Provider, error) {
	var provider Provider
	var apiKey string
	var allowedJSON string
	var created string
	var updated string
	if err := row.Scan(
		&provider.ID,
		&provider.Name,
		&provider.DefaultModel,
		&apiKey,
		&allowedJSON,
		&provider.DefaultTimeoutSeconds,
		&provider.MaxOutputTokens,
		&provider.Enabled,
		&created,
		&updated,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Provider{}, err
		}
		return Provider{}, fmt.Errorf("scan OpenRouter provider: %w", err)
	}
	provider.APIKeySet = apiKey != ""
	if allowedJSON == "" {
		allowedJSON = "[]"
	}
	if err := json.Unmarshal([]byte(allowedJSON), &provider.AllowedModels); err != nil {
		return Provider{}, fmt.Errorf("decode allowed models: %w", err)
	}
	provider.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	provider.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	return provider, nil
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func normalizeList(values []string) []string {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func apiKeyRequired(apiKey string) bool {
	return strings.TrimSpace(apiKey) != ""
}

func isUniqueConstraint(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "unique")
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func newID(prefix string) string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return prefix + "_fallback"
	}
	return prefix + "_" + strings.ToLower(strings.TrimRight(base32.StdEncoding.EncodeToString(raw[:]), "="))
}
