package sprints

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type SprintIDInput struct {
	shared.AuthInput
	SprintID string `path:"sprint_id"`
}

type UpdateSprintInput struct {
	shared.AuthInput
	SprintID string `path:"sprint_id"`
	Body     tracker.UpdateSprintInput
}

type CompleteSprintInput struct {
	shared.AuthInput
	SprintID string `path:"sprint_id"`
}

type SprintOutput struct {
	Body tracker.Sprint
}
