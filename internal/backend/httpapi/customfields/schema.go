package customfields

import (
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
	Body    tracker.UpdateCustomFieldInput
}

type FieldOutput struct {
	Body tracker.CustomFieldDefinition
}
