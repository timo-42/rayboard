package notificationsapi

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/notifications"
)

type Provider struct {
	Notifications *notifications.Service
	Authenticator shared.Authenticator
}
