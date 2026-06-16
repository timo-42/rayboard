package projects

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	sprintapi "github.com/timo-42/rayboard/internal/backend/httpapi/sprints"
	ticketapi "github.com/timo-42/rayboard/internal/backend/httpapi/tickets"
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
	Body shared.ResourceInput[ProjectSpec]
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
	Body      shared.ResourceInput[ticketapi.TicketSpec]
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
	Body      shared.ResourceInput[sprintapi.SprintSpec]
}

type ProjectMetadata struct {
	ID        string    `json:"id"`
	CreatedBy string    `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectSpec struct {
	Key         string `json:"key,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	LeadUserID  string `json:"lead_user_id,omitempty"`
}

type ProjectResourceStatus struct {
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type ProjectResource = shared.Resource[ProjectMetadata, ProjectSpec, ProjectResourceStatus]

type ListProjectsOutput = shared.ListOutput[ProjectResource]
type CreateProjectOutput = shared.CreatedOutput[ProjectResource]
type ProjectOutput struct {
	Body ProjectResource
}
type ListTicketsOutput = shared.ListOutput[ticketapi.TicketResource]
type CreateTicketOutput = shared.CreatedOutput[ticketapi.TicketResource]
type ListStatusesOutput = shared.ListOutput[tracker.ProjectStatus]
type ListBoardsOutput = shared.ListOutput[tracker.Board]
type CreateBoardOutput = shared.CreatedOutput[tracker.Board]
type ListComponentsOutput = shared.ListOutput[tracker.Component]
type CreateComponentOutput = shared.CreatedOutput[tracker.Component]
type ListVersionsOutput = shared.ListOutput[tracker.Version]
type CreateVersionOutput = shared.CreatedOutput[tracker.Version]
type ListCustomFieldsOutput = shared.ListOutput[tracker.CustomFieldDefinition]
type CreateCustomFieldOutput = shared.CreatedOutput[tracker.CustomFieldDefinition]
type RoadmapItemResource struct {
	Epic     ticketapi.TicketResource `json:"epic"`
	Progress tracker.RoadmapProgress  `json:"progress"`
}

type ListRoadmapOutput = shared.ListOutput[RoadmapItemResource]
type ListSprintsOutput = shared.ListOutput[sprintapi.SprintResource]
type CreateSprintOutput = shared.CreatedOutput[sprintapi.SprintResource]

func (spec ProjectSpec) ToCreateInput() tracker.CreateProjectInput {
	return tracker.CreateProjectInput{
		Key:         spec.Key,
		Name:        spec.Name,
		Description: spec.Description,
		LeadUserID:  spec.LeadUserID,
	}
}

func projectResource(project tracker.Project) ProjectResource {
	return ProjectResource{
		Metadata: ProjectMetadata{
			ID:        project.ID,
			CreatedBy: project.CreatedBy,
			CreatedAt: project.CreatedAt,
			UpdatedAt: project.UpdatedAt,
		},
		Spec: ProjectSpec{
			Key:         project.Key,
			Name:        project.Name,
			Description: project.Description,
			LeadUserID:  project.LeadUserID,
		},
		Status: ProjectResourceStatus{
			ArchivedAt: project.ArchivedAt,
			DeletedAt:  project.DeletedAt,
		},
	}
}

func projectResources(projects []tracker.Project) []ProjectResource {
	resources := make([]ProjectResource, 0, len(projects))
	for _, project := range projects {
		resources = append(resources, projectResource(project))
	}
	return resources
}

func roadmapItemResources(items []tracker.RoadmapItem) []RoadmapItemResource {
	resources := make([]RoadmapItemResource, 0, len(items))
	for _, item := range items {
		resources = append(resources, RoadmapItemResource{
			Epic:     ticketapi.ResourceFromTracker(item.Epic),
			Progress: item.Progress,
		})
	}
	return resources
}
