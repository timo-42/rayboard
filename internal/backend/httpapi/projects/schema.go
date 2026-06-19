package projects

import (
	"time"

	boardapi "github.com/timo-42/rayboard/internal/backend/httpapi/boards"
	componentapi "github.com/timo-42/rayboard/internal/backend/httpapi/components"
	fieldapi "github.com/timo-42/rayboard/internal/backend/httpapi/customfields"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	sprintapi "github.com/timo-42/rayboard/internal/backend/httpapi/sprints"
	ticketapi "github.com/timo-42/rayboard/internal/backend/httpapi/tickets"
	versionapi "github.com/timo-42/rayboard/internal/backend/httpapi/versions"
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
	Body      shared.ResourceInput[BacklogOrderSpec]
}

type ScheduleRoadmapInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      shared.ResourceInput[RoadmapScheduleSpec]
}

type ReplaceStatusesInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      shared.ResourceInput[ProjectStatusesSpec]
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
	Body      shared.ResourceInput[boardapi.BoardSpec]
}

type CreateComponentInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      shared.ResourceInput[componentapi.ComponentSpec]
}

type CreateVersionInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      shared.ResourceInput[versionapi.VersionSpec]
}

type ListVersionsInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Status    string `query:"status"`
}

type CreateCustomFieldInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id"`
	Body      shared.ResourceInput[fieldapi.FieldSpec]
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

type BacklogOrderSpec struct {
	TicketIDs []string `json:"ticket_ids,omitempty"`
}

func (spec BacklogOrderSpec) ToReorderInput() tracker.ReorderBacklogInput {
	return tracker.ReorderBacklogInput{TicketIDs: spec.TicketIDs}
}

type RoadmapScheduleSpec struct {
	TicketID  string `json:"ticket_id,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	DueDate   string `json:"due_date,omitempty"`
}

func (spec RoadmapScheduleSpec) ToScheduleInput() tracker.RoadmapScheduleInput {
	return tracker.RoadmapScheduleInput{
		TicketID:  spec.TicketID,
		StartDate: spec.StartDate,
		DueDate:   spec.DueDate,
	}
}

type ProjectStatusMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectStatusSpec struct {
	Slug     string `json:"slug,omitempty"`
	Name     string `json:"name,omitempty"`
	Position int    `json:"position,omitempty"`
}

type ProjectStatusResourceStatus struct {
}

type ProjectStatusResource = shared.Resource[ProjectStatusMetadata, ProjectStatusSpec, ProjectStatusResourceStatus]

type ProjectLabelMetadata struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
}

type ProjectLabelSpec struct {
	Label string `json:"label"`
}

type ProjectLabelStatus struct {
	TicketCount int `json:"ticket_count"`
}

type ProjectLabelResource = shared.Resource[ProjectLabelMetadata, ProjectLabelSpec, ProjectLabelStatus]

type ProjectStatusesSpec struct {
	Statuses []tracker.ProjectStatusInput `json:"statuses,omitempty"`
}

func (spec ProjectStatusesSpec) ToReplaceInput() tracker.ReplaceProjectStatusesInput {
	return tracker.ReplaceProjectStatusesInput{Statuses: spec.Statuses}
}

type ListProjectsOutput = shared.ListOutput[ProjectResource]
type CreateProjectOutput = shared.CreatedOutput[ProjectResource]
type ProjectOutput struct {
	Body ProjectResource
}
type ListTicketsOutput = shared.ListOutput[ticketapi.TicketResource]
type CreateTicketOutput = shared.CreatedOutput[ticketapi.TicketResource]
type ListStatusesOutput = shared.ListOutput[ProjectStatusResource]
type ListLabelsOutput = shared.ListOutput[ProjectLabelResource]
type ListBoardsOutput = shared.ListOutput[boardapi.BoardResource]
type CreateBoardOutput = shared.CreatedOutput[boardapi.BoardResource]
type ListComponentsOutput = shared.ListOutput[componentapi.ComponentResource]
type CreateComponentOutput = shared.CreatedOutput[componentapi.ComponentResource]
type ListVersionsOutput = shared.ListOutput[versionapi.VersionResource]
type CreateVersionOutput = shared.CreatedOutput[versionapi.VersionResource]
type ListCustomFieldsOutput = shared.ListOutput[fieldapi.FieldResource]
type CreateCustomFieldOutput = shared.CreatedOutput[fieldapi.FieldResource]
type RoadmapItemMetadata struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
}

type RoadmapItemSpec struct {
	Epic ticketapi.TicketResource `json:"epic"`
}

type RoadmapItemStatus struct {
	Progress tracker.RoadmapProgress `json:"progress"`
}

type RoadmapItemResource = shared.Resource[RoadmapItemMetadata, RoadmapItemSpec, RoadmapItemStatus]
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

func projectStatusResource(status tracker.ProjectStatus) ProjectStatusResource {
	return ProjectStatusResource{
		Metadata: ProjectStatusMetadata{
			ID:        status.ID,
			ProjectID: status.ProjectID,
			CreatedAt: status.CreatedAt,
			UpdatedAt: status.UpdatedAt,
		},
		Spec: ProjectStatusSpec{
			Slug:     status.Slug,
			Name:     status.Name,
			Position: status.Position,
		},
		Status: ProjectStatusResourceStatus{},
	}
}

func projectStatusResources(statuses []tracker.ProjectStatus) []ProjectStatusResource {
	resources := make([]ProjectStatusResource, 0, len(statuses))
	for _, status := range statuses {
		resources = append(resources, projectStatusResource(status))
	}
	return resources
}

func projectLabelResources(labels []tracker.ProjectLabel) []ProjectLabelResource {
	resources := make([]ProjectLabelResource, 0, len(labels))
	for _, label := range labels {
		resources = append(resources, ProjectLabelResource{
			Metadata: ProjectLabelMetadata{
				ID:        label.Label,
				ProjectID: label.ProjectID,
			},
			Spec: ProjectLabelSpec{
				Label: label.Label,
			},
			Status: ProjectLabelStatus{
				TicketCount: label.TicketCount,
			},
		})
	}
	return resources
}

func roadmapItemResources(items []tracker.RoadmapItem) []RoadmapItemResource {
	resources := make([]RoadmapItemResource, 0, len(items))
	for _, item := range items {
		epic := ticketapi.ResourceFromTracker(item.Epic)
		resources = append(resources, RoadmapItemResource{
			Metadata: RoadmapItemMetadata{
				ID:        item.Epic.ID,
				ProjectID: item.Epic.ProjectID,
			},
			Spec: RoadmapItemSpec{
				Epic: epic,
			},
			Status: RoadmapItemStatus{
				Progress: item.Progress,
			},
		})
	}
	return resources
}
