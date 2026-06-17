package enginesapi

import (
	"github.com/timo-42/rayboard/internal/backend/engines"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type Provider struct {
	Engines       *engines.Service
	Authenticator shared.Authenticator
}
