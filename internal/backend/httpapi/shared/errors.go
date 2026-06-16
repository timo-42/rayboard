package shared

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/cronjobs"
	"github.com/timo-42/rayboard/internal/backend/notifications"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

func AuthServiceError(err error) error {
	var validation *auth.ValidationError
	switch {
	case errors.As(err, &validation):
		return huma.Error400BadRequest(validation.Message)
	case errors.Is(err, auth.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, auth.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, auth.ErrConflict):
		return huma.Error409Conflict("Resource already exists")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}

func TrackerError(err error) error {
	var validation *tracker.ValidationError
	switch {
	case errors.As(err, &validation):
		return huma.Error400BadRequest(validation.Message)
	case errors.Is(err, tracker.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, tracker.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, tracker.ErrConflict):
		return huma.Error409Conflict("Resource conflict")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}

func SearchError(err error) error {
	var validation *search.ValidationError
	switch {
	case errors.As(err, &validation):
		return huma.Error400BadRequest(validation.Message)
	case errors.Is(err, search.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, search.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, search.ErrConflict):
		return huma.Error409Conflict("Resource already exists")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}

func CommentError(err error) error {
	var validation *comments.ValidationError
	switch {
	case errors.As(err, &validation):
		return huma.Error400BadRequest(validation.Message)
	case errors.Is(err, comments.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, comments.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}

func AttachmentError(err error) error {
	var validation *attachments.ValidationError
	switch {
	case errors.As(err, &validation):
		return huma.Error400BadRequest(validation.Message)
	case errors.Is(err, attachments.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, attachments.ErrTooLarge):
		return huma.Error413RequestEntityTooLarge("Attachment is too large")
	case errors.Is(err, attachments.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}

func CronError(err error) error {
	var validation *cronjobs.ValidationError
	switch {
	case errors.As(err, &validation):
		return huma.Error400BadRequest(validation.Message)
	case errors.Is(err, cronjobs.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, cronjobs.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}

func OpenRouterError(err error) error {
	var validation *openrouter.ValidationError
	switch {
	case errors.As(err, &validation):
		return huma.Error400BadRequest(validation.Message)
	case errors.Is(err, openrouter.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, openrouter.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, openrouter.ErrConflict):
		return huma.Error409Conflict("Resource already exists")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}

func NotificationError(err error) error {
	switch {
	case errors.Is(err, notifications.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, notifications.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}
