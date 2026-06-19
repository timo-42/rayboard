package tickets

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/tickets/{ticket_id}", "Tickets", "Get ticket"), provider.getTicket)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/tickets/{ticket_id}", "Tickets", "Update ticket"), provider.updateTicket)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/tickets/{ticket_id}", "Tickets", "Delete ticket", http.StatusNoContent), provider.deleteTicket)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/tickets/{ticket_id}/activity", "Tickets", "List ticket activity"), provider.listActivity)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/tickets/{ticket_id}/watchers", "Ticket Watchers", "List ticket watchers"), provider.listWatchers)
	huma.Register(api, shared.Operation(http.MethodPut, "/api/tickets/{ticket_id}/watchers/me", "Ticket Watchers", "Watch ticket"), provider.watchTicket)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/tickets/{ticket_id}/watchers/me", "Ticket Watchers", "Unwatch ticket", http.StatusNoContent), provider.unwatchTicket)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/tickets/{ticket_id}/links", "Ticket Links", "List ticket links"), provider.listLinks)
	huma.Register(api, shared.OperationWithStatus(http.MethodPost, "/api/tickets/{ticket_id}/links", "Ticket Links", "Create ticket link", http.StatusCreated), provider.createLink)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/tickets/{ticket_id}/links/{link_id}", "Ticket Links", "Delete ticket link", http.StatusNoContent), provider.deleteLink)
	huma.Register(api, shared.Operation(http.MethodPut, "/api/tickets/{ticket_id}/sprint", "Sprints", "Assign ticket to sprint"), provider.assignSprint)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/tickets/{ticket_id}/sprint", "Sprints", "Assign ticket to sprint"), provider.assignSprint)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/tickets/{ticket_id}/sprint", "Sprints", "Remove ticket from sprint", http.StatusNoContent), provider.removeSprint)
}

func (provider Provider) getTicket(ctx context.Context, input *TicketIDInput) (*TicketOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	ticket, err := provider.Tracker.GetTicket(ctx, principal, input.TicketID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &TicketOutput{Body: ResourceFromTracker(ticket)}, nil
}

func (provider Provider) updateTicket(ctx context.Context, input *UpdateTicketInput) (*TicketOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	ticket, err := provider.Tracker.UpdateTicket(ctx, principal, input.TicketID, input.Body.Spec.ToUpdateInput())
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &TicketOutput{Body: ResourceFromTracker(ticket)}, nil
}

func (provider Provider) deleteTicket(ctx context.Context, input *TicketIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Tracker.DeleteTicket(ctx, principal, input.TicketID); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) listActivity(ctx context.Context, input *TicketIDInput) (*ActivityOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListTicketActivity(ctx, principal, input.TicketID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ActivityOutput{Body: shared.NewListResource[ActivityResource](ActivityResourcesFromTracker(items))}, nil
}

func (provider Provider) listWatchers(ctx context.Context, input *TicketIDInput) (*TicketWatchersOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListTicketWatchers(ctx, principal, input.TicketID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &TicketWatchersOutput{Body: shared.NewListResource[TicketWatcherResource](WatcherResourcesFromTracker(items))}, nil
}

func (provider Provider) watchTicket(ctx context.Context, input *TicketIDInput) (*TicketOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	ticket, err := provider.Tracker.WatchTicket(ctx, principal, input.TicketID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &TicketOutput{Body: ResourceFromTracker(ticket)}, nil
}

func (provider Provider) unwatchTicket(ctx context.Context, input *TicketIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if _, err := provider.Tracker.UnwatchTicket(ctx, principal, input.TicketID); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) listLinks(ctx context.Context, input *TicketIDInput) (*TicketLinksOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListTicketLinks(ctx, principal, input.TicketID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &TicketLinksOutput{Body: shared.NewListResource[TicketLinkResource](LinkResourcesFromTracker(items))}, nil
}

func (provider Provider) createLink(ctx context.Context, input *CreateTicketLinkInput) (*CreateTicketLinkOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	link, err := provider.Tracker.CreateTicketLink(ctx, principal, input.TicketID, input.Body.Spec.ToCreateInput())
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &CreateTicketLinkOutput{Body: LinkResourceFromTracker(link)}, nil
}

func (provider Provider) deleteLink(ctx context.Context, input *DeleteTicketLinkInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Tracker.DeleteTicketLink(ctx, principal, input.TicketID, input.LinkID); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) assignSprint(ctx context.Context, input *AssignSprintInput) (*TicketOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	ticket, err := provider.Tracker.SetTicketSprint(ctx, principal, input.TicketID, input.Body.Spec.SprintID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &TicketOutput{Body: ResourceFromTracker(ticket)}, nil
}

func (provider Provider) removeSprint(ctx context.Context, input *TicketIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if _, err := provider.Tracker.SetTicketSprint(ctx, principal, input.TicketID, ""); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}
