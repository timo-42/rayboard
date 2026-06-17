package enginesapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/engines"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodPost, "/api/engines/test", "Engines", "Test automation engine"), provider.testEngine)
}

func (provider Provider) testEngine(ctx context.Context, input *TestEngineInput) (*TestEngineOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	run, err := provider.Engines.Test(ctx, principal, input.Body.Spec.testInput())
	if err != nil && run.ID == "" {
		return nil, engineError(err)
	}
	return &TestEngineOutput{Body: testEngineResource(run, input.Body.Spec)}, nil
}

func engineError(err error) error {
	var validation *engines.ValidationError
	switch {
	case errors.As(err, &validation):
		return shared.NewError(http.StatusBadRequest, "validation_failed", validation.Message, validation.Fields)
	case errors.Is(err, engines.ErrValidation):
		return huma.Error400BadRequest("Validation failed")
	case errors.Is(err, engines.ErrNotFound):
		return huma.Error404NotFound("Resource was not found")
	case errors.Is(err, authz.ErrForbidden):
		return huma.Error403Forbidden("Permission denied")
	default:
		return huma.Error500InternalServerError("Request failed")
	}
}
