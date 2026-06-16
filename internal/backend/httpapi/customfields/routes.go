package customfields

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/custom-fields/{field_id}", "Custom Fields", "Get custom field"), provider.getField)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/custom-fields/{field_id}", "Custom Fields", "Update custom field"), provider.updateField)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/custom-fields/{field_id}", "Custom Fields", "Delete custom field", http.StatusNoContent), provider.deleteField)
}

func (provider Provider) getField(ctx context.Context, input *FieldIDInput) (*FieldOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	field, err := provider.Tracker.GetCustomField(ctx, principal, input.FieldID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &FieldOutput{Body: field}, nil
}

func (provider Provider) updateField(ctx context.Context, input *UpdateFieldInput) (*FieldOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	field, err := provider.Tracker.UpdateCustomField(ctx, principal, input.FieldID, input.Body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &FieldOutput{Body: field}, nil
}

func (provider Provider) deleteField(ctx context.Context, input *FieldIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Tracker.DeleteCustomField(ctx, principal, input.FieldID); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}
