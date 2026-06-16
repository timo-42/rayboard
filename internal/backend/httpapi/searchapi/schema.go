package searchapi

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	ticketapi "github.com/timo-42/rayboard/internal/backend/httpapi/tickets"
	"github.com/timo-42/rayboard/internal/backend/search"
)

type SearchTicketsInput struct {
	shared.AuthInput
	Body shared.ResourceInput[SearchTicketsSpec]
}

type SearchTicketsOutput struct {
	Body SearchTicketsResultResource
}

type ListSavedViewsInput struct {
	shared.AuthInput
	ProjectID string `query:"project_id" doc:"Filter saved views by project ID."`
	Pinned    bool   `query:"pinned" doc:"Only include pinned saved views."`
	Limit     int    `query:"limit" doc:"Maximum number of saved views to return."`
	Offset    int    `query:"offset" doc:"Number of saved views to skip."`
}

type CreateSavedViewInput struct {
	shared.AuthInput
	Body shared.ResourceInput[SavedViewSpec]
}

type SavedViewIDInput struct {
	shared.AuthInput
	ViewID string `path:"view_id" doc:"Saved view ID."`
}

type UpdateSavedViewInput struct {
	shared.AuthInput
	ViewID string `path:"view_id" doc:"Saved view ID."`
	Body   shared.ResourceInput[SavedViewUpdateSpec]
}

type SavedViewOutput struct {
	Body SavedViewResource
}

type ListSavedViewsOutput = shared.ListOutput[SavedViewResource]
type CreateSavedViewOutput = shared.CreatedOutput[SavedViewResource]

type SearchTicketsSpec struct {
	ProjectID string            `json:"project_id,omitempty"`
	Filter    string            `json:"filter,omitempty"`
	Text      string            `json:"text,omitempty"`
	Sort      []search.SortSpec `json:"sort,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Cursor    string            `json:"cursor,omitempty"`
}

type SearchTicketsResultMetadata struct {
	GeneratedAt time.Time `json:"generated_at"`
}

type SearchTicketsResultStatus struct {
	Items      []ticketapi.TicketResource `json:"items"`
	NextCursor string                     `json:"next_cursor,omitempty"`
}

type SearchTicketsResultResource = shared.Resource[SearchTicketsResultMetadata, SearchTicketsSpec, SearchTicketsResultStatus]

type SavedViewMetadata struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SavedViewSpec struct {
	OwnerUserID string                `json:"owner_user_id,omitempty"`
	ProjectID   string                `json:"project_id,omitempty"`
	ScopeType   string                `json:"scope_type,omitempty"`
	Name        string                `json:"name,omitempty"`
	Query       search.SavedViewQuery `json:"query,omitempty"`
	Sort        []search.SortSpec     `json:"sort,omitempty"`
	Columns     []string              `json:"columns,omitempty"`
	DisplayMode string                `json:"display_mode,omitempty"`
	GroupBy     string                `json:"group_by,omitempty"`
	Pinned      bool                  `json:"pinned,omitempty"`
}

type SavedViewUpdateSpec struct {
	Name        *string                `json:"name,omitempty"`
	Query       *search.SavedViewQuery `json:"query,omitempty"`
	Sort        *[]search.SortSpec     `json:"sort,omitempty"`
	Columns     *[]string              `json:"columns,omitempty"`
	DisplayMode *string                `json:"display_mode,omitempty"`
	GroupBy     *string                `json:"group_by,omitempty"`
	Pinned      *bool                  `json:"pinned,omitempty"`
}

type SavedViewStatus struct{}

type SavedViewResource = shared.Resource[SavedViewMetadata, SavedViewSpec, SavedViewStatus]

func (spec SearchTicketsSpec) serviceInput() search.SearchTicketsInput {
	return search.SearchTicketsInput{
		ProjectID: spec.ProjectID,
		Filter:    spec.Filter,
		Text:      spec.Text,
		Sort:      spec.Sort,
		Limit:     spec.Limit,
		Cursor:    spec.Cursor,
	}
}

func (spec SavedViewSpec) createInput() search.CreateSavedViewInput {
	return search.CreateSavedViewInput{
		OwnerUserID: spec.OwnerUserID,
		ProjectID:   spec.ProjectID,
		ScopeType:   spec.ScopeType,
		Name:        spec.Name,
		Query:       spec.Query,
		Sort:        spec.Sort,
		Columns:     spec.Columns,
		DisplayMode: spec.DisplayMode,
		GroupBy:     spec.GroupBy,
		Pinned:      spec.Pinned,
	}
}

func (spec SavedViewUpdateSpec) updateInput() search.UpdateSavedViewInput {
	return search.UpdateSavedViewInput{
		Name:        spec.Name,
		Query:       spec.Query,
		Sort:        spec.Sort,
		Columns:     spec.Columns,
		DisplayMode: spec.DisplayMode,
		GroupBy:     spec.GroupBy,
		Pinned:      spec.Pinned,
	}
}

func searchTicketsResultResource(spec SearchTicketsSpec, result search.SearchTicketsResult) SearchTicketsResultResource {
	return SearchTicketsResultResource{
		Metadata: SearchTicketsResultMetadata{
			GeneratedAt: time.Now().UTC(),
		},
		Spec: spec,
		Status: SearchTicketsResultStatus{
			Items:      ticketapi.ResourcesFromTracker(result.Tickets),
			NextCursor: result.NextCursor,
		},
	}
}

func savedViewResource(view search.SavedView) SavedViewResource {
	return SavedViewResource{
		Metadata: SavedViewMetadata{
			ID:        view.ID,
			CreatedAt: view.CreatedAt,
			UpdatedAt: view.UpdatedAt,
		},
		Spec: SavedViewSpec{
			OwnerUserID: view.OwnerUserID,
			ProjectID:   view.ProjectID,
			ScopeType:   view.ScopeType,
			Name:        view.Name,
			Query:       view.Query,
			Sort:        view.Sort,
			Columns:     view.Columns,
			DisplayMode: view.DisplayMode,
			GroupBy:     view.GroupBy,
			Pinned:      view.Pinned,
		},
		Status: SavedViewStatus{},
	}
}

func savedViewResources(views []search.SavedView) []SavedViewResource {
	resources := make([]SavedViewResource, 0, len(views))
	for _, view := range views {
		resources = append(resources, savedViewResource(view))
	}
	return resources
}
