package notificationsapi

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/notifications"
)

type ListNotificationsInput struct {
	shared.AuthInput
	UnreadOnly bool `query:"unread" doc:"Only include unread notifications."`
	Limit      int  `query:"limit" doc:"Maximum number of notifications to return."`
	Offset     int  `query:"offset" doc:"Number of notifications to skip."`
}

type NotificationIDInput struct {
	shared.AuthInput
	NotificationID string `path:"notification_id" doc:"Notification ID."`
}

type MarkAllReadInput struct {
	shared.AuthInput
}

type ListDestinationsInput struct {
	shared.AuthInput
}

type ProjectDestinationsInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
}

type CreateDestinationInput struct {
	shared.AuthInput
	Body shared.ResourceInput[CreateDestinationSpec]
}

type CreateProjectDestinationInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
	Body      shared.ResourceInput[CreateDestinationSpec]
}

type DestinationIDInput struct {
	shared.AuthInput
	DestinationID string `path:"destination_id" doc:"Notification destination ID."`
}

type UpdateDestinationInput struct {
	shared.AuthInput
	DestinationID string `path:"destination_id" doc:"Notification destination ID."`
	Body          shared.ResourceInput[UpdateDestinationSpec]
}

type NotificationOutput struct {
	Body NotificationResource
}

type ListNotificationsOutput = shared.ListOutput[NotificationResource]
type ListDestinationsOutput = shared.ListOutput[DestinationResource]
type CreateDestinationOutput = shared.CreatedOutput[DestinationResource]

type DestinationOutput struct {
	Body DestinationResource
}

type NotificationMetadata struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationSpec struct {
	Type        string         `json:"type"`
	SubjectType string         `json:"subject_type,omitempty"`
	SubjectID   string         `json:"subject_id,omitempty"`
	Body        string         `json:"body"`
	Data        map[string]any `json:"data"`
}

type NotificationStatus struct {
	ReadAt *time.Time `json:"read_at"`
}

type NotificationResource = shared.Resource[NotificationMetadata, NotificationSpec, NotificationStatus]

type DestinationMetadata struct {
	ID          string    `json:"id"`
	ScopeType   string    `json:"scope_type"`
	ProjectID   string    `json:"project_id,omitempty"`
	DashboardID string    `json:"dashboard_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DestinationSpec struct {
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
}

type CreateDestinationSpec struct {
	Name        string `json:"name,omitempty"`
	ShoutrrrURL string `json:"shoutrrr_url,omitempty" doc:"Shoutrrr service URL. Write-only; never returned in responses."`
	Enabled     *bool  `json:"enabled,omitempty"`
}

type UpdateDestinationSpec struct {
	Name        *string `json:"name,omitempty"`
	ShoutrrrURL *string `json:"shoutrrr_url,omitempty" doc:"Shoutrrr service URL. Omit to leave unchanged; empty string is rejected."`
	Enabled     *bool   `json:"enabled,omitempty"`
}

type DestinationStatus struct {
	URLSet             bool       `json:"url_set"`
	LastDeliveryStatus string     `json:"last_delivery_status,omitempty"`
	LastDeliveryAt     *time.Time `json:"last_delivery_at,omitempty"`
	LastError          string     `json:"last_error,omitempty"`
	Deleted            bool       `json:"deleted"`
}

type DestinationResource = shared.Resource[DestinationMetadata, DestinationSpec, DestinationStatus]

func notificationResource(notification notifications.Notification) NotificationResource {
	return NotificationResource{
		Metadata: NotificationMetadata{
			ID:        notification.ID,
			UserID:    notification.UserID,
			CreatedAt: notification.CreatedAt,
		},
		Spec: NotificationSpec{
			Type:        notification.Type,
			SubjectType: notification.SubjectType,
			SubjectID:   notification.SubjectID,
			Body:        notification.Body,
			Data:        notification.Data,
		},
		Status: NotificationStatus{
			ReadAt: notification.ReadAt,
		},
	}
}

func notificationResources(items []notifications.Notification) []NotificationResource {
	resources := make([]NotificationResource, 0, len(items))
	for _, item := range items {
		resources = append(resources, notificationResource(item))
	}
	return resources
}

func (spec CreateDestinationSpec) createInput(scopeType string, projectID string) notifications.CreateDestinationInput {
	enabled := true
	if spec.Enabled != nil {
		enabled = *spec.Enabled
	}
	return notifications.CreateDestinationInput{
		Name:        spec.Name,
		ScopeType:   scopeType,
		ProjectID:   projectID,
		ShoutrrrURL: spec.ShoutrrrURL,
		Enabled:     enabled,
	}
}

func (spec UpdateDestinationSpec) updateInput() notifications.UpdateDestinationInput {
	return notifications.UpdateDestinationInput{
		Name:        spec.Name,
		ShoutrrrURL: spec.ShoutrrrURL,
		Enabled:     spec.Enabled,
	}
}

func destinationResource(destination notifications.Destination) DestinationResource {
	return DestinationResource{
		Metadata: DestinationMetadata{
			ID:          destination.ID,
			ScopeType:   destination.ScopeType,
			ProjectID:   destination.ProjectID,
			DashboardID: destination.DashboardID,
			CreatedAt:   destination.CreatedAt,
			UpdatedAt:   destination.UpdatedAt,
		},
		Spec: DestinationSpec{
			Name:    destination.Name,
			Type:    destination.Service,
			Enabled: destination.Enabled,
		},
		Status: DestinationStatus{
			URLSet:             destination.URLSet,
			LastDeliveryStatus: destination.LastDeliveryStatus,
			LastDeliveryAt:     destination.LastDeliveryAt,
			LastError:          destination.LastError,
			Deleted:            false,
		},
	}
}

func destinationResources(items []notifications.Destination) []DestinationResource {
	resources := make([]DestinationResource, 0, len(items))
	for _, item := range items {
		resources = append(resources, destinationResource(item))
	}
	return resources
}
