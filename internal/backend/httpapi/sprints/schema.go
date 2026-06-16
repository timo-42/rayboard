package sprints

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type SprintIDInput struct {
	shared.AuthInput
	SprintID string `path:"sprint_id"`
}

type UpdateSprintInput struct {
	shared.AuthInput
	SprintID string `path:"sprint_id"`
	Body     shared.ResourceInput[UpdateSprintSpec]
}

type CompleteSprintInput struct {
	shared.AuthInput
	SprintID string `path:"sprint_id"`
}

type SprintOutput struct {
	Body SprintResource
}

type SprintMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SprintSpec struct {
	Name      string `json:"name,omitempty"`
	Goal      string `json:"goal,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

type UpdateSprintSpec struct {
	Name      *string `json:"name,omitempty"`
	Goal      *string `json:"goal,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
}

type SprintStatus struct {
	State       string     `json:"state"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type SprintResource = shared.Resource[SprintMetadata, SprintSpec, SprintStatus]

func (spec SprintSpec) ToCreateInput(projectID string) tracker.CreateSprintInput {
	return tracker.CreateSprintInput{
		ProjectID: projectID,
		Name:      spec.Name,
		Goal:      spec.Goal,
		StartDate: spec.StartDate,
		EndDate:   spec.EndDate,
	}
}

func (spec UpdateSprintSpec) ToUpdateInput() tracker.UpdateSprintInput {
	return tracker.UpdateSprintInput{
		Name:      spec.Name,
		Goal:      spec.Goal,
		StartDate: spec.StartDate,
		EndDate:   spec.EndDate,
	}
}

func Resource(sprint tracker.Sprint) SprintResource {
	return SprintResource{
		Metadata: SprintMetadata{
			ID:        sprint.ID,
			ProjectID: sprint.ProjectID,
			CreatedAt: sprint.CreatedAt,
			UpdatedAt: sprint.UpdatedAt,
		},
		Spec: SprintSpec{
			Name:      sprint.Name,
			Goal:      sprint.Goal,
			StartDate: sprint.StartDate,
			EndDate:   sprint.EndDate,
		},
		Status: SprintStatus{
			State:       sprint.State,
			StartedAt:   sprint.StartedAt,
			CompletedAt: sprint.CompletedAt,
		},
	}
}

func Resources(sprints []tracker.Sprint) []SprintResource {
	resources := make([]SprintResource, 0, len(sprints))
	for _, sprint := range sprints {
		resources = append(resources, Resource(sprint))
	}
	return resources
}
