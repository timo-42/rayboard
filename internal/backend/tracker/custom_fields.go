package tracker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/authz"
)

func (s *Service) ListCustomFields(ctx context.Context, principal authz.Principal, projectID string) ([]CustomFieldDefinition, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(projectID)); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	fields, err := s.listCustomFields(ctx, s.db, projectID)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

func (s *Service) CreateCustomField(ctx context.Context, principal authz.Principal, input CreateCustomFieldInput) (CustomFieldDefinition, error) {
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	if input.ProjectID == "" {
		return CustomFieldDefinition{}, validationFailed(map[string]string{"project_id": "Required"})
	}
	if err := s.require(principal, authz.PermissionFieldsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return CustomFieldDefinition{}, err
	}
	if _, err := s.repo.GetProject(ctx, input.ProjectID); err != nil {
		return CustomFieldDefinition{}, err
	}
	field, err := s.buildCustomField(input)
	if err != nil {
		return CustomFieldDefinition{}, err
	}
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		if err := s.insertCustomField(ctx, tx, field); err != nil {
			return err
		}
		return s.replaceCustomFieldOptions(ctx, tx, field.ID, field.Options, field.CreatedAt)
	}); err != nil {
		return CustomFieldDefinition{}, err
	}
	return field, nil
}

func (s *Service) GetCustomField(ctx context.Context, principal authz.Principal, fieldID string) (CustomFieldDefinition, error) {
	field, err := s.getCustomField(ctx, s.db, fieldID)
	if err != nil {
		return CustomFieldDefinition{}, err
	}
	if err := s.require(principal, authz.PermissionProjectsRead, authz.ProjectScope(field.ProjectID)); err != nil {
		return CustomFieldDefinition{}, err
	}
	return field, nil
}

func (s *Service) UpdateCustomField(ctx context.Context, principal authz.Principal, fieldID string, input UpdateCustomFieldInput) (CustomFieldDefinition, error) {
	current, err := s.getCustomField(ctx, s.db, fieldID)
	if err != nil {
		return CustomFieldDefinition{}, err
	}
	if err := s.require(principal, authz.PermissionFieldsManage, authz.ProjectScope(current.ProjectID)); err != nil {
		return CustomFieldDefinition{}, err
	}
	updated := current
	if input.Key != nil {
		updated.Key = normalizeCustomFieldKey(*input.Key)
	}
	if input.Name != nil {
		updated.Name = strings.TrimSpace(*input.Name)
	}
	if input.FieldType != nil {
		updated.FieldType = normalizeSlug(*input.FieldType)
	}
	if input.Required != nil {
		updated.Required = *input.Required
	}
	if input.Options != nil {
		options, fields := buildCustomFieldOptions(updated.ID, updated.FieldType, *input.Options, s.now().UTC())
		if len(fields) > 0 {
			return CustomFieldDefinition{}, validationFailed(fields)
		}
		updated.Options = options
	}
	if fields := customFieldDefinitionFields(updated); len(fields) > 0 {
		return CustomFieldDefinition{}, validationFailed(fields)
	}
	updated.UpdatedAt = s.now().UTC()
	if err := s.withTx(ctx, func(tx *sql.Tx) error {
		if err := s.updateCustomField(ctx, tx, updated); err != nil {
			return err
		}
		if input.Options != nil {
			return s.replaceCustomFieldOptions(ctx, tx, updated.ID, updated.Options, updated.UpdatedAt)
		}
		return nil
	}); err != nil {
		return CustomFieldDefinition{}, err
	}
	return updated, nil
}

func (s *Service) DeleteCustomField(ctx context.Context, principal authz.Principal, fieldID string) error {
	field, err := s.getCustomField(ctx, s.db, fieldID)
	if err != nil {
		return err
	}
	if err := s.require(principal, authz.PermissionFieldsManage, authz.ProjectScope(field.ProjectID)); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM custom_field_definitions WHERE id = ?", field.ID)
	if err != nil {
		return fmt.Errorf("delete custom field: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check custom field delete: %w", err)
	}
	if affected == 0 {
		return notFound("custom_field", fieldID)
	}
	return nil
}

