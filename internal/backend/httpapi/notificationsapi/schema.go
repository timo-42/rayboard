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

type PreferencesInput struct {
	shared.AuthInput
}

type UpdatePreferencesInput struct {
	shared.AuthInput
	Body shared.ResourceInput[UpdatePreferencesSpec]
}

type ProjectPreferencesInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
}

type UpdateProjectPreferencesInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
	Body      shared.ResourceInput[UpdatePreferencesSpec]
}

type ListPoliciesInput struct {
	shared.AuthInput
}

type CreatePolicyInput struct {
	shared.AuthInput
	Body shared.ResourceInput[CreatePolicySpec]
}

type ProjectPoliciesInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
}

type CreateProjectPolicyInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
	Body      shared.ResourceInput[CreatePolicySpec]
}

type PolicyIDInput struct {
	shared.AuthInput
	PolicyID string `path:"policy_id" doc:"Notification policy ID."`
}

type UpdatePolicyInput struct {
	shared.AuthInput
	PolicyID string `path:"policy_id" doc:"Notification policy ID."`
	Body     shared.ResourceInput[UpdatePolicySpec]
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

type TestDestinationInput struct {
	shared.AuthInput
	DestinationID string `path:"destination_id" doc:"Notification destination ID."`
	Body          shared.ResourceInput[TestDestinationSpec]
}

type NotificationOutput struct {
	Body NotificationResource
}

type ListNotificationsOutput = shared.ListOutput[NotificationResource]
type PreferencesOutput struct {
	Body PreferencesResource
}
type ListPoliciesOutput = shared.ListOutput[PolicyResource]
type CreatePolicyOutput = shared.CreatedOutput[PolicyResource]
type ListDestinationsOutput = shared.ListOutput[DestinationResource]
type CreateDestinationOutput = shared.CreatedOutput[DestinationResource]

type PolicyOutput struct {
	Body PolicyResource
}

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

type PreferencesMetadata struct {
	ID        string     `json:"id,omitempty"`
	ScopeType string     `json:"scope_type"`
	UserID    string     `json:"user_id,omitempty"`
	ProjectID string     `json:"project_id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type PreferencesSpec struct {
	InAppEnabled             bool `json:"in_app_enabled"`
	ExternalEnabled          bool `json:"external_enabled"`
	AssignmentEnabled        bool `json:"assignment_enabled"`
	CommentEnabled           bool `json:"comment_enabled"`
	StatusChangeEnabled      bool `json:"status_change_enabled"`
	SprintChangeEnabled      bool `json:"sprint_change_enabled"`
	ReleaseChangeEnabled     bool `json:"release_change_enabled"`
	AutomationFailureEnabled bool `json:"automation_failure_enabled"`
}

type UpdatePreferencesSpec struct {
	InAppEnabled             *bool `json:"in_app_enabled,omitempty"`
	ExternalEnabled          *bool `json:"external_enabled,omitempty"`
	AssignmentEnabled        *bool `json:"assignment_enabled,omitempty"`
	CommentEnabled           *bool `json:"comment_enabled,omitempty"`
	StatusChangeEnabled      *bool `json:"status_change_enabled,omitempty"`
	SprintChangeEnabled      *bool `json:"sprint_change_enabled,omitempty"`
	ReleaseChangeEnabled     *bool `json:"release_change_enabled,omitempty"`
	AutomationFailureEnabled *bool `json:"automation_failure_enabled,omitempty"`
}

type PreferencesStatus struct {
	Customized bool `json:"customized"`
}

type PreferencesResource = shared.Resource[PreferencesMetadata, PreferencesSpec, PreferencesStatus]

type PolicyMetadata struct {
	ID        string    `json:"id"`
	ScopeType string    `json:"scope_type"`
	ProjectID string    `json:"project_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PolicySpec struct {
	Name           string   `json:"name,omitempty"`
	EventTypes     []string `json:"event_types,omitempty"`
	DestinationIDs []string `json:"destination_ids,omitempty"`
	Enabled        bool     `json:"enabled,omitempty"`
}

type CreatePolicySpec struct {
	Name           string   `json:"name,omitempty"`
	EventTypes     []string `json:"event_types,omitempty"`
	DestinationIDs []string `json:"destination_ids,omitempty"`
	Enabled        *bool    `json:"enabled,omitempty"`
}

type UpdatePolicySpec struct {
	Name           *string   `json:"name,omitempty"`
	EventTypes     *[]string `json:"event_types,omitempty"`
	DestinationIDs *[]string `json:"destination_ids,omitempty"`
	Enabled        *bool     `json:"enabled,omitempty"`
}

type PolicyStatus struct {
	Deleted bool `json:"deleted"`
}

type PolicyResource = shared.Resource[PolicyMetadata, PolicySpec, PolicyStatus]

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

type TestDestinationSpec struct {
	Message string `json:"message,omitempty" doc:"Optional test notification message. Defaults to a Rayboard test message."`
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

func (spec UpdatePreferencesSpec) updateInput() notifications.UpdatePreferencesInput {
	return notifications.UpdatePreferencesInput{
		InAppEnabled:             spec.InAppEnabled,
		ExternalEnabled:          spec.ExternalEnabled,
		AssignmentEnabled:        spec.AssignmentEnabled,
		CommentEnabled:           spec.CommentEnabled,
		StatusChangeEnabled:      spec.StatusChangeEnabled,
		SprintChangeEnabled:      spec.SprintChangeEnabled,
		ReleaseChangeEnabled:     spec.ReleaseChangeEnabled,
		AutomationFailureEnabled: spec.AutomationFailureEnabled,
	}
}

func preferencesResource(preferences notifications.Preferences) PreferencesResource {
	return PreferencesResource{
		Metadata: PreferencesMetadata{
			ID:        preferences.ID,
			ScopeType: preferences.ScopeType,
			UserID:    preferences.UserID,
			ProjectID: preferences.ProjectID,
			CreatedAt: optionalTime(preferences.CreatedAt),
			UpdatedAt: optionalTime(preferences.UpdatedAt),
		},
		Spec: PreferencesSpec{
			InAppEnabled:             preferences.InAppEnabled,
			ExternalEnabled:          preferences.ExternalEnabled,
			AssignmentEnabled:        preferences.AssignmentEnabled,
			CommentEnabled:           preferences.CommentEnabled,
			StatusChangeEnabled:      preferences.StatusChangeEnabled,
			SprintChangeEnabled:      preferences.SprintChangeEnabled,
			ReleaseChangeEnabled:     preferences.ReleaseChangeEnabled,
			AutomationFailureEnabled: preferences.AutomationFailureEnabled,
		},
		Status: PreferencesStatus{
			Customized: preferences.Customized,
		},
	}
}

func (spec CreatePolicySpec) createInput(scopeType string, projectID string) notifications.CreatePolicyInput {
	enabled := true
	if spec.Enabled != nil {
		enabled = *spec.Enabled
	}
	return notifications.CreatePolicyInput{
		Name:           spec.Name,
		ScopeType:      scopeType,
		ProjectID:      projectID,
		EventTypes:     spec.EventTypes,
		DestinationIDs: spec.DestinationIDs,
		Enabled:        enabled,
	}
}

func (spec UpdatePolicySpec) updateInput() notifications.UpdatePolicyInput {
	return notifications.UpdatePolicyInput{
		Name:           spec.Name,
		EventTypes:     spec.EventTypes,
		DestinationIDs: spec.DestinationIDs,
		Enabled:        spec.Enabled,
	}
}

func policyResource(policy notifications.Policy) PolicyResource {
	return PolicyResource{
		Metadata: PolicyMetadata{
			ID:        policy.ID,
			ScopeType: policy.ScopeType,
			ProjectID: policy.ProjectID,
			CreatedAt: policy.CreatedAt,
			UpdatedAt: policy.UpdatedAt,
		},
		Spec: PolicySpec{
			Name:           policy.Name,
			EventTypes:     policy.EventTypes,
			DestinationIDs: policy.DestinationIDs,
			Enabled:        policy.Enabled,
		},
		Status: PolicyStatus{
			Deleted: false,
		},
	}
}

func policyResources(policies []notifications.Policy) []PolicyResource {
	resources := make([]PolicyResource, 0, len(policies))
	for _, policy := range policies {
		resources = append(resources, policyResource(policy))
	}
	return resources
}

func optionalTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
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

func (spec TestDestinationSpec) testInput() notifications.TestDestinationInput {
	return notifications.TestDestinationInput{Message: spec.Message}
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
