package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

const SessionCookieName = "rayboard_session"

var (
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrUnauthenticated    = errors.New("auth: unauthenticated")
	ErrDisabledUser       = errors.New("auth: disabled user")
)

type Service struct {
	db         *sql.DB
	now        func() time.Time
	sessionTTL time.Duration
}

type Option func(*Service)

func NewService(db *sql.DB, options ...Option) *Service {
	service := &Service{
		db:         db,
		now:        func() time.Time { return time.Now().UTC() },
		sessionTTL: 24 * time.Hour,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithNow(now func() time.Time) Option {
	return func(service *Service) {
		service.now = now
	}
}

func WithSessionTTL(ttl time.Duration) Option {
	return func(service *Service) {
		service.sessionTTL = ttl
	}
}

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Disabled    bool   `json:"disabled"`
}

type Session struct {
	ID        string
	User      User
	Secret    string
	ExpiresAt time.Time
}

type APIToken struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type CreatedAPIToken struct {
	APIToken
	Token string `json:"token"`
}

func (s *Service) Login(ctx context.Context, username string, password string) (Session, error) {
	var user User
	var passwordHash string
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, username, display_name, password_hash, is_disabled
		FROM users
		WHERE username = ? AND deleted_at IS NULL
	`, username).Scan(&user.ID, &user.Username, &user.DisplayName, &passwordHash, &user.Disabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrInvalidCredentials
		}
		return Session{}, fmt.Errorf("query login user: %w", err)
	}
	if user.Disabled {
		return Session{}, ErrDisabledUser
	}
	if passwordHash == "" || !VerifyPassword(passwordHash, password) {
		return Session{}, ErrInvalidCredentials
	}

	secret, err := randomSecret()
	if err != nil {
		return Session{}, err
	}
	session := Session{
		ID:        newID("session"),
		User:      user,
		Secret:    secret,
		ExpiresAt: s.now().Add(s.sessionTTL),
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, token_hash, expires_at)
		VALUES (?, ?, ?, ?)
	`, session.ID, user.ID, hashSecret(secret), formatTime(session.ExpiresAt)); err != nil {
		return Session{}, fmt.Errorf("insert session: %w", err)
	}
	return session, nil
}

func (s *Service) AuthenticateSession(ctx context.Context, secret string) (authz.Principal, User, error) {
	return s.authenticateSecret(ctx, `
		SELECT u.id, u.username, u.display_name, u.is_disabled
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = ?
		  AND s.revoked_at IS NULL
		  AND s.expires_at > ?
		  AND u.deleted_at IS NULL
	`, secret, authz.AuthKindSession, func(ctx context.Context, userID string) error {
		_, err := s.db.ExecContext(ctx, `
			UPDATE sessions
			SET last_seen_at = ?
			WHERE token_hash = ?
		`, formatTime(s.now()), hashSecret(secret))
		return err
	})
}

func (s *Service) AuthenticateBearer(ctx context.Context, token string) (authz.Principal, User, error) {
	return s.authenticateSecret(ctx, `
		SELECT u.id, u.username, u.display_name, u.is_disabled
		FROM api_tokens t
		JOIN users u ON u.id = t.user_id
		WHERE t.token_hash = ?
		  AND t.revoked_at IS NULL
		  AND (t.expires_at IS NULL OR t.expires_at > ?)
		  AND u.deleted_at IS NULL
	`, token, authz.AuthKindAPIToken, func(ctx context.Context, userID string) error {
		_, err := s.db.ExecContext(ctx, `
			UPDATE api_tokens
			SET last_used_at = ?
			WHERE token_hash = ?
		`, formatTime(s.now()), hashSecret(token))
		return err
	})
}

func (s *Service) authenticateSecret(ctx context.Context, query string, secret string, authKind authz.AuthKind, touch func(context.Context, string) error) (authz.Principal, User, error) {
	if secret == "" {
		return authz.Principal{}, User{}, ErrUnauthenticated
	}

	var user User
	if err := s.db.QueryRowContext(ctx, query, hashSecret(secret), formatTime(s.now())).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Disabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return authz.Principal{}, User{}, ErrUnauthenticated
		}
		return authz.Principal{}, User{}, fmt.Errorf("authenticate secret: %w", err)
	}
	if user.Disabled {
		return authz.Principal{}, User{}, ErrDisabledUser
	}
	if err := touch(ctx, user.ID); err != nil {
		return authz.Principal{}, User{}, fmt.Errorf("touch credential: %w", err)
	}

	return authz.Principal{
		UserID:      user.ID,
		AuthKind:    authKind,
		ActorUserID: user.ID,
	}, user, nil
}

func (s *Service) Logout(ctx context.Context, secret string) error {
	if secret == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE sessions
		SET revoked_at = ?
		WHERE token_hash = ? AND revoked_at IS NULL
	`, formatTime(s.now()), hashSecret(secret))
	return err
}

func (s *Service) CreateAPIToken(ctx context.Context, userID string, name string) (CreatedAPIToken, error) {
	if name == "" {
		name = "API token"
	}
	secret, err := randomSecret()
	if err != nil {
		return CreatedAPIToken{}, err
	}

	created := s.now()
	token := CreatedAPIToken{
		APIToken: APIToken{
			ID:        newID("token"),
			Name:      name,
			CreatedAt: created,
		},
		Token: secret,
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO api_tokens (id, user_id, name, token_hash, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, token.ID, userID, token.Name, hashSecret(secret), formatTime(created)); err != nil {
		return CreatedAPIToken{}, fmt.Errorf("insert API token: %w", err)
	}
	return token, nil
}

func (s *Service) ListAPITokens(ctx context.Context, userID string) ([]APIToken, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, created_at, last_used_at, expires_at, revoked_at
		FROM api_tokens
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list API tokens: %w", err)
	}
	defer rows.Close()

	var tokens []APIToken
	for rows.Next() {
		var token APIToken
		var created string
		var lastUsed sql.NullString
		var expires sql.NullString
		var revoked sql.NullString
		if err := rows.Scan(&token.ID, &token.Name, &created, &lastUsed, &expires, &revoked); err != nil {
			return nil, fmt.Errorf("scan API token: %w", err)
		}
		token.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		token.LastUsedAt = parseNullableTime(lastUsed)
		token.ExpiresAt = parseNullableTime(expires)
		token.RevokedAt = parseNullableTime(revoked)
		tokens = append(tokens, token)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate API tokens: %w", err)
	}
	return tokens, nil
}

func (s *Service) RevokeAPIToken(ctx context.Context, userID string, tokenID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE api_tokens
		SET revoked_at = ?
		WHERE id = ? AND user_id = ? AND revoked_at IS NULL
	`, formatTime(s.now()), tokenID, userID)
	return err
}

func randomSecret() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate secret: %w", err)
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

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return nil
	}
	return &parsed
}