func (s *Service) buildCustomField(input CreateCustomFieldInput) (CustomFieldDefinition, error) {
	id, err := newID("field")
	if err != nil {
		return CustomFieldDefinition{}, err
	}
	now := s.now().UTC()
	field := CustomFieldDefinition{
		ID:        id,
		ProjectID: strings.TrimSpace(input.ProjectID),
		Key:       normalizeCustomFieldKey(input.Key),
		Name:      strings.TrimSpace(input.Name),
		FieldType: normalizeSlug(input.FieldType),
		Required:  input.Required,
		CreatedAt: now,
		UpdatedAt: now,
	}
	options, optionFields := buildCustomFieldOptions(field.ID, field.FieldType, input.Options, now)
	field.Options = options
	fields := customFieldDefinitionFields(field)
	for key, value := range optionFields {
		fields[key] = value
	}
	if len(fields) > 0 {
		return CustomFieldDefinition{}, validationFailed(fields)
	}
	return field, nil
}

func customFieldDefinitionFields(field CustomFieldDefinition) map[string]string {
	fields := map[string]string{}
	if field.Key == "" {
		fields["key"] = "Required"
	} else if !slugPattern.MatchString(field.Key) {
		fields["key"] = "Must be a lowercase slug"
	}
	if field.Name == "" {
		fields["name"] = "Required"
	} else if len(field.Name) > 200 {
		fields["name"] = "Must be 200 characters or fewer"
	}
	if !validCustomFieldType(field.FieldType) {
		fields["field_type"] = "Invalid custom field type"
	}
	if selectCustomField(field.FieldType) && len(field.Options) == 0 {
		fields["options"] = "At least one option is required"
	}
	if !selectCustomField(field.FieldType) && len(field.Options) > 0 {
		fields["options"] = "Options are only valid for select fields"
	}
	return fields
}

func buildCustomFieldOptions(fieldID string, fieldType string, values []string, createdAt time.Time) ([]CustomFieldOption, map[string]string) {
	fields := map[string]string{}
	seen := map[string]struct{}{}
	options := make([]CustomFieldOption, 0, len(values))
	for index, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			fields["options"] = "Options cannot be empty"
			continue
		}
		if len(trimmed) > 200 {
			fields["options"] = "Options must be 200 characters or fewer"
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			fields["options"] = "Options must be unique"
			continue
		}
		seen[key] = struct{}{}
		options = append(options, CustomFieldOption{
			ID:        fmt.Sprintf("%s_option_%06d", fieldID, index+1),
			FieldID:   fieldID,
			Value:     trimmed,
			Position:  index,
			CreatedAt: createdAt,
		})
	}
	return options, fields
}

func validCustomFieldType(fieldType string) bool {
	switch fieldType {
	case CustomFieldTypeText, CustomFieldTypeNumber, CustomFieldTypeBoolean, CustomFieldTypeDate, CustomFieldTypeSingleSelect, CustomFieldTypeMultiSelect, CustomFieldTypeUser:
		return true
	default:
		return false
	}
}

func selectCustomField(fieldType string) bool {
	return fieldType == CustomFieldTypeSingleSelect || fieldType == CustomFieldTypeMultiSelect
}

func normalizeCustomFieldKey(key string) string {
	return normalizeSlug(key)
}

func (s *Service) insertCustomField(ctx context.Context, q sqlRunner, field CustomFieldDefinition) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO custom_field_definitions (
			id, project_id, key, name, field_type, required, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, field.ID, field.ProjectID, field.Key, field.Name, field.FieldType, boolInt(field.Required), formatTime(field.CreatedAt), formatTime(field.UpdatedAt))
	if err != nil {
		if isUniqueConstraint(err) {
			return conflict("custom_field", "key", field.Key)
		}
		return fmt.Errorf("insert custom field: %w", err)
	}
	return nil
}

func (s *Service) updateCustomField(ctx context.Context, q sqlRunner, field CustomFieldDefinition) error {
	_, err := q.ExecContext(ctx, `
		UPDATE custom_field_definitions
		SET key = ?, name = ?, field_type = ?, required = ?, updated_at = ?
		WHERE id = ?
	`, field.Key, field.Name, field.FieldType, boolInt(field.Required), formatTime(field.UpdatedAt), field.ID)
	if err != nil {
		if isUniqueConstraint(err) {
			return conflict("custom_field", "key", field.Key)
		}
		return fmt.Errorf("update custom field: %w", err)
	}
	return nil
}

