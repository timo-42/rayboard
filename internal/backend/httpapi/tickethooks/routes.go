package tickethooks

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/ticket-hooks", "Ticket Hooks", "List project ticket hooks"), provider.listProjectHooks)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects/{project_id}/ticket-hooks", "Ticket Hooks", "Create project ticket hook", http.StatusCreated), provider.createProjectHook)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/ticket-hooks/{hook_id}", "Ticket Hooks", "Get ticket hook"), provider.getHook)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/ticket-hooks/{hook_id}", "Ticket Hooks", "Update ticket hook"), provider.updateHook)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/ticket-hooks/{hook_id}", "Ticket Hooks", "Delete ticket hook", http.StatusNoContent), provider.deleteHook)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/ticket-hooks/{hook_id}/preview", "Ticket Hooks", "Preview ticket hook"), provider.previewHook)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/ticket-hooks/{hook_id}/runs", "Ticket Hooks", "List ticket hook runs"), provider.listHookRuns)
}

func (provider Provider) listProjectHooks(ctx context.Context, input *ProjectHooksInput) (*ListHooksOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	hooks, err := provider.Hooks.List(ctx, principal, tracker.ListHooksInput{
		ProjectID: input.ProjectID,
		Event:     input.Event,
		Phase:     input.Phase,
		Limit:     input.Limit,
		Offset:    input.Offset,
	})
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListHooksOutput{Body: shared.NewListResource[HookResource](hookResources(hooks))}, nil
}

func (provider Provider) createProjectHook(ctx context.Context, input *CreateProjectHookInput) (*CreateHookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	hook, err := provider.Hooks.Create(ctx, principal, input.Body.Spec.createInput(input.ProjectID))
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &CreateHookOutput{Body: hookResource(hook)}, nil
}

func (provider Provider) getHook(ctx context.Context, input *HookIDInput) (*HookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	hook, err := provider.Hooks.Get(ctx, principal, input.HookID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &HookOutput{Body: hookResource(hook)}, nil
}

func (provider Provider) updateHook(ctx context.Context, input *UpdateHookInput) (*HookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	hook, err := provider.Hooks.Update(ctx, principal, input.HookID, input.Body.Spec.updateInput())
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &HookOutput{Body: hookResource(hook)}, nil
}

func (provider Provider) deleteHook(ctx context.Context, input *HookIDInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Hooks.Delete(ctx, principal, input.HookID); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) previewHook(ctx context.Context, input *PreviewHookInput) (*PreviewHookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	preview, err := provider.Hooks.Preview(ctx, principal, input.HookID, input.Body.Spec.previewInput())
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &PreviewHookOutput{Body: previewResource(preview)}, nil
}

func (provider Provider) listHookRuns(ctx context.Context, input *ListHookRunsInput) (*ListHookRunsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	runs, err := provider.Hooks.ListRuns(ctx, principal, input.HookID, input.Limit, input.Offset)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListHookRunsOutput{Body: shared.NewListResource[HookRunResource](hookRunResources(runs))}, nil
}
