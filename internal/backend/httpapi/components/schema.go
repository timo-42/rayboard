package components

import (
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type ComponentIDInput struct {
	shared.AuthInput
	ComponentID string `path:"component_id"`
}

type UpdateComponentInput struct {
	shared.AuthInput
	ComponentID string `path:"component_id"`
	Body        tracker.UpdateComponentInput
}

type ComponentOutput struct {
	Body tracker.Component
}