func (s *Service) replaceCustomFieldOptions(ctx context.Context, q sqlRunner, fieldID string, options []CustomFieldOption, createdAt time.Time) error {
	if _, err := q.ExecContext(ctx, "DELETE FROM custom_field_options WHERE field_id = ?", fieldID); err != nil {
		return fmt.Errorf("delete custom field options: %w", err)
	}
	for _, option := range options {
		if option.CreatedAt.IsZero() {
			option.CreatedAt = createdAt
		}
		if _, err := q.ExecContext(ctx, `
			INSERT INTO custom_field_options (id, field_id, value, position, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, option.ID, fieldID, option.Value, option.Position, formatTime(option.CreatedAt)); err != nil {
			return fmt.Errorf("insert custom field option: %w", err)
		}
	}
	return nil
}

func (s *Service) getCustomField(ctx context.Context, q sqlRunner, fieldID string) (CustomFieldDefinition, error) {
	fieldID = strings.TrimSpace(fieldID)
	if fieldID == "" {
		return CustomFieldDefinition{}, validationFailed(map[string]string{"field_id": "Required"})
	}
	field, err := scanCustomField(q.QueryRowContext(ctx, `
		SELECT id, project_id, key, name, field_type, required, created_at, updated_at
		FROM custom_field_definitions
		WHERE id = ?
	`, fieldID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CustomFieldDefinition{}, notFound("custom_field", fieldID)
		}
		return CustomFieldDefinition{}, fmt.Errorf("get custom field: %w", err)
	}
	options, err := s.listCustomFieldOptions(ctx, q, field.ID)
	if err != nil {
		return CustomFieldDefinition{}, err
	}
	field.Options = options
	return field, nil
}

func (s *Service) listCustomFields(ctx context.Context, q sqlRunner, projectID string) ([]CustomFieldDefinition, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, project_id, key, name, field_type, required, created_at, updated_at
		FROM custom_field_definitions
		WHERE project_id = ?
		ORDER BY name ASC, id ASC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list custom fields: %w", err)
	}
	defer rows.Close()
	var fields []CustomFieldDefinition
	for rows.Next() {
		field, err := scanCustomField(rows)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate custom fields: %w", err)
	}
	for index := range fields {
		options, err := s.listCustomFieldOptions(ctx, q, fields[index].ID)
		if err != nil {
			return nil, err
		}
		fields[index].Options = options
	}
	return fields, nil
}

func (s *Service) listCustomFieldOptions(ctx context.Context, q sqlRunner, fieldID string) ([]CustomFieldOption, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, field_id, value, position, created_at
		FROM custom_field_options
		WHERE field_id = ?
		ORDER BY position ASC, id ASC
	`, fieldID)
	if err != nil {
		return nil, fmt.Errorf("list custom field options: %w", err)
	}
	defer rows.Close()
	var options []CustomFieldOption
	for rows.Next() {
		var option CustomFieldOption
		var createdAt string
		if err := rows.Scan(&option.ID, &option.FieldID, &option.Value, &option.Position, &createdAt); err != nil {
			return nil, err
		}
		parsed, err := parseTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse custom field option created_at: %w", err)
		}
		option.CreatedAt = parsed
		options = append(options, option)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate custom field options: %w", err)
	}
	return options, nil
}

