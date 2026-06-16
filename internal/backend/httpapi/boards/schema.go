package boards

import (
	"time"

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
	Body    shared.ResourceInput[UpdateBoardSpec]
}

type BoardOutput struct {
	Body BoardResource
}

type BoardMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedBy string    `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BoardSpec struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	StatusSlugs []string `json:"status_slugs,omitempty"`
}

type UpdateBoardSpec struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	StatusSlugs *[]string `json:"status_slugs,omitempty"`
}

type BoardStatus struct {
	Columns []tracker.BoardColumn `json:"columns,omitempty"`
}

type BoardResource = shared.Resource[BoardMetadata, BoardSpec, BoardStatus]

type BoardTicketsOutput struct {
	Body BoardTicketsResource
}

type BoardTicketsMetadata struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
}

type BoardTicketsSpec struct {
	Board BoardResource `json:"board"`
}

type BoardTicketsStatus struct {
	Columns []BoardTicketsColumnResource `json:"columns"`
}

type BoardTicketsResource = shared.Resource[BoardTicketsMetadata, BoardTicketsSpec, BoardTicketsStatus]

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
		Metadata: BoardTicketsMetadata{
			ID:        boardTickets.Board.ID,
			ProjectID: boardTickets.Board.ProjectID,
		},
		Spec: BoardTicketsSpec{
			Board: ResourceFromTracker(boardTickets.Board),
		},
		Status: BoardTicketsStatus{
			Columns: columns,
		},
	}
}

func (spec BoardSpec) ToCreateInput(projectID string) tracker.CreateBoardInput {
	return tracker.CreateBoardInput{
		ProjectID:   projectID,
		Name:        spec.Name,
		Description: spec.Description,
		StatusSlugs: spec.StatusSlugs,
	}
}

func (spec UpdateBoardSpec) ToUpdateInput() tracker.UpdateBoardInput {
	return tracker.UpdateBoardInput{
		Name:        spec.Name,
		Description: spec.Description,
		StatusSlugs: spec.StatusSlugs,
	}
}

func ResourceFromTracker(board tracker.Board) BoardResource {
	statusSlugs := make([]string, 0, len(board.Columns))
	for _, column := range board.Columns {
		statusSlugs = append(statusSlugs, column.StatusSlug)
	}
	return BoardResource{
		Metadata: BoardMetadata{
			ID:        board.ID,
			ProjectID: board.ProjectID,
			CreatedBy: board.CreatedBy,
			CreatedAt: board.CreatedAt,
			UpdatedAt: board.UpdatedAt,
		},
		Spec: BoardSpec{
			Name:        board.Name,
			Description: board.Description,
			StatusSlugs: statusSlugs,
		},
		Status: BoardStatus{
			Columns: board.Columns,
		},
	}
}

func ResourcesFromTracker(boards []tracker.Board) []BoardResource {
	resources := make([]BoardResource, 0, len(boards))
	for _, board := range boards {
		resources = append(resources, ResourceFromTracker(board))
	}
	return resources
}
