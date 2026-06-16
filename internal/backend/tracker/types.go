package tracker

import "time"

type Project struct {
	ID          string     `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	LeadUserID  string     `json:"lead_user_id,omitempty"`
	CreatedBy   string     `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateProjectInput struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LeadUserID  string `json:"lead_user_id"`
}

type ListProjectsInput struct {
	IncludeArchived bool `json:"include_archived"`
	Limit           int  `json:"limit"`
	Offset          int  `json:"offset"`
}

type Ticket struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"project_id"`
	Key            string     `json:"key"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority,omitempty"`
	Type           string     `json:"type,omitempty"`
	ReporterID     string     `json:"reporter_id,omitempty"`
	AssigneeID     string     `json:"assignee_id,omitempty"`
	ParentTicketID string     `json:"parent_ticket_id,omitempty"`
	Rank           string     `json:"rank,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

type CreateTicketInput struct {
	ProjectID      string `json:"project_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	Priority       string `json:"priority"`
	Type           string `json:"type"`
	ReporterID     string `json:"reporter_id"`
	AssigneeID     string `json:"assignee_id"`
	ParentTicketID string `json:"parent_ticket_id"`
	Rank           string `json:"rank"`
}

type ListTicketsInput struct {
	ProjectID  string `json:"project_id"`
	Status     string `json:"status"`
	AssigneeID string `json:"assignee_id"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

type UpdateTicketInput struct {
	Title          *string `json:"title,omitempty"`
	Description    *string `json:"description,omitempty"`
	Status         *string `json:"status,omitempty"`
	Priority       *string `json:"priority,omitempty"`
	Type           *string `json:"type,omitempty"`
	AssigneeID     *string `json:"assignee_id,omitempty"`
	ParentTicketID *string `json:"parent_ticket_id,omitempty"`
	Rank           *string `json:"rank,omitempty"`
}

type TicketActivity struct {
	ID           string         `json:"id"`
	TicketID     string         `json:"ticket_id"`
	ActorID      string         `json:"actor_id,omitempty"`
	ActivityType string         `json:"activity_type"`
	Data         map[string]any `json:"data"`
	CreatedAt    time.Time      `json:"created_at"`
}