func scanCustomField(scanner rowScanner) (CustomFieldDefinition, error) {
	var field CustomFieldDefinition
	var required int
	var createdAt string
	var updatedAt string
	if err := scanner.Scan(&field.ID, &field.ProjectID, &field.Key, &field.Name, &field.FieldType, &required, &createdAt, &updatedAt); err != nil {
		return CustomFieldDefinition{}, err
	}
	field.Required = required != 0
	var err error
	field.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return CustomFieldDefinition{}, fmt.Errorf("parse custom field created_at: %w", err)
	}
	field.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return CustomFieldDefinition{}, fmt.Errorf("parse custom field updated_at: %w", err)
	}
	return field, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (s *Service) validateCustomFieldValues(ctx context.Context, q sqlRunner, projectID string, values map[string]any, requireAll bool) (map[string]customFieldValue, error) {
	fields, err := s.listCustomFields(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	byKey := map[string]CustomFieldDefinition{}
	for _, field := range fields {
		byKey[field.Key] = field
	}
	fieldErrors := map[string]string{}
	normalized := map[string]customFieldValue{}
	for key, value := range values {
		normalizedKey := normalizeCustomFieldKey(key)
		field, ok := byKey[normalizedKey]
		if !ok {
			fieldErrors["custom_fields."+key] = "Unknown custom field"
			continue
		}
		converted, err := s.validateCustomFieldValue(ctx, q, field, value)
		if err != nil {
			fieldErrors["custom_fields."+field.Key] = err.Error()
			continue
		}
		if converted == nil {
			continue
		}
		normalized[field.ID] = customFieldValue{Field: field, Value: converted}
	}
	if requireAll {
		for _, field := range fields {
			if field.Required {
				if _, ok := normalized[field.ID]; !ok {
					fieldErrors["custom_fields."+field.Key] = "Required"
				}
			}
		}
	}
	if len(fieldErrors) > 0 {
		return nil, validationFailed(fieldErrors)
	}
	return normalized, nil
}

type customFieldValue struct {
	Field CustomFieldDefinition
	Value any
}

func (s *Service) validateCustomFieldValue(ctx context.Context, q sqlRunner, field CustomFieldDefinition, value any) (any, error) {
	if value == nil {
		if field.Required {
			return nil, errors.New("Required")
		}
		return nil, nil
	}
	switch field.FieldType {
	case CustomFieldTypeText:
		text, ok := value.(string)
		if !ok {
			return nil, errors.New("Must be a string")
		}
		text = strings.TrimSpace(text)
		if text == "" {
			if field.Required {
				return nil, errors.New("Required")
			}
			return nil, nil
		}
		if len(text) > 2000 {
			return nil, errors.New("Must be 2000 characters or fewer")
		}
		return text, nil
	case CustomFieldTypeNumber:
		number, ok := numberValue(value)
		if !ok {
			return nil, errors.New("Must be a number")
		}
		return number, nil
	case CustomFieldTypeBoolean:
		boolean, ok := value.(bool)
		if !ok {
			return nil, errors.New("Must be a boolean")
		}
		return boolean, nil
	case CustomFieldTypeDate:
		date, ok := value.(string)
		if !ok {
			return nil, errors.New("Must be a date string")
		}
		date = strings.TrimSpace(date)
		dateFields := map[string]string{}
		validateSprintDate(dateFields, field.Key, date)
		if len(dateFields) > 0 {
			return nil, errors.New("Must use YYYY-MM-DD")
		}
		if date == "" {
			if field.Required {
				return nil, errors.New("Required")
			}
			return nil, nil
		}
		return date, nil
	case CustomFieldTypeSingleSelect:
		text, ok := value.(string)
		if !ok {
			return nil, errors.New("Must be a string")
		}
		text = strings.TrimSpace(text)
		if text == "" {
			if field.Required {
				return nil, errors.New("Required")
			}
			return nil, nil
		}
		if !fieldOptionExists(field, text) {
			return nil, errors.New("Must be one of the configured options")
		}
		return text, nil
	case CustomFieldTypeMultiSelect:
		values, ok := value.([]any)
		if !ok {
			return nil, errors.New("Must be an array")
		}
		selected := make([]string, 0, len(values))
		seen := map[string]struct{}{}
		for _, item := range values {
			text, ok := item.(string)
			if !ok {
				return nil, errors.New("Must contain only strings")
			}
			text = strings.TrimSpace(text)
			if text == "" {
				continue
			}
			if !fieldOptionExists(field, text) {
				return nil, errors.New("Must contain only configured options")
			}
			key := strings.ToLower(text)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			selected = append(selected, text)
		}
		sort.Strings(selected)
		if len(selected) == 0 {
			if field.Required {
				return nil, errors.New("Required")
			}
			return nil, nil
		}
		return selected, nil
	case CustomFieldTypeUser:
		userID, ok := value.(string)
		if !ok {
			return nil, errors.New("Must be a user ID")
		}
		userID = strings.TrimSpace(userID)
		if userID == "" {
			if field.Required {
				return nil, errors.New("Required")
			}
			return nil, nil
		}
		if err := s.requireUser(ctx, q, field.Key, userID); err != nil {
			return nil, errors.New("User not found")
		}
		return userID, nil
	default:
		return nil, errors.New("Unsupported field type")
	}
}

func numberValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	default:
		return 0, false
	}
}

