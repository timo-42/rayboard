package projects

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type ListProjectsInput struct {
	shared.AuthInput
	IncludeArchived bool `query:"include_archived"`
	Limit           int  `query:"limit"`
	Offset          int  `query:"offset"`
}

type CreateProjectInput struct {
	shared.AuthInput
	Body tracker.CreateProjectInput
}

type ProjectIDInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
}

type ReorderBacklogInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      tracker.ReorderBacklogInput
}

type ReplaceStatusesInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      tracker.ReplaceProjectStatusesInput
}

type ListTicketsInput struct {
	shared.AuthInput
	ProjectID   string `path:"project_id"`
	Status      string `query:"status"`
	AssigneeID  string `query:"assignee_id"`
	SprintID    string `query:"sprint_id"`
	ComponentID string `query:"component_id"`
	VersionID   string `query:"version_id"`
	Label       string `query:"label"`
	Limit       int    `query:"limit"`
	Offset      int    `query:"offset"`
}

type CreateTicketInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      tracker.CreateTicketInput
}

type CreateBoardInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      tracker.CreateBoardInput
}

type CreateComponentInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      tracker.CreateComponentInput
}

type CreateVersionInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      tracker.CreateVersionInput
}

type ListVersionsInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Status    string `query:"status"`
}

type CreateCustomFieldInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      tracker.CreateCustomFieldInput
}

type ListSprintsInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	State     string `query:"state"`
}

type CreateSprintInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      tracker.CreateSprintInput
}

type ListProjectsOutput = shared.ListOutput[tracker.Project]
type CreateProjectOutput = shared.CreatedOutput[tracker.Project]
type ProjectOutput struct {
	Body tracker.Project
}
type ListTicketsOutput = shared.ListOutput[tracker.Ticket]
type CreateTicketOutput = shared.CreatedOutput[tracker.Ticket]
type ListStatusesOutput = shared.ListOutput[tracker.ProjectStatus]
type ListBoardsOutput = shared.ListOutput[tracker.Board]
type CreateBoardOutput = shared.CreatedOutput[tracker.Board]
type ListComponentsOutput = shared.ListOutput[tracker.Component]
type CreateComponentOutput = shared.CreatedOutput[tracker.Component]
type ListVersionsOutput = shared.ListOutput[tracker.Version]
type CreateVersionOutput = shared.CreatedOutput[tracker.Version]
type ListCustomFieldsOutput = shared.ListOutput[tracker.CustomFieldDefinition]
type CreateCustomFieldOutput = shared.CreatedOutput[tracker.CustomFieldDefinition]
type ListRoadmapOutput = shared.ListOutput[tracker.RoadmapItem]
type ListSprintsOutput = shared.ListOutput[tracker.Sprint]
type CreateSprintOutput = shared.CreatedOutput[tracker.Sprint]
