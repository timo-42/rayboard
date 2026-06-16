package components

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type ComponentIDInput struct {
	shared.AuthInput
	ComponentID string `path:"component_id"`
}

type UpdateComponentInput struct {
	shared.AuthInput
	ComponentID string `path:"component_id"`
	Body        shared.ResourceInput[UpdateComponentSpec]
}

type ComponentOutput struct {
	Body ComponentResource
}

type ComponentMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ComponentSpec struct {
	Name              string `json:"name,omitempty"`
	Description       string `json:"description,omitempty"`
	OwnerUserID       string `json:"owner_user_id,omitempty"`
	DefaultAssigneeID string `json:"default_assignee_id,omitempty"`
}

type UpdateComponentSpec struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	OwnerUserID       *string `json:"owner_user_id,omitempty"`
	DefaultAssigneeID *string `json:"default_assignee_id,omitempty"`
}

type ComponentStatus struct {
}

type ComponentResource = shared.Resource[ComponentMetadata, ComponentSpec, ComponentStatus]

func (spec ComponentSpec) ToCreateInput(projectID string) tracker.CreateComponentInput {
	return tracker.CreateComponentInput{
		ProjectID:         projectID,
		Name:              spec.Name,
		Description:       spec.Description,
		OwnerUserID:       spec.OwnerUserID,
		DefaultAssigneeID: spec.DefaultAssigneeID,
	}
}

func (spec UpdateComponentSpec) ToUpdateInput() tracker.UpdateComponentInput {
	return tracker.UpdateComponentInput{
		Name:              spec.Name,
		Description:       spec.Description,
		OwnerUserID:       spec.OwnerUserID,
		DefaultAssigneeID: spec.DefaultAssigneeID,
	}
}

func ResourceFromTracker(component tracker.Component) ComponentResource {
	return ComponentResource{
		Metadata: ComponentMetadata{
			ID:        component.ID,
			ProjectID: component.ProjectID,
			CreatedAt: component.CreatedAt,
			UpdatedAt: component.UpdatedAt,
		},
		Spec: ComponentSpec{
			Name:              component.Name,
			Description:       component.Description,
			OwnerUserID:       component.OwnerUserID,
			DefaultAssigneeID: component.DefaultAssigneeID,
		},
		Status: ComponentStatus{},
	}
}

func ResourcesFromTracker(components []tracker.Component) []ComponentResource {
	resources := make([]ComponentResource, 0, len(components))
	for _, component := range components {
		resources = append(resources, ResourceFromTracker(component))
	}
	return resources
}
