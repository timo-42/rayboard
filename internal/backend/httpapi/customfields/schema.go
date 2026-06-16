package customfields

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type FieldIDInput struct {
	shared.AuthInput
	FieldID string `path:"field_id"`
}

type UpdateFieldInput struct {
	shared.AuthInput
	FieldID string `path:"field_id"`
	Body    shared.ResourceInput[UpdateFieldSpec]
}

type FieldOutput struct {
	Body FieldResource
}

type FieldMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FieldSpec struct {
	Key       string   `json:"key,omitempty"`
	Name      string   `json:"name,omitempty"`
	FieldType string   `json:"field_type,omitempty"`
	Required  bool     `json:"required,omitempty"`
	Options   []string `json:"options,omitempty"`
}

type UpdateFieldSpec struct {
	Key       *string   `json:"key,omitempty"`
	Name      *string   `json:"name,omitempty"`
	FieldType *string   `json:"field_type,omitempty"`
	Required  *bool     `json:"required,omitempty"`
	Options   *[]string `json:"options,omitempty"`
}

type FieldStatus struct {
	Options []tracker.CustomFieldOption `json:"options,omitempty"`
}

type FieldResource = shared.Resource[FieldMetadata, FieldSpec, FieldStatus]

func (spec FieldSpec) ToCreateInput(projectID string) tracker.CreateCustomFieldInput {
	return tracker.CreateCustomFieldInput{
		ProjectID: projectID,
		Key:       spec.Key,
		Name:      spec.Name,
		FieldType: spec.FieldType,
		Required:  spec.Required,
		Options:   spec.Options,
	}
}

func (spec UpdateFieldSpec) ToUpdateInput() tracker.UpdateCustomFieldInput {
	return tracker.UpdateCustomFieldInput{
		Key:       spec.Key,
		Name:      spec.Name,
		FieldType: spec.FieldType,
		Required:  spec.Required,
		Options:   spec.Options,
	}
}

func ResourceFromTracker(field tracker.CustomFieldDefinition) FieldResource {
	options := make([]string, 0, len(field.Options))
	for _, option := range field.Options {
		options = append(options, option.Value)
	}
	return FieldResource{
		Metadata: FieldMetadata{
			ID:        field.ID,
			ProjectID: field.ProjectID,
			CreatedAt: field.CreatedAt,
			UpdatedAt: field.UpdatedAt,
		},
		Spec: FieldSpec{
			Key:       field.Key,
			Name:      field.Name,
			FieldType: field.FieldType,
			Required:  field.Required,
			Options:   options,
		},
		Status: FieldStatus{
			Options: field.Options,
		},
	}
}

func ResourcesFromTracker(fields []tracker.CustomFieldDefinition) []FieldResource {
	resources := make([]FieldResource, 0, len(fields))
	for _, field := range fields {
		resources = append(resources, ResourceFromTracker(field))
	}
	return resources
}
