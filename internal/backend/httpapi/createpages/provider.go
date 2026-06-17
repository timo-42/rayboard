package createpages

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type Provider struct {
	CreatePages   *tracker.CreatePageService
	Authenticator shared.Authenticator
}
