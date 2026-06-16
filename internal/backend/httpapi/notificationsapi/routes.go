package notificationsapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/notifications"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notifications", "Notifications", "List notifications"), provider.listNotifications)
	huma.Register(api, operation(http.MethodPost, "/api/notifications/read-all", "Notifications", "Mark all notifications read", http.StatusNoContent), provider.markAllRead)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/notifications/{notification_id}/read", "Notifications", "Mark notification read"), provider.markRead)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/notifications/{notification_id}/unread", "Notifications", "Mark notification unread"), provider.markUnread)
}

func (provider Provider) listNotifications(ctx context.Context, input *ListNotificationsInput) (*ListNotificationsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	items, err := provider.Notifications.List(ctx, principal, notifications.ListInput{
		UnreadOnly: input.UnreadOnly,
		Limit:      input.Limit,
		Offset:     input.Offset,
	})
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &ListNotificationsOutput{Body: shared.ItemList[NotificationResource]{Items: notificationResources(items)}}, nil
}

func (provider Provider) markAllRead(ctx context.Context, input *MarkAllReadInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Notifications.MarkAllRead(ctx, principal); err != nil {
		return nil, shared.NotificationError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) markRead(ctx context.Context, input *NotificationIDInput) (*NotificationOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	notification, err := provider.Notifications.SetRead(ctx, principal, input.NotificationID, true)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &NotificationOutput{Body: notificationResource(notification)}, nil
}

func (provider Provider) markUnread(ctx context.Context, input *NotificationIDInput) (*NotificationOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	notification, err := provider.Notifications.SetRead(ctx, principal, input.NotificationID, false)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &NotificationOutput{Body: notificationResource(notification)}, nil
}

func operation(method string, path string, tag string, summary string, status int) huma.Operation {
	op := shared.Operation(method, path, tag, summary)
	op.DefaultStatus = status
	return op
}
