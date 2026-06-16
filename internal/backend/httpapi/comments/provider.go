package comments

import (
	commentservice "github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type Provider struct {
	Comments      *commentservice.Service
	Authenticator shared.Authenticator
}
