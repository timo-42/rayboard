package shared

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/attachments"
	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/cronjobs"
	"github.com/timo-42/rayboard/internal/backend/notifications"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
	"github.com/timo-42/rayboard/internal/backend/search"
	"github.com/timo-42/rayboard/internal/backend/tracker"
	"github.com/timo-42/rayboard/internal/backend/webhooks"
)

type ErrorBody struct {
	Code    string            `json:"code" doc:"Stable machine-readable error code."`
	Message string            `json:"message" doc:"Human-readable error message."`
	Fields  map[string]string `json:"fields" doc:"Field-level validation messages, or an empty object when the error is not tied to fields."`
}

type ErrorResponse struct {
	status int
	Body   ErrorBody `json:"error"`
}

func init() {
	huma.NewError = func(status int, message string, errs ...error) huma.StatusError {
		return NewError(status, errorCode(status), message, detailFields(errs...))
	}
}

func NewError(status int, code string, message string, fields map[string]string) *ErrorResponse {
	if message == "" {
		message = http.StatusText(status)
	}
	if code == "" {
		code = errorCode(status)
	}
	return &ErrorResponse{
		status: status,
		Body: ErrorBody{
			Code:    code,
			Message: message,
			Fields:  nonNilFields(fields),
		},
	}
}

func (e *ErrorResponse) Error() string {
	if e == nil {
		return ""
	}
	return e.Body.Message
}

func (e *ErrorResponse) GetStatus() int {
	if e == nil {
		return http.StatusInternalServerError
	}
	return e.status
}

func ErrorResponseSchema() *huma.Schema {
	return &huma.Schema{
		Type:     huma.TypeObject,
		Required: []string{"error"},
		Properties: map[string]*huma.Schema{
			"error": {
				Type:     huma.TypeObject,
				Required: []string{"code", "message", "fields"},
				Properties: map[string]*huma.Schema{
					"code":    {Type: huma.TypeString, Description: "Stable machine-readable error code."},
					"message": {Type: huma.TypeString, Description: "Human-readable error message."},
					"fields": {
						Type:                 huma.TypeObject,
						Description:          "Field-level validation messages, or an empty object when the error is not tied to fields.",
						AdditionalProperties: &huma.Schema{Type: huma.TypeString},
					},
				},
				AdditionalProperties: false,
			},
		},
		AdditionalProperties: false,
	}
}

func validationError(message string, fields map[string]string) error {
	return NewError(http.StatusBadRequest, "validation_failed", message, fields)
}

func errorCode(status int) string {
	switch status {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return "validation_failed"
	case http.StatusUnauthorized:
		return "unauthenticated"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusRequestEntityTooLarge:
		return "payload_too_large"
	case http.StatusBadGateway:
		return "bad_gateway"
	default:
		if status >= 500 {
			return "internal_error"
		}
		return "request_failed"
	}
}

func detailFields(errs ...error) map[string]string {
	if len(errs) == 0 {
		return nil
	}
	fields := map[string]string{}
	for index, err := range errs {
		if err == nil {
			continue
		}
		if detailer, ok := err.(huma.ErrorDetailer); ok {
			detail := detailer.ErrorDetail()
			if detail != nil && strings.TrimSpace(detail.Location) != "" {
				fields[detail.Location] = strings.TrimSpace(detail.Message)
				continue
			}
		}
		fields[fmt.Sprintf("error_%d", index+1)] = err.Error()
	}
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func nonNilFields(fields map[string]string) map[string]string {
	if len(fields) == 0 {
		return map[string]string{}
	}
	cleaned := map[string]string{}
	for key, value := range fields {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			cleaned[key] = value
		}
	}
	if len(cleaned) == 0 {
		return map[string]string{}
	}
	return cleaned
}

func AuthServiceError(err error) error {
	var validation *auth.ValidationError
	switch {
	case errors.As(err, &validation):
		return validationError(validation.Message, validation.Fields)
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
		return validationError(validation.Message, validation.Fields)
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
		return validationError(validation.Message, validation.Fields)
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
		return validationError(validation.Message, validation.Fields)
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
		return validationError(validation.Message, validation.Fields)
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
		return validationError(validation.Message, validation.Fields)
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
		return validationError(validation.Message, validation.Fields)
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
	case errors.Is(err, automation.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, notifications.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, automation.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	case errors.Is(err, notifications.ErrDelivery):
		return huma.Error502BadGateway("Notification delivery failed")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}

func WebhookError(err error) error {
	switch {
	case errors.Is(err, webhooks.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, webhooks.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}
