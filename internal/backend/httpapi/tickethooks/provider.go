package tickethooks

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type Provider struct {
	Hooks         *tracker.HookService
	Authenticator shared.Authenticator
}
