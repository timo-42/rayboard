package openrouterapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/openrouter-providers", "OpenRouter Providers", "List OpenRouter providers"), provider.listProviders)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/openrouter-providers", "OpenRouter Providers", "Create OpenRouter provider", http.StatusCreated), provider.createProvider)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/openrouter-providers/{provider_id}", "OpenRouter Providers", "Get OpenRouter provider"), provider.getProvider)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/openrouter-providers/{provider_id}", "OpenRouter Providers", "Update OpenRouter provider"), provider.updateProvider)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/openrouter-providers/{provider_id}", "OpenRouter Providers", "Delete OpenRouter provider", http.StatusNoContent), provider.deleteProvider)
}

func (provider Provider) listProviders(ctx context.Context, input *struct{ shared.AuthInput }) (*ListProvidersOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionAIManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	providers, err := provider.OpenRouter.ListProviders(ctx)
	if err != nil {
		return nil, shared.OpenRouterError(err)
	}
	return &ListProvidersOutput{Body: shared.ItemList[ProviderResource]{Items: providerResources(providers)}}, nil
}

func (provider Provider) createProvider(ctx context.Context, input *CreateProviderInput) (*CreateProviderOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionAIManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	created, err := provider.OpenRouter.CreateProvider(ctx, input.Body.Spec.createInput())
	if err != nil {
		return nil, shared.OpenRouterError(err)
	}
	if err := provider.recordAudit(ctx, principal, audit.RecordInput{
		EventType:   "openrouter.provider_created",
		SubjectType: "openrouter_provider",
		SubjectID:   created.ID,
		Payload: map[string]any{
			"provider_id":   created.ID,
			"name":          created.Name,
			"default_model": created.DefaultModel,
			"api_key_set":   created.APIKeySet,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &CreateProviderOutput{Body: providerResource(created)}, nil
}

func (provider Provider) getProvider(ctx context.Context, input *ProviderIDInput) (*ProviderOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionAIManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	found, err := provider.OpenRouter.GetProvider(ctx, input.ProviderID)
	if err != nil {
		return nil, shared.OpenRouterError(err)
	}
	return &ProviderOutput{Body: providerResource(found)}, nil
}

func (provider Provider) updateProvider(ctx context.Context, input *UpdateProviderInput) (*ProviderOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionAIManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	updated, err := provider.OpenRouter.UpdateProvider(ctx, input.ProviderID, input.Body.Spec.updateInput())
	if err != nil {
		return nil, shared.OpenRouterError(err)
	}
	payload := map[string]any{
		"provider_id":   updated.ID,
		"name":          updated.Name,
		"default_model": updated.DefaultModel,
		"api_key_set":   updated.APIKeySet,
	}
	if input.Body.Spec.APIKey != nil {
		payload["api_key_rotated"] = true
	}
	if err := provider.recordAudit(ctx, principal, audit.RecordInput{
		EventType:   "openrouter.provider_updated",
		SubjectType: "openrouter_provider",
		SubjectID:   updated.ID,
		Payload:     payload,
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &ProviderOutput{Body: providerResource(updated)}, nil
}

func (provider Provider) deleteProvider(ctx context.Context, input *ProviderIDInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionAIManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	if err := provider.OpenRouter.DeleteProvider(ctx, input.ProviderID); err != nil {
		return nil, shared.OpenRouterError(err)
	}
	if err := provider.recordAudit(ctx, principal, audit.RecordInput{
		EventType:   "openrouter.provider_deleted",
		SubjectType: "openrouter_provider",
		SubjectID:   input.ProviderID,
		Payload: map[string]any{
			"provider_id": input.ProviderID,
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &shared.EmptyOutput{}, nil
}
