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

type ProjectStatus struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReplaceProjectStatusesInput struct {
	Statuses []ProjectStatusInput `json:"statuses"`
}

type ProjectStatusInput struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Board struct {
	ID          string        `json:"id"`
	ProjectID   string        `json:"project_id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	CreatedBy   string        `json:"created_by,omitempty"`
	Columns     []BoardColumn `json:"columns,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type BoardColumn struct {
	ID         string    `json:"id"`
	BoardID    string    `json:"board_id"`
	StatusSlug string    `json:"status_slug"`
	Name       string    `json:"name"`
	Position   int       `json:"position"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateBoardInput struct {
	ProjectID   string   `json:"project_id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	StatusSlugs []string `json:"status_slugs,omitempty"`
}

type UpdateBoardInput struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	StatusSlugs *[]string `json:"status_slugs,omitempty"`
}

type BoardTickets struct {
	Board   Board                `json:"board"`
	Columns []BoardTicketsColumn `json:"columns"`
}

type BoardTicketsColumn struct {
	Column  BoardColumn `json:"column"`
	Tickets []Ticket    `json:"tickets"`
}

type CreateProjectInput struct {
	Key         string `json:"key,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	LeadUserID  string `json:"lead_user_id,omitempty"`
}

type ListProjectsInput struct {
	IncludeArchived bool `json:"include_archived"`
	Limit           int  `json:"limit"`
	Offset          int  `json:"offset"`
}

type Ticket struct {
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	Key            string         `json:"key"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Status         string         `json:"status"`
	Priority       string         `json:"priority,omitempty"`
	Type           string         `json:"type,omitempty"`
	ReporterID     string         `json:"reporter_id,omitempty"`
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
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      *time.Time     `json:"deleted_at,omitempty"`
}

type CreateTicketInput struct {
	ProjectID      string         `json:"project_id,omitempty"`
	Title          string         `json:"title,omitempty"`
	Description    string         `json:"description,omitempty"`
	Status         string         `json:"status,omitempty"`
	Priority       string         `json:"priority,omitempty"`
	Type           string         `json:"type,omitempty"`
	ReporterID     string         `json:"reporter_id,omitempty"`
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

type ListTicketsInput struct {
	ProjectID   string `json:"project_id"`
	Status      string `json:"status"`
	AssigneeID  string `json:"assignee_id"`
	SprintID    string `json:"sprint_id"`
	ComponentID string `json:"component_id"`
	VersionID   string `json:"version_id"`
	Label       string `json:"label"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

type UpdateTicketInput struct {
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

type RoadmapItem struct {
	Epic     Ticket          `json:"epic"`
	Progress RoadmapProgress `json:"progress"`
}

type RoadmapProgress struct {
	Total    int            `json:"total"`
	Done     int            `json:"done"`
	ByStatus map[string]int `json:"by_status"`
}

type TicketActivity struct {
	ID           string         `json:"id"`
	TicketID     string         `json:"ticket_id"`
	ActorID      string         `json:"actor_id,omitempty"`
	ActivityType string         `json:"activity_type"`
	Data         map[string]any `json:"data"`
	CreatedAt    time.Time      `json:"created_at"`
}

type Sprint struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	Name        string     `json:"name"`
	Goal        string     `json:"goal,omitempty"`
	State       string     `json:"state"`
	StartDate   string     `json:"start_date,omitempty"`
	EndDate     string     `json:"end_date,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ProjectLabel struct {
	ProjectID   string `json:"project_id"`
	Label       string `json:"label"`
	TicketCount int    `json:"ticket_count"`
}

type Component struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	OwnerUserID       string    `json:"owner_user_id,omitempty"`
	DefaultAssigneeID string    `json:"default_assignee_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CreateComponentInput struct {
	ProjectID         string `json:"project_id,omitempty"`
	Name              string `json:"name,omitempty"`
	Description       string `json:"description,omitempty"`
	OwnerUserID       string `json:"owner_user_id,omitempty"`
	DefaultAssigneeID string `json:"default_assignee_id,omitempty"`
}

type UpdateComponentInput struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	OwnerUserID       *string `json:"owner_user_id,omitempty"`
	DefaultAssigneeID *string `json:"default_assignee_id,omitempty"`
}

type Version struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	TargetDate  string    `json:"target_date,omitempty"`
	ReleaseDate string    `json:"release_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateVersionInput struct {
	ProjectID   string `json:"project_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	TargetDate  string `json:"target_date,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
}

type UpdateVersionInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	TargetDate  *string `json:"target_date,omitempty"`
	ReleaseDate *string `json:"release_date,omitempty"`
}

const (
	CustomFieldTypeText         = "text"
	CustomFieldTypeNumber       = "number"
	CustomFieldTypeBoolean      = "boolean"
	CustomFieldTypeDate         = "date"
	CustomFieldTypeSingleSelect = "single_select"
	CustomFieldTypeMultiSelect  = "multi_select"
	CustomFieldTypeUser         = "user"
)

type CustomFieldDefinition struct {
	ID        string              `json:"id"`
	ProjectID string              `json:"project_id"`
	Key       string              `json:"key"`
	Name      string              `json:"name"`
	FieldType string              `json:"field_type"`
	Required  bool                `json:"required"`
	Options   []CustomFieldOption `json:"options,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type CustomFieldOption struct {
	ID        string    `json:"id"`
	FieldID   string    `json:"field_id"`
	Value     string    `json:"value"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateCustomFieldInput struct {
	ProjectID string   `json:"project_id,omitempty"`
	Key       string   `json:"key,omitempty"`
	Name      string   `json:"name,omitempty"`
	FieldType string   `json:"field_type,omitempty"`
	Required  bool     `json:"required,omitempty"`
	Options   []string `json:"options,omitempty"`
}

type UpdateCustomFieldInput struct {
	Key       *string   `json:"key,omitempty"`
	Name      *string   `json:"name,omitempty"`
	FieldType *string   `json:"field_type,omitempty"`
	Required  *bool     `json:"required,omitempty"`
	Options   *[]string `json:"options,omitempty"`
}

type CreateSprintInput struct {
	ProjectID string `json:"project_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Goal      string `json:"goal,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

type UpdateSprintInput struct {
	Name      *string `json:"name,omitempty"`
	Goal      *string `json:"goal,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
}