func fieldOptionExists(field CustomFieldDefinition, value string) bool {
	for _, option := range field.Options {
		if strings.EqualFold(option.Value, value) {
			return true
		}
	}
	return false
}

func (s *Service) replaceTicketCustomFieldValues(ctx context.Context, q sqlRunner, ticketID string, values map[string]customFieldValue, updatedAt time.Time) error {
	if _, err := q.ExecContext(ctx, "DELETE FROM ticket_custom_field_values WHERE ticket_id = ?", ticketID); err != nil {
		return fmt.Errorf("delete ticket custom field values: %w", err)
	}
	for fieldID, value := range values {
		encoded, err := json.Marshal(value.Value)
		if err != nil {
			return fmt.Errorf("encode custom field value: %w", err)
		}
		if _, err := q.ExecContext(ctx, `
			INSERT INTO ticket_custom_field_values (ticket_id, field_id, value_json, updated_at)
			VALUES (?, ?, ?, ?)
		`, ticketID, fieldID, string(encoded), formatTime(updatedAt)); err != nil {
			return fmt.Errorf("insert ticket custom field value: %w", err)
		}
	}
	return nil
}

func (s *Service) loadTicketCustomFields(ctx context.Context, ticketID string) (map[string]any, error) {
	values, err := s.loadTicketCustomFieldsForTickets(ctx, []string{ticketID})
	if err != nil {
		return nil, err
	}
	return values[ticketID], nil
}

func (s *Service) loadTicketCustomFieldsForTickets(ctx context.Context, ticketIDs []string) (map[string]map[string]any, error) {
	result := map[string]map[string]any{}
	if len(ticketIDs) == 0 {
		return result, nil
	}
	placeholders := make([]string, len(ticketIDs))
	args := make([]any, len(ticketIDs))
	for index, id := range ticketIDs {
		placeholders[index] = "?"
		args[index] = id
		result[id] = map[string]any{}
	}
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT v.ticket_id, f.key, v.value_json
		FROM ticket_custom_field_values v
		JOIN custom_field_definitions f ON f.id = v.field_id
		WHERE v.ticket_id IN (%s)
		ORDER BY f.name ASC, f.id ASC
	`, strings.Join(placeholders, ",")), args...)
	if err != nil {
		return nil, fmt.Errorf("list ticket custom field values: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ticketID string
		var key string
		var raw string
		if err := rows.Scan(&ticketID, &key, &raw); err != nil {
			return nil, err
		}
		var value any
		if err := json.Unmarshal([]byte(raw), &value); err != nil {
			return nil, fmt.Errorf("decode custom field value: %w", err)
		}
		result[ticketID][key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket custom field values: %w", err)
	}
	return result, nil
}

func customFieldValueMap(values map[string]customFieldValue) map[string]any {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]any, len(values))
	for _, value := range values {
		result[value.Field.Key] = value.Value
	}
	return result
}

func (s *Service) attachCustomFields(ctx context.Context, ticket Ticket) (Ticket, error) {
	values, err := s.loadTicketCustomFields(ctx, ticket.ID)
	if err != nil {
		return Ticket{}, err
	}
	if len(values) > 0 {
		ticket.CustomFields = values
	}
	return ticket, nil
}

func (s *Service) attachCustomFieldsToTickets(ctx context.Context, tickets []Ticket) ([]Ticket, error) {
	ids := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		ids = append(ids, ticket.ID)
	}
	values, err := s.loadTicketCustomFieldsForTickets(ctx, ids)
	if err != nil {
		return nil, err
	}
	for index := range tickets {
		if len(values[tickets[index].ID]) > 0 {
			tickets[index].CustomFields = values[tickets[index].ID]
		}
	}
	return tickets, nil
}
