package searchapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/search"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodPost, "/api/search", "Search", "Search tickets with CEL query"), provider.searchTickets)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/saved-views", "Saved Views", "List saved views"), provider.listSavedViews)
	huma.Register(api, operation(http.MethodPost, "/api/saved-views", "Saved Views", "Create saved view", http.StatusCreated), provider.createSavedView)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/saved-views/{view_id}", "Saved Views", "Get saved view"), provider.getSavedView)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/saved-views/{view_id}", "Saved Views", "Update saved view"), provider.updateSavedView)
	huma.Register(api, operation(http.MethodDelete, "/api/saved-views/{view_id}", "Saved Views", "Delete saved view", http.StatusNoContent), provider.deleteSavedView)
}

func (provider Provider) searchTickets(ctx context.Context, input *SearchTicketsInput) (*SearchTicketsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	result, err := provider.Search.SearchTickets(ctx, principal, input.Body)
	if err != nil {
		return nil, shared.SearchError(err)
	}
	return &SearchTicketsOutput{Body: result}, nil
}

func (provider Provider) listSavedViews(ctx context.Context, input *ListSavedViewsInput) (*ListSavedViewsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	views, err := provider.Search.ListSavedViews(ctx, principal, search.ListSavedViewsInput{
		ProjectID: input.ProjectID,
		Pinned:    input.Pinned,
		Limit:     input.Limit,
		Offset:    input.Offset,
	})
	if err != nil {
		return nil, shared.SearchError(err)
	}
	return &ListSavedViewsOutput{Body: shared.ItemList[search.SavedView]{Items: views}}, nil
}

func (provider Provider) createSavedView(ctx context.Context, input *CreateSavedViewInput) (*CreateSavedViewOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	view, err := provider.Search.CreateSavedView(ctx, principal, input.Body)
	if err != nil {
		return nil, shared.SearchError(err)
	}
	return &CreateSavedViewOutput{Body: view}, nil
}

func (provider Provider) getSavedView(ctx context.Context, input *SavedViewIDInput) (*SavedViewOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	view, err := provider.Search.GetSavedView(ctx, principal, input.ViewID)
	if err != nil {
		return nil, shared.SearchError(err)
	}
	return &SavedViewOutput{Body: view}, nil
}

func (provider Provider) updateSavedView(ctx context.Context, input *UpdateSavedViewInput) (*SavedViewOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	view, err := provider.Search.UpdateSavedView(ctx, principal, input.ViewID, input.Body)
	if err != nil {
		return nil, shared.SearchError(err)
	}
	return &SavedViewOutput{Body: view}, nil
}

func (provider Provider) deleteSavedView(ctx context.Context, input *SavedViewIDInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Search.DeleteSavedView(ctx, principal, input.ViewID); err != nil {
		return nil, shared.SearchError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func operation(method string, path string, tag string, summary string, status int) huma.Operation {
	op := shared.Operation(method, path, tag, summary)
	op.DefaultStatus = status
	return op
}
