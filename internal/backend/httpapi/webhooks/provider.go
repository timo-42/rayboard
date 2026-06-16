package webhooksapi

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/webhooks"
)

type Provider struct {
	Webhooks      *webhooks.Service
	Authenticator shared.Authenticator
}
