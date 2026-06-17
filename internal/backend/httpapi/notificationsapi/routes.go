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
	huma.Register(api, shared.Operation(http.MethodGet, "/api/me/notification-preferences", "Notification Preferences", "Get current user notification preferences"), provider.getMyPreferences)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/me/notification-preferences", "Notification Preferences", "Update current user notification preferences"), provider.updateMyPreferences)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/notification-preferences", "Notification Preferences", "Get project notification defaults"), provider.getProjectPreferences)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/projects/{project_id}/notification-preferences", "Notification Preferences", "Update project notification defaults"), provider.updateProjectPreferences)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notification-policies", "Notification Policies", "List global notification policies"), provider.listGlobalPolicies)
	huma.Register(api, operation(http.MethodPost, "/api/notification-policies", "Notification Policies", "Create global notification policy", http.StatusCreated), provider.createGlobalPolicy)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/notification-policies", "Notification Policies", "List project notification policies"), provider.listProjectPolicies)
	huma.Register(api, operation(http.MethodPost, "/api/projects/{project_id}/notification-policies", "Notification Policies", "Create project notification policy", http.StatusCreated), provider.createProjectPolicy)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notification-policies/{policy_id}", "Notification Policies", "Get notification policy"), provider.getPolicy)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/notification-policies/{policy_id}", "Notification Policies", "Update notification policy"), provider.updatePolicy)
	huma.Register(api, operation(http.MethodDelete, "/api/notification-policies/{policy_id}", "Notification Policies", "Delete notification policy", http.StatusNoContent), provider.deletePolicy)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notification-hooks", "Notification Hooks", "List global notification hooks"), provider.listGlobalHooks)
	huma.Register(api, operation(http.MethodPost, "/api/notification-hooks", "Notification Hooks", "Create global notification hook", http.StatusCreated), provider.createGlobalHook)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/notification-hooks", "Notification Hooks", "List project notification hooks"), provider.listProjectHooks)
	huma.Register(api, operation(http.MethodPost, "/api/projects/{project_id}/notification-hooks", "Notification Hooks", "Create project notification hook", http.StatusCreated), provider.createProjectHook)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notification-hooks/{hook_id}", "Notification Hooks", "Get notification hook"), provider.getHook)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/notification-hooks/{hook_id}", "Notification Hooks", "Update notification hook"), provider.updateHook)
	huma.Register(api, operation(http.MethodDelete, "/api/notification-hooks/{hook_id}", "Notification Hooks", "Delete notification hook", http.StatusNoContent), provider.deleteHook)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notification-deliveries", "Notification Deliveries", "List global notification deliveries"), provider.listGlobalDeliveries)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/notification-deliveries", "Notification Deliveries", "List project notification deliveries"), provider.listProjectDeliveries)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notification-deliveries/{delivery_id}", "Notification Deliveries", "Get notification delivery"), provider.getDelivery)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/notification-deliveries/{delivery_id}/retry", "Notification Deliveries", "Retry notification delivery"), provider.retryDelivery)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notification-destinations", "Notification Destinations", "List global notification destinations"), provider.listGlobalDestinations)
	huma.Register(api, operation(http.MethodPost, "/api/notification-destinations", "Notification Destinations", "Create global notification destination", http.StatusCreated), provider.createGlobalDestination)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/notification-destinations", "Notification Destinations", "List project notification destinations"), provider.listProjectDestinations)
	huma.Register(api, operation(http.MethodPost, "/api/projects/{project_id}/notification-destinations", "Notification Destinations", "Create project notification destination", http.StatusCreated), provider.createProjectDestination)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/notification-destinations/{destination_id}", "Notification Destinations", "Get notification destination"), provider.getDestination)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/notification-destinations/{destination_id}", "Notification Destinations", "Update notification destination"), provider.updateDestination)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/notification-destinations/{destination_id}/test-send", "Notification Destinations", "Send notification destination test message"), provider.testDestination)
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
	return &ListNotificationsOutput{Body: shared.NewListResource[NotificationResource](notificationResources(items))}, nil
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

