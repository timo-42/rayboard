package createpages

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	ticketapi "github.com/timo-42/rayboard/internal/backend/httpapi/tickets"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/ticket-create-pages", "Ticket Create Pages", "List project ticket create pages"), provider.listProjectPages)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects/{project_id}/ticket-create-pages", "Ticket Create Pages", "Create ticket create page", http.StatusCreated), provider.createProjectPage)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/ticket-create-pages/{page_id}", "Ticket Create Pages", "Get ticket create page"), provider.getPage)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/ticket-create-pages/{page_id}", "Ticket Create Pages", "Update ticket create page"), provider.updatePage)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/ticket-create-pages/{page_id}", "Ticket Create Pages", "Delete ticket create page", http.StatusNoContent), provider.deletePage)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/ticket-create-pages/{page_id}/preview", "Ticket Create Pages", "Preview ticket create page form logic"), provider.previewPage)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/ticket-create-pages/{page_id}/runs", "Ticket Create Pages", "List ticket create page runs"), provider.listPageRuns)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/ticket-create-pages/{slug}/schema", "Ticket Create Pages", "Resolve ticket create page schema"), provider.resolvePage)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/projects/{project_id}/ticket-create-pages/{slug}/submit", "Ticket Create Pages", "Submit ticket create page", http.StatusCreated), provider.submitPage)
}

func (provider Provider) listProjectPages(ctx context.Context, input *ProjectPagesInput) (*ListPagesOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	pages, err := provider.CreatePages.List(ctx, principal, tracker.ListCreatePagesInput{
		ProjectID:       input.ProjectID,
		IncludeDisabled: input.IncludeDisabled,
		Limit:           input.Limit,
		Offset:          input.Offset,
	})
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListPagesOutput{Body: shared.NewListResource[PageResource](pageResources(pages))}, nil
}

func (provider Provider) createProjectPage(ctx context.Context, input *CreateProjectPageInput) (*CreatePageOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	page, err := provider.CreatePages.Create(ctx, principal, input.Body.Spec.createInput(input.ProjectID))
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &CreatePageOutput{Body: pageResource(page)}, nil
}

func (provider Provider) getPage(ctx context.Context, input *PageIDInput) (*PageOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	page, err := provider.CreatePages.Get(ctx, principal, input.PageID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &PageOutput{Body: pageResource(page)}, nil
}

func (provider Provider) updatePage(ctx context.Context, input *UpdatePageInput) (*PageOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	page, err := provider.CreatePages.Update(ctx, principal, input.PageID, input.Body.Spec.updateInput())
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &PageOutput{Body: pageResource(page)}, nil
}

func (provider Provider) deletePage(ctx context.Context, input *PageIDInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.CreatePages.Delete(ctx, principal, input.PageID); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) previewPage(ctx context.Context, input *PageIDInput) (*PreviewPageOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	page, err := provider.CreatePages.Preview(ctx, principal, input.PageID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &PreviewPageOutput{Body: schemaResource(page)}, nil
}

func (provider Provider) listPageRuns(ctx context.Context, input *ListPageRunsInput) (*ListPageRunsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	runs, err := provider.CreatePages.ListRuns(ctx, principal, input.PageID, input.Limit, input.Offset)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ListPageRunsOutput{Body: shared.NewListResource[PageRunResource](pageRunResources(runs))}, nil
}

func (provider Provider) resolvePage(ctx context.Context, input *ResolvePageInput) (*SchemaOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	page, err := provider.CreatePages.Resolve(ctx, principal, input.ProjectID, input.Slug)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &SchemaOutput{Body: schemaResource(page)}, nil
}

func (provider Provider) submitPage(ctx context.Context, input *SubmitPageInput) (*SubmitPageOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	ticket, err := provider.CreatePages.Submit(ctx, principal, input.ProjectID, input.Slug, input.Body.Spec.submitInput())
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &SubmitPageOutput{Body: ticketapi.ResourceFromTracker(ticket)}, nil
}
