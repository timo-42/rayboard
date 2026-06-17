package webhooksapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/webhooks"
)

func Register(api huma.API, provider Provider) {
	huma.Register(api, shared.Operation(http.MethodGet, "/api/projects/{project_id}/webhooks", "Webhooks", "List project webhooks"), provider.listProjectWebhooks)
	huma.Register(api, operation(http.MethodPost, "/api/projects/{project_id}/webhooks", "Webhooks", "Create project webhook", http.StatusCreated), provider.createProjectWebhook)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/webhook-definitions/{webhook_id}", "Webhooks", "Get webhook"), provider.getWebhook)
	huma.Register(api, shared.Operation(http.MethodPatch, "/api/webhook-definitions/{webhook_id}", "Webhooks", "Update webhook"), provider.updateWebhook)
	huma.Register(api, operation(http.MethodDelete, "/api/webhook-definitions/{webhook_id}", "Webhooks", "Delete webhook", http.StatusNoContent), provider.deleteWebhook)
	huma.Register(api, shared.Operation(http.MethodPost, "/api/webhook-definitions/{webhook_id}/rotate-token", "Webhooks", "Rotate incoming webhook token"), provider.rotateWebhookToken)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/webhook-definitions/{webhook_id}/runs", "Webhooks", "List webhook runs"), provider.listWebhookRuns)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/webhook-definitions/{webhook_id}/deliveries", "Webhook Deliveries", "List outgoing webhook deliveries"), provider.listOutgoingWebhookDeliveries)
	huma.Register(api, shared.Operation(http.MethodGet, "/api/webhook-deliveries/{delivery_id}", "Webhook Deliveries", "Get outgoing webhook delivery"), provider.getOutgoingWebhookDelivery)
	huma.Register(api, shared.PublicOperation(http.MethodPost, "/api/webhooks/incoming/{webhook_id}", "Webhooks", "Receive incoming webhook"), provider.receiveIncomingWebhook)
}

func (provider Provider) listProjectWebhooks(ctx context.Context, input *ProjectWebhooksInput) (*ListWebhooksOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	hooks, err := provider.Webhooks.List(ctx, principal, webhooks.ListInput{
		ProjectID: input.ProjectID,
		Direction: input.Direction,
		Limit:     input.Limit,
		Offset:    input.Offset,
	})
	if err != nil {
		return nil, shared.WebhookError(err)
	}
	return &ListWebhooksOutput{Body: shared.NewListResource[WebhookResource](webhookResources(hooks))}, nil
}

func (provider Provider) createProjectWebhook(ctx context.Context, input *CreateProjectWebhookInput) (*CreateWebhookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	hook, err := provider.Webhooks.Create(ctx, principal, input.Body.Spec.createInput(input.ProjectID))
	if err != nil {
		return nil, shared.WebhookError(err)
	}
	return &CreateWebhookOutput{Body: createdWebhookResource(hook)}, nil
}

func (provider Provider) getWebhook(ctx context.Context, input *WebhookIDInput) (*WebhookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	hook, err := provider.Webhooks.Get(ctx, principal, input.WebhookID)
	if err != nil {
		return nil, shared.WebhookError(err)
	}
	return &WebhookOutput{Body: webhookResource(hook)}, nil
}

func (provider Provider) updateWebhook(ctx context.Context, input *UpdateWebhookInput) (*WebhookOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	hook, err := provider.Webhooks.Update(ctx, principal, input.WebhookID, input.Body.Spec.updateInput())
	if err != nil {
		return nil, shared.WebhookError(err)
	}
	return &WebhookOutput{Body: webhookResource(hook)}, nil
}

func (provider Provider) deleteWebhook(ctx context.Context, input *WebhookIDInput) (*shared.EmptyOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	if err := provider.Webhooks.Delete(ctx, principal, input.WebhookID); err != nil {
		return nil, shared.WebhookError(err)
	}
	return &shared.EmptyOutput{}, nil
}

func (provider Provider) rotateWebhookToken(ctx context.Context, input *WebhookIDInput) (*RotateWebhookTokenOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, true)
	if err != nil {
		return nil, err
	}
	hook, err := provider.Webhooks.RotateIncomingToken(ctx, principal, input.WebhookID)
	if err != nil {
		return nil, shared.WebhookError(err)
	}
	return &RotateWebhookTokenOutput{Body: createdWebhookResource(hook)}, nil
}

func (provider Provider) listWebhookRuns(ctx context.Context, input *ListWebhookRunsInput) (*ListWebhookRunsOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	runs, err := provider.Webhooks.ListRuns(ctx, principal, input.WebhookID, input.Limit, input.Offset)
	if err != nil {
		return nil, shared.WebhookError(err)
	}
	return &ListWebhookRunsOutput{Body: shared.NewListResource[WebhookRunResource](runResources(runs))}, nil
}

func (provider Provider) listOutgoingWebhookDeliveries(ctx context.Context, input *ListOutgoingWebhookDeliveriesInput) (*ListOutgoingWebhookDeliveriesOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	deliveries, err := provider.Webhooks.ListOutgoingDeliveries(ctx, principal, input.WebhookID, input.Limit, input.Offset)
	if err != nil {
		return nil, shared.WebhookError(err)
	}
	return &ListOutgoingWebhookDeliveriesOutput{Body: shared.NewListResource[OutgoingWebhookDeliveryResource](outgoingDeliveryResources(deliveries))}, nil
}

func (provider Provider) getOutgoingWebhookDelivery(ctx context.Context, input *OutgoingWebhookDeliveryIDInput) (*OutgoingWebhookDeliveryOutput, error) {
	ctx, principal, _, err := provider.Authenticator.Authenticate(ctx, input.AuthInput, false)
	if err != nil {
		return nil, err
	}
	delivery, err := provider.Webhooks.GetOutgoingDelivery(ctx, principal, input.DeliveryID)
	if err != nil {
		return nil, shared.WebhookError(err)
	}
	return &OutgoingWebhookDeliveryOutput{Body: outgoingDeliveryResource(delivery)}, nil
}

func (provider Provider) receiveIncomingWebhook(ctx context.Context, input *IncomingWebhookInput) (*IncomingWebhookOutput, error) {
	result, err := provider.Webhooks.ReceiveIncoming(ctx, input.WebhookID, bearerToken(input.Authorization), webhooks.IncomingInput{
		Headers: input.Body.Spec.Headers,
		Query:   input.Body.Spec.Query,
		Payload: input.Body.Spec.Payload,
	})
	if err != nil {
		return nil, shared.WebhookError(err)
	}
	return &IncomingWebhookOutput{Body: incomingWebhookResource(input.Body.Spec, result)}, nil
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
		return ""
	}
	return header[len(prefix):]
}

func operation(method string, path string, tag string, summary string, status int) huma.Operation {
	op := shared.OperationWithStatus(method, path, tag, summary, status)
	op.OperationID = ""
	return op
}
