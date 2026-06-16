package versions

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type VersionIDInput struct {
	shared.AuthInput
	VersionID string `path:"version_id"`
}

type UpdateVersionInput struct {
	shared.AuthInput
	VersionID string `path:"version_id"`
	Body      tracker.UpdateVersionInput
}

type VersionOutput struct {
	Body tracker.Version
}
