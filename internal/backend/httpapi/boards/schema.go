package boards

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	ticketapi "github.com/timo-42/rayboard/internal/backend/httpapi/tickets"
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
	Body BoardTicketsResource
}

type BoardTicketsResource struct {
	Board   tracker.Board                `json:"board"`
	Columns []BoardTicketsColumnResource `json:"columns"`
}

type BoardTicketsColumnResource struct {
	Column  tracker.BoardColumn        `json:"column"`
	Tickets []ticketapi.TicketResource `json:"tickets"`
}

func BoardTicketsResourceFromTracker(boardTickets tracker.BoardTickets) BoardTicketsResource {
	columns := make([]BoardTicketsColumnResource, 0, len(boardTickets.Columns))
	for _, column := range boardTickets.Columns {
		columns = append(columns, BoardTicketsColumnResource{
			Column:  column.Column,
			Tickets: ticketapi.ResourcesFromTracker(column.Tickets),
		})
	}
	return BoardTicketsResource{
		Board:   boardTickets.Board,
		Columns: columns,
	}
}
