package authapi

import (
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type Provider struct {
	Auth          *auth.Service
	Authenticator shared.Authenticator
}
