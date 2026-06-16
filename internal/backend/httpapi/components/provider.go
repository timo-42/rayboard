package components

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type Provider struct {
	Tracker       *tracker.Service
	Authenticator shared.Authenticator
}
