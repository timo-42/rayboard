package searchapi

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/search"
)

type Provider struct {
	Search        *search.Service
	Authenticator shared.Authenticator
}
