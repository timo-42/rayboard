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

type NotificationOutput struct {
	Body NotificationResource
}

type ListNotificationsOutput = shared.ListOutput[NotificationResource]

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