func (provider Provider) getMyPreferences(ctx context.Context, input *PreferencesInput) (*PreferencesOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	preferences, err := provider.Notifications.GetUserPreferences(ctx, principal.UserID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &PreferencesOutput{Body: preferencesResource(preferences)}, nil
}

func (provider Provider) updateMyPreferences(ctx context.Context, input *UpdatePreferencesInput) (*PreferencesOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	preferences, err := provider.Notifications.UpdateUserPreferences(ctx, principal.UserID, input.Body.Spec.updateInput())
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &PreferencesOutput{Body: preferencesResource(preferences)}, nil
}

func (provider Provider) getProjectPreferences(ctx context.Context, input *ProjectPreferencesInput) (*PreferencesOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	preferences, err := provider.Notifications.GetProjectPreferences(ctx, input.ProjectID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &PreferencesOutput{Body: preferencesResource(preferences)}, nil
}

func (provider Provider) updateProjectPreferences(ctx context.Context, input *UpdateProjectPreferencesInput) (*PreferencesOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	preferences, err := provider.Notifications.UpdateProjectPreferences(ctx, input.ProjectID, input.Body.Spec.updateInput())
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &PreferencesOutput{Body: preferencesResource(preferences)}, nil
}

func (provider Provider) listGlobalPolicies(ctx context.Context, input *ListPoliciesInput) (*ListPoliciesOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	items, err := provider.Notifications.ListPolicies(ctx, notifications.ListPoliciesInput{ScopeType: notifications.PolicyScopeGlobal})
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &ListPoliciesOutput{Body: shared.NewListResource[PolicyResource](policyResources(items))}, nil
}

func (provider Provider) createGlobalPolicy(ctx context.Context, input *CreatePolicyInput) (*CreatePolicyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	policy, err := provider.Notifications.CreatePolicy(ctx, input.Body.Spec.createInput(notifications.PolicyScopeGlobal, ""))
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &CreatePolicyOutput{Body: policyResource(policy)}, nil
}

func (provider Provider) listProjectPolicies(ctx context.Context, input *ProjectPoliciesInput) (*ListPoliciesOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	items, err := provider.Notifications.ListPolicies(ctx, notifications.ListPoliciesInput{
		ScopeType: notifications.PolicyScopeProject,
		ProjectID: input.ProjectID,
	})
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &ListPoliciesOutput{Body: shared.NewListResource[PolicyResource](policyResources(items))}, nil
}

func (provider Provider) createProjectPolicy(ctx context.Context, input *CreateProjectPolicyInput) (*CreatePolicyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	policy, err := provider.Notifications.CreatePolicy(ctx, input.Body.Spec.createInput(notifications.PolicyScopeProject, input.ProjectID))
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &CreatePolicyOutput{Body: policyResource(policy)}, nil
}

func (provider Provider) getPolicy(ctx context.Context, input *PolicyIDInput) (*PolicyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	policy, err := provider.Notifications.GetPolicy(ctx, input.PolicyID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requirePolicyManage(principal, policy); err != nil {
		return nil, err
	}
	return &PolicyOutput{Body: policyResource(policy)}, nil
}

func (provider Provider) updatePolicy(ctx context.Context, input *UpdatePolicyInput) (*PolicyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	current, err := provider.Notifications.GetPolicy(ctx, input.PolicyID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requirePolicyManage(principal, current); err != nil {
		return nil, err
	}
	updated, err := provider.Notifications.UpdatePolicy(ctx, input.PolicyID, input.Body.Spec.updateInput())
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &PolicyOutput{Body: policyResource(updated)}, nil
}

func (provider Provider) deletePolicy(ctx context.Context, input *PolicyIDInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	current, err := provider.Notifications.GetPolicy(ctx, input.PolicyID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requirePolicyManage(principal, current); err != nil {
		return nil, err
	}
	if err := provider.Notifications.DeletePolicy(ctx, input.PolicyID); err != nil {
		return nil, shared.NotificationError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) listGlobalHooks(ctx context.Context, input *ListHooksInput) (*ListHooksOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	hooks, err := provider.Notifications.ListHooks(ctx, notifications.ListHooksInput{ScopeType: notifications.PolicyScopeGlobal})
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &ListHooksOutput{Body: shared.NewListResource[NotificationHookResource](hookResources(hooks))}, nil
}

func (provider Provider) createGlobalHook(ctx context.Context, input *CreateHookInput) (*CreateHookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	hook, err := provider.Notifications.CreateHook(ctx, input.Body.Spec.createInput(notifications.PolicyScopeGlobal, ""))
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &CreateHookOutput{Body: hookResource(hook)}, nil
}

func (provider Provider) listProjectHooks(ctx context.Context, input *ProjectHooksInput) (*ListHooksOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	hooks, err := provider.Notifications.ListHooks(ctx, notifications.ListHooksInput{ScopeType: notifications.PolicyScopeProject, ProjectID: input.ProjectID})
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &ListHooksOutput{Body: shared.NewListResource[NotificationHookResource](hookResources(hooks))}, nil
}

func (provider Provider) createProjectHook(ctx context.Context, input *CreateProjectHookInput) (*CreateHookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	hook, err := provider.Notifications.CreateHook(ctx, input.Body.Spec.createInput(notifications.PolicyScopeProject, input.ProjectID))
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &CreateHookOutput{Body: hookResource(hook)}, nil
}

func (provider Provider) getHook(ctx context.Context, input *HookIDInput) (*HookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	hook, err := provider.Notifications.GetHook(ctx, input.HookID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requireHookManage(principal, hook); err != nil {
		return nil, err
	}
	return &HookOutput{Body: hookResource(hook)}, nil
}

func (provider Provider) updateHook(ctx context.Context, input *UpdateHookInput) (*HookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	current, err := provider.Notifications.GetHook(ctx, input.HookID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requireHookManage(principal, current); err != nil {
		return nil, err
	}
	updated, err := provider.Notifications.UpdateHook(ctx, input.HookID, input.Body.Spec.updateInput())
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &HookOutput{Body: hookResource(updated)}, nil
}

func (provider Provider) deleteHook(ctx context.Context, input *HookIDInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	current, err := provider.Notifications.GetHook(ctx, input.HookID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requireHookManage(principal, current); err != nil {
		return nil, err
	}
	if err := provider.Notifications.DeleteHook(ctx, input.HookID); err != nil {
		return nil, shared.NotificationError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) listGlobalDeliveries(ctx context.Context, input *ListDeliveriesInput) (*ListDeliveriesOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	items, err := provider.Notifications.ListDeliveries(ctx, notifications.ListDeliveriesInput{
		ScopeType:     notifications.PolicyScopeGlobal,
		Status:        input.Status,
		PolicyID:      input.PolicyID,
		DestinationID: input.DestinationID,
		Limit:         input.Limit,
		Offset:        input.Offset,
	})
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &ListDeliveriesOutput{Body: shared.NewListResource[DeliveryResource](deliveryResources(items))}, nil
}

func (provider Provider) listProjectDeliveries(ctx context.Context, input *ProjectDeliveriesInput) (*ListDeliveriesOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, authz.ProjectScope(input.ProjectID)); err != nil {
		return nil, err
	}
	items, err := provider.Notifications.ListDeliveries(ctx, notifications.ListDeliveriesInput{
		ScopeType:     notifications.PolicyScopeProject,
		ProjectID:     input.ProjectID,
		Status:        input.Status,
		PolicyID:      input.PolicyID,
		DestinationID: input.DestinationID,
		Limit:         input.Limit,
		Offset:        input.Offset,
	})
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &ListDeliveriesOutput{Body: shared.NewListResource[DeliveryResource](deliveryResources(items))}, nil
}

func (provider Provider) getDelivery(ctx context.Context, input *DeliveryIDInput) (*DeliveryOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	delivery, err := provider.Notifications.GetDelivery(ctx, input.DeliveryID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requireDeliveryManage(principal, delivery); err != nil {
		return nil, err
	}
	return &DeliveryOutput{Body: deliveryResource(delivery)}, nil
}

func (provider Provider) retryDelivery(ctx context.Context, input *DeliveryIDInput) (*DeliveryOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	current, err := provider.Notifications.GetDelivery(ctx, input.DeliveryID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	if err := provider.requireDeliveryManage(principal, current); err != nil {
		return nil, err
	}
	delivery, err := provider.Notifications.RetryDelivery(ctx, input.DeliveryID)
	if err != nil {
		return nil, shared.NotificationError(err)
	}
	return &DeliveryOutput{Body: deliveryResource(delivery)}, nil
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
	return &ListDestinationsOutput{Body: shared.NewListResource[DestinationResource](destinationResources(items))}, nil
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
	return &ListDestinationsOutput{Body: shared.NewListResource[DestinationResource](destinationResources(items))}, nil
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

func (provider Provider) testDestination(ctx context.Context, input *TestDestinationInput) (*DestinationOutput, error) {
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
	tested, err := provider.Notifications.TestDestination(ctx, input.DestinationID, input.Body.Spec.testInput())
	if err != nil {
		_ = provider.auditDestination(ctx, principal, "notification.destination_test_sent", current, map[string]any{"delivery_status": "failed"})
		return nil, shared.NotificationError(err)
	}
	if err := provider.auditDestination(ctx, principal, "notification.destination_test_sent", tested, map[string]any{"delivery_status": "delivered"}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &DestinationOutput{Body: destinationResource(tested)}, nil
}

func (provider Provider) requireDestinationManage(principal authz.Principal, destination notifications.Destination) error {
	scope := authz.GlobalScope()
	if destination.ScopeType == notifications.DestinationScopeProject || destination.ScopeType == notifications.DestinationScopeDashboard {
		scope = authz.ProjectScope(destination.ProjectID)
	}
	return provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, scope)
}

func (provider Provider) requirePolicyManage(principal authz.Principal, policy notifications.Policy) error {
	scope := authz.GlobalScope()
	if policy.ScopeType == notifications.PolicyScopeProject {
		scope = authz.ProjectScope(policy.ProjectID)
	}
	return provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, scope)
}

func (provider Provider) requireHookManage(principal authz.Principal, hook notifications.Hook) error {
	scope := authz.GlobalScope()
	if hook.ScopeType == notifications.PolicyScopeProject {
		scope = authz.ProjectScope(hook.ProjectID)
	}
	return provider.Authenticator.Require(principal, authz.PermissionNotificationsManage, scope)
}

func (provider Provider) requireDeliveryManage(principal authz.Principal, delivery notifications.Delivery) error {
	scope := authz.GlobalScope()
	if delivery.ScopeType == notifications.PolicyScopeProject {
		scope = authz.ProjectScope(delivery.ProjectID)
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
