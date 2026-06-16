package components

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/components/{component_id}", "Components", "Get component"), provider.getComponent)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/components/{component_id}", "Components", "Update component"), provider.updateComponent)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/components/{component_id}", "Components", "Delete component", http.StatusNoContent), provider.deleteComponent)
}

func (provider Provider) getComponent(ctx context.Context, input *ComponentIDInput) (*ComponentOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	component, err := provider.Tracker.GetComponent(ctx, principal, input.ComponentID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ComponentOutput{Body: ResourceFromTracker(component)}, nil
}

func (provider Provider) updateComponent(ctx context.Context, input *UpdateComponentInput) (*ComponentOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	component, err := provider.Tracker.UpdateComponent(ctx, principal, input.ComponentID, input.Body.Spec.ToUpdateInput())
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ComponentOutput{Body: ResourceFromTracker(component)}, nil
}

func (provider Provider) deleteComponent(ctx context.Context, input *ComponentIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Tracker.DeleteComponent(ctx, principal, input.ComponentID); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}
