package settingsapi

import (
	"context"

	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/settings"
)

type Provider struct {
	Settings      *settings.Service
	Audit         *audit.Store
	Authenticator shared.Authenticator
}

func (provider Provider) recordAudit(ctx context.Context, principal authz.Principal, input audit.RecordInput) error {
	if provider.Audit == nil {
		return nil
	}
	actorID := principal.ActorUserID
	if actorID == "" {
		actorID = principal.UserID
	}
	input.ActorID = actorID
	input.AuthKind = principal.AuthKind
	_, err := provider.Audit.Record(ctx, input)
	return err
}
