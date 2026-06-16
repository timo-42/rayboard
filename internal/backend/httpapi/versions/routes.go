package versions

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/versions/{version_id}", "Versions", "Get version"), provider.getVersion)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/versions/{version_id}", "Versions", "Update version"), provider.updateVersion)
	huma.Register(api, shared.OperationWithStatus(http.MethodDelete, "/api/versions/{version_id}", "Versions", "Delete version", http.StatusNoContent), provider.deleteVersion)
}

func (provider Provider) getVersion(ctx context.Context, input *VersionIDInput) (*VersionOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	version, err := provider.Tracker.GetVersion(ctx, principal, input.VersionID)
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &VersionOutput{Body: ResourceFromTracker(version)}, nil
}

func (provider Provider) updateVersion(ctx context.Context, input *UpdateVersionInput) (*VersionOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	version, err := provider.Tracker.UpdateVersion(ctx, principal, input.VersionID, input.Body.Spec.ToUpdateInput())
	if err != nil {
		return nil, shared.TrackerError(err)
	}
	return &VersionOutput{Body: ResourceFromTracker(version)}, nil
}

func (provider Provider) deleteVersion(ctx context.Context, input *VersionIDInput) (*shared.EmptyOutput, error) {
	_, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Tracker.DeleteVersion(ctx, principal, input.VersionID); err != nil {
		return nil, shared.TrackerError(err)
	}
	return &shared.EmptyOutput{}, nil
}
