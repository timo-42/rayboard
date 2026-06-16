package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	PolicyScopeGlobal  = DestinationScopeGlobal
	PolicyScopeProject = DestinationScopeProject
)

var allowedPolicyEvents = []string{
	"automation_failed",
	"comment_added",
	"release_changed",
	"sprint_changed",
	"ticket_assigned",
	"ticket_status_changed",
}

type Policy struct {
	ID             string
	Name           string
	ScopeType      string
	ProjectID      string
	EventTypes     []string
	DestinationIDs []string
	Enabled        bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ListPoliciesInput struct {
	ScopeType string
	ProjectID string
}

type CreatePolicyInput struct {
	Name           string
	ScopeType      string
	ProjectID      string
	EventTypes     []string
	DestinationIDs []string
	Enabled        bool
}

type UpdatePolicyInput struct {
	Name           *string
	EventTypes     *[]string
	DestinationIDs *[]string
	Enabled        *bool
}

func (s *Service) ListPolicies(ctx context.Context, input ListPoliciesInput) ([]Policy, error) {
	scopeType := strings.TrimSpace(input.ScopeType)
	if scopeType == "" {
		scopeType = PolicyScopeGlobal
	}
	projectID := strings.TrimSpace(input.ProjectID)
	if err := validatePolicyScope(scopeType, projectID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, scope_type, project_id, event_types_json, destination_ids_json, enabled, created_at, updated_at
		FROM notification_policies
		WHERE deleted_at IS NULL AND scope_type = ? AND scope_key = ?
		ORDER BY name ASC
	`, scopeType, policyScopeKey(scopeType, projectID))
	if err != nil {
		return nil, fmt.Errorf("list notification policies: %w", err)
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		policy, err := scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification policies: %w", err)
	}
	return policies, nil
}

func (s *Service) CreatePolicy(ctx context.Context, input CreatePolicyInput) (Policy, error) {
	id, err := newID("policy")
	if err != nil {
		return Policy{}, err
	}
	policy := Policy{
		ID:             id,
		Name:           normalizePolicyName(input.Name),
		ScopeType:      strings.TrimSpace(input.ScopeType),
		ProjectID:      strings.TrimSpace(input.ProjectID),
		EventTypes:     normalizePolicyEventTypes(input.EventTypes),
		DestinationIDs: normalizeDestinationIDs(input.DestinationIDs),
		Enabled:        input.Enabled,
		CreatedAt:      s.now().UTC(),
		UpdatedAt:      s.now().UTC(),
	}
	if policy.ScopeType == "" {
		policy.ScopeType = PolicyScopeGlobal
	}
	if err := s.validatePolicy(ctx, policy); err != nil {
		return Policy{}, err
	}
	eventTypesJSON, err := marshalStringList(policy.EventTypes)
	if err != nil {
		return Policy{}, err
	}
	destinationIDsJSON, err := marshalStringList(policy.DestinationIDs)
	if err != nil {
		return Policy{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO notification_policies (
			id, name, scope_type, scope_key, project_id, event_types_json,
			destination_ids_json, enabled, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, policy.ID, policy.Name, policy.ScopeType, policyScopeKey(policy.ScopeType, policy.ProjectID),
		nullableString(policy.ProjectID), eventTypesJSON, destinationIDsJSON, policy.Enabled,
		formatTime(policy.CreatedAt), formatTime(policy.UpdatedAt)); err != nil {
		if isUniqueConstraint(err) {
			return Policy{}, fmt.Errorf("%w: policy name already exists in scope", ErrValidation)
		}
		return Policy{}, fmt.Errorf("insert notification policy: %w", err)
	}
	return policy, nil
}

func (s *Service) GetPolicy(ctx context.Context, policyID string) (Policy, error) {
	policy, err := scanPolicy(s.db.QueryRowContext(ctx, `
		SELECT id, name, scope_type, project_id, event_types_json, destination_ids_json, enabled, created_at, updated_at
		FROM notification_policies
		WHERE id = ? AND deleted_at IS NULL
	`, policyID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Policy{}, ErrNotFound
		}
		return Policy{}, fmt.Errorf("get notification policy: %w", err)
	}
	return policy, nil
}

func (s *Service) UpdatePolicy(ctx context.Context, policyID string, input UpdatePolicyInput) (Policy, error) {
	policy, err := s.GetPolicy(ctx, policyID)
	if err != nil {
		return Policy{}, err
	}
	if input.Name != nil {
		policy.Name = normalizePolicyName(*input.Name)
	}
	if input.EventTypes != nil {
		policy.EventTypes = normalizePolicyEventTypes(*input.EventTypes)
	}
	if input.DestinationIDs != nil {
		policy.DestinationIDs = normalizeDestinationIDs(*input.DestinationIDs)
	}
	if input.Enabled != nil {
		policy.Enabled = *input.Enabled
	}
	policy.UpdatedAt = s.now().UTC()
	if err := s.validatePolicy(ctx, policy); err != nil {
		return Policy{}, err
	}
	eventTypesJSON, err := marshalStringList(policy.EventTypes)
	if err != nil {
		return Policy{}, err
	}
	destinationIDsJSON, err := marshalStringList(policy.DestinationIDs)
	if err != nil {
		return Policy{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE notification_policies
		SET name = ?, event_types_json = ?, destination_ids_json = ?, enabled = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, policy.Name, eventTypesJSON, destinationIDsJSON, policy.Enabled, formatTime(policy.UpdatedAt), policyID)
	if err != nil {
		if isUniqueConstraint(err) {
			return Policy{}, fmt.Errorf("%w: policy name already exists in scope", ErrValidation)
		}
		return Policy{}, fmt.Errorf("update notification policy: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Policy{}, fmt.Errorf("check notification policy update: %w", err)
	}
	if affected == 0 {
		return Policy{}, ErrNotFound
	}
	return s.GetPolicy(ctx, policyID)
}

func (s *Service) DeletePolicy(ctx context.Context, policyID string) error {
	now := s.now().UTC()
	result, err := s.db.ExecContext(ctx, `
		UPDATE notification_policies
		SET deleted_at = ?, enabled = 0, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, formatTime(now), formatTime(now), policyID)
	if err != nil {
		return fmt.Errorf("delete notification policy: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check notification policy delete: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) validatePolicy(ctx context.Context, policy Policy) error {
	fields := map[string]string{}
	if err := validatePolicyScope(policy.ScopeType, policy.ProjectID); err != nil {
		return err
	}
	if policy.ScopeType == PolicyScopeProject {
		if err := s.requireProject(ctx, policy.ProjectID); err != nil {
			return err
		}
	}
	if policy.Name == "" {
		fields["name"] = "Required"
	}
	if len(policy.Name) > 80 {
		fields["name"] = "Must be 80 characters or fewer"
	}
	if len(policy.EventTypes) == 0 {
		fields["event_types"] = "At least one event type is required"
	}
	for _, eventType := range policy.EventTypes {
		if !slices.Contains(allowedPolicyEvents, eventType) {
			fields["event_types"] = "Contains an unsupported event type"
			break
		}
	}
	if len(policy.DestinationIDs) == 0 {
		fields["destination_ids"] = "At least one destination is required"
	}
	if len(fields) > 0 {
		return fmt.Errorf("%w: invalid notification policy", ErrValidation)
	}
	if err := s.validatePolicyDestinations(ctx, policy.ScopeType, policy.ProjectID, policy.DestinationIDs); err != nil {
		return err
	}
	return nil
}

func (s *Service) validatePolicyDestinations(ctx context.Context, scopeType string, projectID string, destinationIDs []string) error {
	for _, destinationID := range destinationIDs {
		destination, err := s.GetDestination(ctx, destinationID)
		if err != nil {
			return err
		}
		switch scopeType {
		case PolicyScopeGlobal:
			if destination.ScopeType != DestinationScopeGlobal {
				return fmt.Errorf("%w: global policies can only use global destinations", ErrValidation)
			}
		case PolicyScopeProject:
			if destination.ScopeType == DestinationScopeGlobal {
				continue
			}
			if destination.ScopeType != DestinationScopeProject || destination.ProjectID != projectID {
				return fmt.Errorf("%w: project policies can only use global or same-project destinations", ErrValidation)
			}
		}
	}
	return nil
}

func scanPolicy(scanner interface{ Scan(...any) error }) (Policy, error) {
	var policy Policy
	var projectID sql.NullString
	var eventTypesJSON string
	var destinationIDsJSON string
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(
		&policy.ID,
		&policy.Name,
		&policy.ScopeType,
		&projectID,
		&eventTypesJSON,
		&destinationIDsJSON,
		&policy.Enabled,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Policy{}, err
	}
	policy.ProjectID = nullString(projectID)
	eventTypes, err := unmarshalStringList(eventTypesJSON)
	if err != nil {
		return Policy{}, err
	}
	destinationIDs, err := unmarshalStringList(destinationIDsJSON)
	if err != nil {
		return Policy{}, err
	}
	created, err := parseTime(createdAt)
	if err != nil {
		return Policy{}, err
	}
	updated, err := parseTime(updatedAt)
	if err != nil {
		return Policy{}, err
	}
	policy.EventTypes = eventTypes
	policy.DestinationIDs = destinationIDs
	policy.CreatedAt = created
	policy.UpdatedAt = updated
	return policy, nil
}

func validatePolicyScope(scopeType string, projectID string) error {
	switch scopeType {
	case PolicyScopeGlobal:
		if projectID != "" {
			return fmt.Errorf("%w: global policies cannot include project ids", ErrValidation)
		}
	case PolicyScopeProject:
		if projectID == "" {
			return fmt.Errorf("%w: project policies require project_id", ErrValidation)
		}
	default:
		return fmt.Errorf("%w: invalid policy scope", ErrValidation)
	}
	return nil
}

func policyScopeKey(scopeType string, projectID string) string {
	if scopeType == PolicyScopeProject {
		return projectID
	}
	return "global"
}

func normalizePolicyName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func normalizePolicyEventTypes(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeDestinationIDs(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	return normalized
}

func marshalStringList(values []string) (string, error) {
	data, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("marshal notification policy list: %w", err)
	}
	return string(data), nil
}

func unmarshalStringList(value string) ([]string, error) {
	var values []string
	if value == "" {
		return values, nil
	}
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return nil, fmt.Errorf("unmarshal notification policy list: %w", err)
	}
	if values == nil {
		values = []string{}
	}
	return values, nil
}
