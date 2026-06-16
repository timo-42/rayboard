package tickets

import (
	"time"

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
	Body     shared.ResourceInput[TicketUpdateSpec]
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
	Body TicketResource
}

type ActivityOutput = shared.ListOutput[tracker.TicketActivity]

type TicketMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TicketSpec struct {
	Title          string         `json:"title,omitempty"`
	Description    string         `json:"description,omitempty"`
	Status         string         `json:"status,omitempty"`
	Priority       string         `json:"priority,omitempty"`
	Type           string         `json:"type,omitempty"`
	AssigneeID     string         `json:"assignee_id,omitempty"`
	ParentTicketID string         `json:"parent_ticket_id,omitempty"`
	SprintID       string         `json:"sprint_id,omitempty"`
	ComponentID    string         `json:"component_id,omitempty"`
	VersionID      string         `json:"version_id,omitempty"`
	Rank           string         `json:"rank,omitempty"`
	StartDate      string         `json:"start_date,omitempty"`
	DueDate        string         `json:"due_date,omitempty"`
	Labels         []string       `json:"labels,omitempty"`
	CustomFields   map[string]any `json:"custom_fields,omitempty"`
}

type TicketUpdateSpec struct {
	Title          *string         `json:"title,omitempty"`
	Description    *string         `json:"description,omitempty"`
	Status         *string         `json:"status,omitempty"`
	Priority       *string         `json:"priority,omitempty"`
	Type           *string         `json:"type,omitempty"`
	AssigneeID     *string         `json:"assignee_id,omitempty"`
	ParentTicketID *string         `json:"parent_ticket_id,omitempty"`
	SprintID       *string         `json:"sprint_id,omitempty"`
	ComponentID    *string         `json:"component_id,omitempty"`
	VersionID      *string         `json:"version_id,omitempty"`
	Rank           *string         `json:"rank,omitempty"`
	StartDate      *string         `json:"start_date,omitempty"`
	DueDate        *string         `json:"due_date,omitempty"`
	Labels         *[]string       `json:"labels,omitempty"`
	CustomFields   *map[string]any `json:"custom_fields,omitempty"`
}

type TicketStatus struct {
	Key        string     `json:"key"`
	ReporterID string     `json:"reporter_id,omitempty"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type TicketResource = shared.Resource[TicketMetadata, TicketSpec, TicketStatus]

func (spec TicketSpec) ToCreateInput(projectID string) tracker.CreateTicketInput {
	return tracker.CreateTicketInput{
		ProjectID:      projectID,
		Title:          spec.Title,
		Description:    spec.Description,
		Status:         spec.Status,
		Priority:       spec.Priority,
		Type:           spec.Type,
		AssigneeID:     spec.AssigneeID,
		ParentTicketID: spec.ParentTicketID,
		SprintID:       spec.SprintID,
		ComponentID:    spec.ComponentID,
		VersionID:      spec.VersionID,
		Rank:           spec.Rank,
		StartDate:      spec.StartDate,
		DueDate:        spec.DueDate,
		Labels:         spec.Labels,
		CustomFields:   spec.CustomFields,
	}
}

func (spec TicketUpdateSpec) ToUpdateInput() tracker.UpdateTicketInput {
	return tracker.UpdateTicketInput{
		Title:          spec.Title,
		Description:    spec.Description,
		Status:         spec.Status,
		Priority:       spec.Priority,
		Type:           spec.Type,
		AssigneeID:     spec.AssigneeID,
		ParentTicketID: spec.ParentTicketID,
		SprintID:       spec.SprintID,
		ComponentID:    spec.ComponentID,
		VersionID:      spec.VersionID,
		Rank:           spec.Rank,
		StartDate:      spec.StartDate,
		DueDate:        spec.DueDate,
		Labels:         spec.Labels,
		CustomFields:   spec.CustomFields,
	}
}

func ResourceFromTracker(ticket tracker.Ticket) TicketResource {
	return TicketResource{
		Metadata: TicketMetadata{
			ID:        ticket.ID,
			ProjectID: ticket.ProjectID,
			CreatedAt: ticket.CreatedAt,
			UpdatedAt: ticket.UpdatedAt,
		},
		Spec: TicketSpec{
			Title:          ticket.Title,
			Description:    ticket.Description,
			Status:         ticket.Status,
			Priority:       ticket.Priority,
			Type:           ticket.Type,
			AssigneeID:     ticket.AssigneeID,
			ParentTicketID: ticket.ParentTicketID,
			SprintID:       ticket.SprintID,
			ComponentID:    ticket.ComponentID,
			VersionID:      ticket.VersionID,
			Rank:           ticket.Rank,
			StartDate:      ticket.StartDate,
			DueDate:        ticket.DueDate,
			Labels:         ticket.Labels,
			CustomFields:   ticket.CustomFields,
		},
		Status: TicketStatus{
			Key:        ticket.Key,
			ReporterID: ticket.ReporterID,
			DeletedAt:  ticket.DeletedAt,
		},
	}
}

func ResourcesFromTracker(tickets []tracker.Ticket) []TicketResource {
	resources := make([]TicketResource, 0, len(tickets))
	for _, ticket := range tickets {
		resources = append(resources, ResourceFromTracker(ticket))
	}
	return resources
}
