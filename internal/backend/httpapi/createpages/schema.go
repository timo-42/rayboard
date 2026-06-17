package createpages

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	ticketapi "github.com/timo-42/rayboard/internal/backend/httpapi/tickets"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type ProjectPagesInput struct {
	shared.AuthInput
	ProjectID       string `path:"project_id" doc:"Project ID."`
	IncludeDisabled bool   `query:"include_disabled" doc:"Include disabled pages."`
	Limit           int    `query:"limit" doc:"Maximum number of pages to return."`
	Offset          int    `query:"offset" doc:"Number of pages to skip."`
}

type CreateProjectPageInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
	Body      shared.ResourceInput[PageSpec]
}

type PageIDInput struct {
	shared.AuthInput
	PageID string `path:"page_id" doc:"Ticket create page ID."`
}

type UpdatePageInput struct {
	shared.AuthInput
	PageID string `path:"page_id" doc:"Ticket create page ID."`
	Body   shared.ResourceInput[UpdatePageSpec]
}

type ResolvePageInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
	Slug      string `path:"slug" doc:"Ticket create page slug."`
}

type SubmitPageInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
	Slug      string `path:"slug" doc:"Ticket create page slug."`
	Body      shared.ResourceInput[SubmitPageSpec]
}

type PageMetadata struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	OwnerUserID string    `json:"owner_user_id,omitempty"`
	CreatedBy   string    `json:"created_by,omitempty"`
	UpdatedBy   string    `json:"updated_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PageSpec struct {
	Name          string           `json:"name,omitempty"`
	Slug          string           `json:"slug,omitempty"`
	Description   string           `json:"description,omitempty"`
	Enabled       bool             `json:"enabled,omitempty"`
	TargetType    string           `json:"target_type,omitempty"`
	TargetStatus  string           `json:"target_status,omitempty"`
	FieldLayout   []map[string]any `json:"field_layout,omitempty"`
	Defaults      map[string]any   `json:"defaults,omitempty"`
	FormLuaScript string           `json:"form_lua_script,omitempty"`
	OwnerUserID   string           `json:"owner_user_id,omitempty"`
}

type UpdatePageSpec struct {
	Name          *string           `json:"name,omitempty"`
	Slug          *string           `json:"slug,omitempty"`
	Description   *string           `json:"description,omitempty"`
	Enabled       *bool             `json:"enabled,omitempty"`
	TargetType    *string           `json:"target_type,omitempty"`
	TargetStatus  *string           `json:"target_status,omitempty"`
	FieldLayout   *[]map[string]any `json:"field_layout,omitempty"`
	Defaults      *map[string]any   `json:"defaults,omitempty"`
	FormLuaScript *string           `json:"form_lua_script,omitempty"`
	OwnerUserID   *string           `json:"owner_user_id,omitempty"`
}

type PageStatus struct {
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type SchemaMetadata struct {
	PageID    string `json:"page_id"`
	ProjectID string `json:"project_id"`
	Slug      string `json:"slug"`
}

type SchemaStatus struct {
	Enabled bool `json:"enabled"`
}

type SubmitPageSpec struct {
	Ticket ticketapi.TicketSpec `json:"ticket"`
}

type PageResource = shared.Resource[PageMetadata, PageSpec, PageStatus]
type SchemaResource = shared.Resource[SchemaMetadata, PageSpec, SchemaStatus]

type ListPagesOutput = shared.ListOutput[PageResource]
type CreatePageOutput = shared.CreatedOutput[PageResource]
type SubmitPageOutput = shared.CreatedOutput[ticketapi.TicketResource]

type PageOutput struct {
	Body PageResource
}

type SchemaOutput struct {
	Body SchemaResource
}

func (spec PageSpec) createInput(projectID string) tracker.CreateCreatePageInput {
	return tracker.CreateCreatePageInput{
		ProjectID:     projectID,
		Name:          spec.Name,
		Slug:          spec.Slug,
		Description:   spec.Description,
		Enabled:       spec.Enabled,
		TargetType:    spec.TargetType,
		TargetStatus:  spec.TargetStatus,
		FieldLayout:   spec.FieldLayout,
		Defaults:      spec.Defaults,
		FormLuaScript: spec.FormLuaScript,
		OwnerUserID:   spec.OwnerUserID,
	}
}

func (spec UpdatePageSpec) updateInput() tracker.UpdateCreatePageInput {
	return tracker.UpdateCreatePageInput{
		Name:          spec.Name,
		Slug:          spec.Slug,
		Description:   spec.Description,
		Enabled:       spec.Enabled,
		TargetType:    spec.TargetType,
		TargetStatus:  spec.TargetStatus,
		FieldLayout:   spec.FieldLayout,
		Defaults:      spec.Defaults,
		FormLuaScript: spec.FormLuaScript,
		OwnerUserID:   spec.OwnerUserID,
	}
}

func (spec SubmitPageSpec) submitInput() tracker.SubmitCreatePageInput {
	return tracker.SubmitCreatePageInput{Ticket: spec.Ticket.ToCreateInput("")}
}

func pageResource(page tracker.CreatePage) PageResource {
	return PageResource{
		Metadata: PageMetadata{
			ID:          page.ID,
			ProjectID:   page.ProjectID,
			OwnerUserID: page.OwnerUserID,
			CreatedBy:   page.CreatedBy,
			UpdatedBy:   page.UpdatedBy,
			CreatedAt:   page.CreatedAt,
			UpdatedAt:   page.UpdatedAt,
		},
		Spec: PageSpec{
			Name:          page.Name,
			Slug:          page.Slug,
			Description:   page.Description,
			Enabled:       page.Enabled,
			TargetType:    page.TargetType,
			TargetStatus:  page.TargetStatus,
			FieldLayout:   page.FieldLayout,
			Defaults:      page.Defaults,
			FormLuaScript: page.FormLuaScript,
			OwnerUserID:   page.OwnerUserID,
		},
		Status: PageStatus{DeletedAt: page.DeletedAt},
	}
}

func schemaResource(page tracker.CreatePage) SchemaResource {
	resource := pageResource(page)
	resource.Spec.FormLuaScript = ""
	return SchemaResource{
		Metadata: SchemaMetadata{
			PageID:    page.ID,
			ProjectID: page.ProjectID,
			Slug:      page.Slug,
		},
		Spec:   resource.Spec,
		Status: SchemaStatus{Enabled: page.Enabled},
	}
}

func pageResources(pages []tracker.CreatePage) []PageResource {
	resources := make([]PageResource, 0, len(pages))
	for _, page := range pages {
		resources = append(resources, pageResource(page))
	}
	return resources
}
