package settingsapi

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/settings"
)

type GetSettingsInput struct {
	shared.AuthInput
}

type UpdateSettingsInput struct {
	shared.AuthInput
	Body shared.ResourceInput[UpdateSettingsSpec]
}

type SettingsOutput struct {
	Body SettingsResource
}

type SettingsMetadata struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by,omitempty"`
}

type SettingsSpec struct {
	AttachmentMaxSizeBytes        int64    `json:"attachment_max_size_bytes"`
	AttachmentAllowedContentTypes []string `json:"attachment_allowed_content_types"`
	WebhookAllowedBaseURLs        []string `json:"webhook_allowed_base_urls"`
	DemoWarningEnabled            bool     `json:"demo_warning_enabled"`
	BackupEnabled                 bool     `json:"backup_enabled"`
	SystemHealthNote              string   `json:"system_health_note,omitempty"`
}

type UpdateSettingsSpec struct {
	AttachmentMaxSizeBytes        *int64    `json:"attachment_max_size_bytes,omitempty"`
	AttachmentAllowedContentTypes *[]string `json:"attachment_allowed_content_types,omitempty"`
	WebhookAllowedBaseURLs        *[]string `json:"webhook_allowed_base_urls,omitempty"`
	DemoWarningEnabled            *bool     `json:"demo_warning_enabled,omitempty"`
	BackupEnabled                 *bool     `json:"backup_enabled,omitempty"`
	SystemHealthNote              *string   `json:"system_health_note,omitempty"`
}

type SettingsStatus struct {
	AttachmentPolicyActive bool `json:"attachment_policy_active"`
	WebhookAllowlistActive bool `json:"webhook_allowlist_active"`
	DemoWarningVisible     bool `json:"demo_warning_visible"`
	BackupAvailable        bool `json:"backup_available"`
}

type SettingsResource = shared.Resource[SettingsMetadata, SettingsSpec, SettingsStatus]

func (spec UpdateSettingsSpec) updateInput(actorID string) settings.UpdateGlobalInput {
	return settings.UpdateGlobalInput{
		AttachmentMaxSizeBytes:        spec.AttachmentMaxSizeBytes,
		AttachmentAllowedContentTypes: spec.AttachmentAllowedContentTypes,
		WebhookAllowedBaseURLs:        spec.WebhookAllowedBaseURLs,
		DemoWarningEnabled:            spec.DemoWarningEnabled,
		BackupEnabled:                 spec.BackupEnabled,
		SystemHealthNote:              spec.SystemHealthNote,
		UpdatedBy:                     actorID,
	}
}

func settingsResource(global settings.GlobalSettings) SettingsResource {
	spec := SettingsSpec{
		AttachmentMaxSizeBytes:        global.AttachmentMaxSizeBytes,
		AttachmentAllowedContentTypes: global.AttachmentAllowedContentTypes,
		WebhookAllowedBaseURLs:        global.WebhookAllowedBaseURLs,
		DemoWarningEnabled:            global.DemoWarningEnabled,
		BackupEnabled:                 global.BackupEnabled,
		SystemHealthNote:              global.SystemHealthNote,
	}
	return SettingsResource{
		Metadata: SettingsMetadata{
			ID:        settings.GlobalSettingsKey,
			UpdatedAt: global.UpdatedAt,
			UpdatedBy: global.UpdatedBy,
		},
		Spec: spec,
		Status: SettingsStatus{
			AttachmentPolicyActive: global.AttachmentMaxSizeBytes > 0 || len(global.AttachmentAllowedContentTypes) > 0,
			WebhookAllowlistActive: len(global.WebhookAllowedBaseURLs) > 0,
			DemoWarningVisible:     global.DemoWarningEnabled,
			BackupAvailable:        global.BackupEnabled,
		},
	}
}
