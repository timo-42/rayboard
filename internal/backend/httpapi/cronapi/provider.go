package cronapi

import (
	"github.com/timo-42/rayboard/internal/backend/cronjobs"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type Provider struct {
	Cron          *cronjobs.Service
	Authenticator shared.Authenticator
}
