package auditapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/audit-log", "Audit Log", "List audit log entries"), provider.listAuditLog)
}

func (provider Provider) listAuditLog(ctx context.Context, input *ListAuditLogInput) (*ListAuditLogOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	if err := provider.Authenticator.Require(principal, authz.PermissionSettingsManage, authz.GlobalScope()); err != nil {
		return nil, err
	}
	if provider.Audit == nil {
		return nil, huma.Error500InternalServerError("Audit log is not configured")
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 100
	}
	entries, err := provider.Audit.ListEntries(ctx, audit.ListInput{
		Limit:       limit,
		EventType:   input.EventType,
		ActorID:     input.ActorID,
		SubjectType: input.SubjectType,
		SubjectID:   input.SubjectID,
		Outcome:     input.Outcome,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("Could not list audit log entries")
	}
	return &ListAuditLogOutput{Body: shared.NewListResource(auditEntryResources(entries))}, nil
}
