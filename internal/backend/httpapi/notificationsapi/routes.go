package notificationsapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/notifications"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notifications", "Notifications", "List notifications"), provider.listNotifications)
	huma.Register(api, operation(http.MethodPost, "/api/notifications/read-all", "Notifications", "Mark all notifications read", http.StatusNoContent), provider.markAllRead)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/notifications/{notification_id}/read", "Notifications", "Mark notification read"), provider.markRead)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/notifications/{notification_id}/unread", "Notifications", "Mark notification unread"), provider.markUnread)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notification-destinations", "Notification Destinations", "List global notification destinations"), provider.listGlobalDestinations)
	huma.Register(api, operation(http.MethodPost, "/api/notification-destinations", "Notification Destinations", "Create global notification destination", http.StatusCreated), provider.createGlobalDestination)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/notification-destinations", "Notification Destinations", "List project notification destinations"), provider.listProjectDestinations)
	huma.Register(api, operation(http.MethodPost, "/api/projects/{project_id}/notification-destinations", "Notification Destinations", "Create project notification destination", http.StatusCreated), provider.createProjectDestination)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notification-destinations/{destination_id}", "Notification Destinations", "Get notification destination"), provider.getDestination)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/notification-destinations/{destination_id}", "Notification Destinations", "Update notification destination"), provider.updateDestination)
	huma.Register(api, operation(http.MethodDelete, "/api/notification-destinations/{destination_id}", "Notification Destinations", "Delete notification destination", http.StatusNoContent), provider.deleteDestination)
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

func (provider Provider) listGlobalDestinations(ctx context.Context, input *ListDestinationsInput) (*ListDestinationsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	items, err := provider.Notifications.ListDestinations(ctx, notifications.ListDestinationsInput{ScopeType: notifications.DestinationScopeGlobal})
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &ListDestinationsOutput{Body: shared.ItemList[DestinationResource]{Items: destinationResources(items)}}, nil
}

func (provider Provider) createGlobalDestination(ctx context.Context, input *CreateDestinationInput) (*CreateDestinationOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	destination, err := provider.Notifications.CreateDestination(ctx, input.Body.Spec.createInput(notifications.DestinationScopeGlobal, ""))
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.auditDestination(ctx, principal, "notification.destination_created", destination, map[string]any{"url_set": destination.URLSet}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &CreateDestinationOutput{Body: destinationResource(destination)}, nil
}

func (provider Provider) listProjectDestinations(ctx context.Context, input *ProjectDestinationsInput) (*ListDestinationsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	items, err := provider.Notifications.ListDestinations(ctx, notifications.ListDestinationsInput{
		ScopeType: notifications.DestinationScopeProject,
		ProjectID: input.ProjectID,
	})
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &ListDestinationsOutput{Body: shared.ItemList[DestinationResource]{Items: destinationResources(items)}}, nil
}

func (provider Provider) createProjectDestination(ctx context.Context, input *CreateProjectDestinationInput) (*CreateDestinationOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	destination, err := provider.Notifications.CreateDestination(ctx, input.Body.Spec.createInput(notifications.DestinationScopeProject, input.ProjectID))
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.auditDestination(ctx, principal, "notification.destination_created", destination, map[string]any{"url_set": destination.URLSet}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &CreateDestinationOutput{Body: destinationResource(destination)}, nil
}

func (provider Provider) getDestination(ctx context.Context, input *DestinationIDInput) (*DestinationOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	destination, err := provider.Notifications.GetDestination(ctx, input.DestinationID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requireDestinationManage(principal, destination); err != nil {
		return nil, err
	}
	return &DestinationOutput{Body: destinationResource(destination)}, nil
}

func (provider Provider) updateDestination(ctx context.Context, input *UpdateDestinationInput) (*DestinationOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	current, err := provider.Notifications.GetDestination(ctx, input.DestinationID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requireDestinationManage(principal, current); err != nil {
		return nil, err
	}
	updated, err := provider.Notifications.UpdateDestination(ctx, input.DestinationID, input.Body.Spec.updateInput())
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	payload := map[string]any{}
	if input.Body.Spec.ShoutrrrURL != nil {
		payload["url_rotated"] = true
	}
	if err := provider.auditDestination(ctx, principal, "notification.destination_updated", updated, payload); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &DestinationOutput{Body: destinationResource(updated)}, nil
}

func (provider Provider) deleteDestination(ctx context.Context, input *DestinationIDInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	current, err := provider.Notifications.GetDestination(ctx, input.DestinationID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requireDestinationManage(principal, current); err != nil {
		return nil, err
	}
	if err := provider.Notifications.DeleteDestination(ctx, input.DestinationID); err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.auditDestination(ctx, principal, "notification.destination_deleted", current, nil); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) requireDestinationManage(principal authz.Principal, destination notifications.Destination) error {
	scope := authz.GlobalScope()
	if destination.ScopeType == notifications.DestinationScopeProject || destination.ScopeType == notifications.DestinationScopeDashboard {
		scope = authz.ProjectScope(destination.ProjectID)
	}
	return provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, scope)
}

func (provider Provider) auditDestination(ctx context.Context, principal authz.Principal, eventType string, destination notifications.Destination, extra map[string]any) error {
	payload := map[string]any{
		"destination_id": destination.ID,
		"name":           destination.Name,
		"scope_type":     destination.ScopeType,
		"project_id":     destination.ProjectID,
		"dashboard_id":   destination.DashboardID,
		"type":           destination.Service,
		"enabled":        destination.Enabled,
		"url_set":        destination.URLSet,
	}
	for key, value := range extra {
		payload[key] = value
	}
	return provider.recordAudit(ctx, principal, audit.RecordInput{
		EventType:   eventType,
		SubjectType: "notification_destination",
		SubjectID:   destination.ID,
		Payload:     payload,
	})
}

func operation(method string, path string, tag string, summary string, status int) huma.Operation {
	op := shared.Operation(method, path, tag, summary)
	op.DefaultStatus = status
	return op
}
