package search

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/tracker"
)

const (
	SavedViewScopeUser    = "user"
	SavedViewScopeProject = "project"
	SavedViewScopeGlobal  = "global"

	SortDirectionAsc  = "asc"
	SortDirectionDesc = "desc"
)

type Ticket = tracker.Ticket

type SavedView struct {
	ID          string         `json:"id"`
	OwnerUserID string         `json:"owner_user_id,omitempty"`
	ProjectID   string         `json:"project_id,omitempty"`
	ScopeType   string         `json:"scope_type"`
	Name        string         `json:"name"`
	Query       SavedViewQuery `json:"query"`
	Sort        []SortSpec     `json:"sort"`
	Columns     []string       `json:"columns"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type SavedViewQuery struct {
	Filter string `json:"filter,omitempty"`
	Text   string `json:"text,omitempty"`
}

type SortSpec struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type CreateSavedViewInput struct {
	OwnerUserID string         `json:"owner_user_id,omitempty"`
	ProjectID   string         `json:"project_id,omitempty"`
	ScopeType   string         `json:"scope_type"`
	Name        string         `json:"name"`
	Query       SavedViewQuery `json:"query"`
	Sort        []SortSpec     `json:"sort"`
	Columns     []string       `json:"columns"`
}

type UpdateSavedViewInput struct {
	Name    *string         `json:"name,omitempty"`
	Query   *SavedViewQuery `json:"query,omitempty"`
	Sort    *[]SortSpec     `json:"sort,omitempty"`
	Columns *[]string       `json:"columns,omitempty"`
}

type ListSavedViewsInput struct {
	ProjectID string `json:"project_id,omitempty"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

type SearchTicketsInput struct {
	ProjectID string     `json:"project_id,omitempty"`
	Filter    string     `json:"filter,omitempty"`
	Text      string     `json:"text,omitempty"`
	Sort      []SortSpec `json:"sort"`
	Limit     int        `json:"limit"`
	Cursor    string     `json:"cursor,omitempty"`
}

type SearchTicketsResult struct {
	Tickets    []Ticket `json:"items"`
	NextCursor string   `json:"next_cursor,omitempty"`
}
