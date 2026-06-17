package settingsapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/settings"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/settings", "Settings", "Get global settings"), provider.getSettings)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/settings", "Settings", "Update global settings"), provider.updateSettings)
}

func (provider Provider) getSettings(ctx context.Context, input *GetSettingsInput) (*SettingsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionSettingsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	global, err := provider.Settings.GetGlobal(ctx)
	if err != nil {
		return nil, settingsError(err)
	}
	return &SettingsOutput{Body: settingsResource(global)}, nil
}

func (provider Provider) updateSettings(ctx context.Context, input *UpdateSettingsInput) (*SettingsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionSettingsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	current, err := provider.Settings.GetGlobal(ctx)
	if err != nil {
		return nil, settingsError(err)
	}
	actorID := principal.ActorUserID
	if actorID == "" {
		actorID = principal.UserID
	}
	updated, err := provider.Settings.UpdateGlobal(ctx, input.Body.Spec.updateInput(actorID))
	if err != nil {
		return nil, settingsError(err)
	}
	if err := provider.recordAudit(ctx, principal, audit.RecordInput{
		EventType:   "settings.updated",
		SubjectType: "settings",
		SubjectID:   settings.GlobalSettingsKey,
		Payload: map[string]any{
			"changed_fields": changedFields(current, updated),
		},
	}); err != nil {
		return nil, huma.Error500InternalServerError("Could not write audit log")
	}
	return &SettingsOutput{Body: settingsResource(updated)}, nil
}

func settingsError(err error) error {
	var validation *settings.ValidationError
	switch {
	case errors.As(err, &validation):
		return huma.Error400BadRequest(validation.Message)
	case errors.Is(err, settings.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, settings.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}

func changedFields(before settings.GlobalSettings, after settings.GlobalSettings) []string {
	fields := []string{}
	if before.AttachmentMaxSizeBytes != after.AttachmentMaxSizeBytes {
		fields = append(fields, "attachment_max_size_bytes")
	}
	if !stringSlicesEqual(before.AttachmentAllowedContentTypes, after.AttachmentAllowedContentTypes) {
		fields = append(fields, "attachment_allowed_content_types")
	}
	if !stringSlicesEqual(before.WebhookAllowedBaseURLs, after.WebhookAllowedBaseURLs) {
		fields = append(fields, "webhook_allowed_base_urls")
	}
	if before.DemoWarningEnabled != after.DemoWarningEnabled {
		fields = append(fields, "demo_warning_enabled")
	}
	if before.BackupEnabled != after.BackupEnabled {
		fields = append(fields, "backup_enabled")
	}
	if before.SystemHealthNote != after.SystemHealthNote {
		fields = append(fields, "system_health_note")
	}
	return fields
}

func stringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
