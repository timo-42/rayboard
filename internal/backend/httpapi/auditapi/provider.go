package auditapi

import (
	"github.com/timo-42/rayboard/internal/backend/audit"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type Provider struct {
	Audit         *audit.Store
	Authenticator shared.Authenticator
}
