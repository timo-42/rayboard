package sprints

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/sprints/{sprint_id}", "Sprints", "Get sprint"), provider.getSprint)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/sprints/{sprint_id}", "Sprints", "Update sprint"), provider.updateSprint)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/sprints/{sprint_id}", "Sprints", "Delete sprint", http.StatusNoContent), provider.deleteSprint)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/sprints/{sprint_id}/start", "Sprints", "Start sprint"), provider.startSprint)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/sprints/{sprint_id}/complete", "Sprints", "Complete sprint"), provider.completeSprint)
}

func (provider Provider) getSprint(ctx context.Context, input *SprintIDInput) (*SprintOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	sprint, err := provider.Tracker.GetSprint(ctx, principal, input.SprintID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &SprintOutput{Body: sprint}, nil
}

func (provider Provider) updateSprint(ctx context.Context, input *UpdateSprintInput) (*SprintOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	sprint, err := provider.Tracker.UpdateSprint(ctx, principal, input.SprintID, input.Body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &SprintOutput{Body: sprint}, nil
}

func (provider Provider) deleteSprint(ctx context.Context, input *SprintIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Tracker.DeleteSprint(ctx, principal, input.SprintID); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) startSprint(ctx context.Context, input *SprintIDInput) (*SprintOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	sprint, err := provider.Tracker.StartSprint(ctx, principal, input.SprintID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &SprintOutput{Body: sprint}, nil
}

func (provider Provider) completeSprint(ctx context.Context, input *CompleteSprintInput) (*SprintOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	sprint, err := provider.Tracker.CompleteSprint(ctx, principal, input.SprintID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &SprintOutput{Body: sprint}, nil
}
