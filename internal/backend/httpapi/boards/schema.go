package boards

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type BoardIDInput struct {
	shared.AuthInput
	BoardID string `path:"board_id"`
}

type UpdateBoardInput struct {
	shared.AuthInput
	BoardID string `path:"board_id"`
	Body    tracker.UpdateBoardInput
}

type BoardOutput struct {
	Body tracker.Board
}

type BoardTicketsOutput struct {
	Body tracker.BoardTickets
}
