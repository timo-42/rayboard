package tickets

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type TicketIDInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id"`
}

type UpdateTicketInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id"`
	Body     tracker.UpdateTicketInput
}

type AssignSprintInput struct {
	shared.AuthInput
	TicketID string `path:"ticket_id"`
	Body     AssignSprintInputBody
}

type AssignSprintInputBody struct {
	SprintID string `json:"sprint_id"`
}

type TicketOutput struct {
	Body tracker.Ticket
}

type ActivityOutput = shared.ListOutput[tracker.TicketActivity]
