package boards

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/boards/{board_id}", "Boards", "Get board"), provider.getBoard)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/boards/{board_id}", "Boards", "Update board"), provider.updateBoard)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/boards/{board_id}", "Boards", "Delete board", http.StatusNoContent), provider.deleteBoard)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/boards/{board_id}/tickets", "Boards", "List board tickets"), provider.listBoardTickets)
}

func (provider Provider) getBoard(ctx context.Context, input *BoardIDInput) (*BoardOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	board, err := provider.Tracker.GetBoard(ctx, principal, input.BoardID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &BoardOutput{Body: board}, nil
}

func (provider Provider) updateBoard(ctx context.Context, input *UpdateBoardInput) (*BoardOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	board, err := provider.Tracker.UpdateBoard(ctx, principal, input.BoardID, input.Body)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &BoardOutput{Body: board}, nil
}

func (provider Provider) deleteBoard(ctx context.Context, input *BoardIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Tracker.DeleteBoard(ctx, principal, input.BoardID); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) listBoardTickets(ctx context.Context, input *BoardIDInput) (*BoardTicketsOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	tickets, err := provider.Tracker.ListBoardTickets(ctx, principal, input.BoardID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &BoardTicketsOutput{Body: tickets}, nil
}
