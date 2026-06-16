package attachments

import (
	attachmentservice "github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type Provider struct {
	Attachments   *attachmentservice.Service
	Authenticator shared.Authenticator
}
