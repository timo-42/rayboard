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
	huma.Register(api, shared.Operation(http.MethodGet, "/api/tickets/{ticket_id}/activity", "Tickets", "List ticket activity"), provider.listActivity)
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

func (provider Provider) listActivity(ctx context.Context, input *TicketIDInput) (*ActivityOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Tracker.ListTicketActivity(ctx, principal, input.TicketID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &ActivityOutput{Body: shared.ItemList[ActivityResource]{Items: ActivityResourcesFromTracker(items)}}, nil
}

func (provider Provider) assignSprint(ctx context.Context, input *AssignSprintInput) (*TicketOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	ticket, err := provider.Tracker.SetTicketSprint(ctx, principal, input.TicketID, input.Body.SprintID)
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
