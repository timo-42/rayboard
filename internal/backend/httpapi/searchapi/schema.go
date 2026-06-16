package searchapi

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/search"
)

type SearchTicketsInput struct {
	shared.AuthInput
	Body search.SearchTicketsInput
}

type SearchTicketsOutput struct {
	Body search.SearchTicketsResult
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
	Body search.CreateSavedViewInput
}

type SavedViewIDInput struct {
	shared.AuthInput
	ViewID string `path:"view_id" doc:"Saved view ID."`
}

type UpdateSavedViewInput struct {
	shared.AuthInput
	ViewID string `path:"view_id" doc:"Saved view ID."`
	Body   search.UpdateSavedViewInput
}

type SavedViewOutput struct {
	Body search.SavedView
}

type ListSavedViewsOutput = shared.ListOutput[search.SavedView]
type CreateSavedViewOutput = shared.CreatedOutput[search.SavedView]
