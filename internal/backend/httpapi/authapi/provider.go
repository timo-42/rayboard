package authapi

import (
	"context"

	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type Provider struct {
	Auth          *auth.Service
	Audit         *audit.Store
	Authenticator shared.Authenticator
}

func (provider Provider) recordAudit(ctx context.Context, input audit.RecordInput) error {
	if provider.Audit == nil {
		return nil
	}
	_, err := provider.Audit.Record(ctx, input)
	return err
}

func auditActor(principal authz.Principal) (string, authz.AuthKind) {
	actorID := principal.ActorUserID
	if actorID == "" {
		actorID = principal.UserID
	}
	return actorID, principal.AuthKind
}
