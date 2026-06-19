package webhooksapi

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/webhooks"
)

type ProjectWebhooksInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
	Direction string `query:"direction" doc:"Filter by webhook direction."`
	Limit     int    `query:"limit" doc:"Maximum number of webhooks to return."`
	Offset    int    `query:"offset" doc:"Number of webhooks to skip."`
}

type CreateProjectWebhookInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
	Body      shared.ResourceInput[CreateWebhookSpec]
}

type WebhookIDInput struct {
	shared.AuthInput
	WebhookID string `path:"webhook_id" doc:"Webhook ID."`
}

type ListWebhookRunsInput struct {
	shared.AuthInput
	WebhookID string `path:"webhook_id" doc:"Webhook ID."`
	Limit     int    `query:"limit" doc:"Maximum number of runs to return."`
	Offset    int    `query:"offset" doc:"Number of runs to skip."`
}

type ListOutgoingWebhookDeliveriesInput struct {
	shared.AuthInput
	WebhookID string `path:"webhook_id" doc:"Webhook ID."`
	Limit     int    `query:"limit" doc:"Maximum number of deliveries to return."`
	Offset    int    `query:"offset" doc:"Number of deliveries to skip."`
}

type OutgoingWebhookDeliveryIDInput struct {
	shared.AuthInput
	DeliveryID string `path:"delivery_id" doc:"Outgoing webhook delivery ID."`
}

type IncomingWebhookInput struct {
	WebhookID     string `path:"webhook_id" doc:"Incoming webhook ID."`
	Authorization string `header:"Authorization" doc:"Bearer webhook token."`
	Body          shared.ResourceInput[IncomingWebhookSpec]
}

type UpdateWebhookInput struct {
	shared.AuthInput
	WebhookID string `path:"webhook_id" doc:"Webhook ID."`
	Body      shared.ResourceInput[UpdateWebhookSpec]
}

type WebhookMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EngineSpec struct {
	Type       string `json:"type"`
	Script     string `json:"script,omitempty"`
	Prompt     string `json:"prompt,omitempty"`
	ProviderID string `json:"provider_id,omitempty"`
}

func (EngineSpec) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		OneOf: []*huma.Schema{
			engineVariantSchema("lua", []string{"type", "script"}, map[string]*huma.Schema{
				"type":   {Type: huma.TypeString, Enum: []any{"lua"}},
				"script": {Type: huma.TypeString, Description: "Lua script source."},
			}),
			engineVariantSchema("ai", []string{"type", "prompt", "provider_id"}, map[string]*huma.Schema{
				"type":        {Type: huma.TypeString, Enum: []any{"ai"}},
				"prompt":      {Type: huma.TypeString, Description: "AI prompt sent to the selected OpenRouter provider."},
				"provider_id": {Type: huma.TypeString, Description: "Admin-managed OpenRouter provider configuration ID."},
			}),
		},
		Discriminator: &huma.Discriminator{PropertyName: "type"},
	}
}

type WebhookSpec struct {
	Name        string     `json:"name,omitempty"`
	Direction   string     `json:"direction,omitempty"`
	Enabled     bool       `json:"enabled,omitempty"`
	ActorUserID string     `json:"actor_user_id,omitempty"`
	EventTypes  []string   `json:"event_types,omitempty"`
	Engine      EngineSpec `json:"engine"`
}

type CreateWebhookSpec = WebhookSpec

type UpdateWebhookSpec struct {
	Name        *string     `json:"name,omitempty"`
	Direction   *string     `json:"direction,omitempty"`
	Enabled     *bool       `json:"enabled,omitempty"`
	ActorUserID *string     `json:"actor_user_id,omitempty"`
	EventTypes  *[]string   `json:"event_types,omitempty"`
	Engine      *EngineSpec `json:"engine,omitempty"`
}

type IncomingWebhookSpec struct {
	Headers map[string]string `json:"headers,omitempty"`
	Query   map[string]string `json:"query,omitempty"`
	Payload map[string]any    `json:"payload,omitempty"`
}

type WebhookStatus struct {
	TokenSet       bool       `json:"token_set"`
	TokenRotatedAt *time.Time `json:"token_rotated_at,omitempty"`
	LastError      string     `json:"last_error,omitempty"`
}

type CreatedWebhookStatus struct {
	WebhookStatus
	Token string `json:"token,omitempty"`
}

type WebhookRunMetadata struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type WebhookRunSpec struct {
	TriggerType string         `json:"trigger_type"`
	TriggerRef  string         `json:"trigger_ref,omitempty"`
	ProjectID   string         `json:"project_id,omitempty"`
	TicketID    string         `json:"ticket_id,omitempty"`
	Input       map[string]any `json:"input"`
}

type WebhookRunStatus struct {
	State      string         `json:"state"`
	Output     map[string]any `json:"output"`
	Error      string         `json:"error,omitempty"`
	StartedAt  *time.Time     `json:"started_at,omitempty"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
}

type OutgoingWebhookDeliveryMetadata struct {
	ID             string    `json:"id"`
	WebhookID      string    `json:"webhook_id,omitempty"`
	WebhookName    string    `json:"webhook_name,omitempty"`
	DomainEventID  string    `json:"domain_event_id,omitempty"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	ProjectID      string    `json:"project_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type OutgoingWebhookDeliverySpec struct {
	EventType   string         `json:"event_type"`
	SubjectType string         `json:"subject_type,omitempty"`
	SubjectID   string         `json:"subject_id,omitempty"`
	Payload     map[string]any `json:"payload"`
	MaxAttempts int            `json:"max_attempts"`
}

type OutgoingWebhookDeliveryStatus struct {
	State         string     `json:"state"`
	AttemptCount  int        `json:"attempt_count"`
	NextAttemptAt *time.Time `json:"next_attempt_at,omitempty"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
}

type WebhookResource = shared.Resource[WebhookMetadata, WebhookSpec, WebhookStatus]
type CreatedWebhookResource = shared.Resource[WebhookMetadata, WebhookSpec, CreatedWebhookStatus]
type IncomingWebhookResource = shared.Resource[WebhookRunMetadata, IncomingWebhookSpec, WebhookRunStatus]
type WebhookRunResource = shared.Resource[WebhookRunMetadata, WebhookRunSpec, WebhookRunStatus]
type OutgoingWebhookDeliveryResource = shared.Resource[OutgoingWebhookDeliveryMetadata, OutgoingWebhookDeliverySpec, OutgoingWebhookDeliveryStatus]

type ListWebhooksOutput = shared.ListOutput[WebhookResource]
type CreateWebhookOutput = shared.CreatedOutput[CreatedWebhookResource]
type ListWebhookRunsOutput = shared.ListOutput[WebhookRunResource]
type ListOutgoingWebhookDeliveriesOutput = shared.ListOutput[OutgoingWebhookDeliveryResource]

type WebhookOutput struct {
	Body WebhookResource
}

type RotateWebhookTokenOutput struct {
	Body CreatedWebhookResource
}

type IncomingWebhookOutput struct {
	Body IncomingWebhookResource
}

type OutgoingWebhookDeliveryOutput struct {
	Body OutgoingWebhookDeliveryResource
}

func (spec WebhookSpec) createInput(projectID string) webhooks.CreateInput {
	return webhooks.CreateInput{
		ProjectID:   projectID,
		Name:        spec.Name,
		Direction:   spec.Direction,
		Enabled:     spec.Enabled,
		ActorUserID: spec.ActorUserID,
		EventTypes:  spec.EventTypes,
		Engine:      spec.Engine.toService(),
	}
}

func (spec UpdateWebhookSpec) updateInput() webhooks.UpdateInput {
	var engine *webhooks.EngineSpec
	if spec.Engine != nil {
		value := spec.Engine.toService()
		engine = &value
	}
	return webhooks.UpdateInput{
		Name:        spec.Name,
		Direction:   spec.Direction,
		Enabled:     spec.Enabled,
		ActorUserID: spec.ActorUserID,
		EventTypes:  spec.EventTypes,
		Engine:      engine,
	}
}

func (spec EngineSpec) toService() webhooks.EngineSpec {
	return webhooks.EngineSpec{
		Type:       spec.Type,
		Script:     spec.Script,
		Prompt:     spec.Prompt,
		ProviderID: spec.ProviderID,
	}
}

func engineFromService(engine webhooks.EngineSpec) EngineSpec {
	return EngineSpec{
		Type:       engine.Type,
		Script:     engine.Script,
		Prompt:     engine.Prompt,
		ProviderID: engine.ProviderID,
	}
}

func webhookResource(hook webhooks.Webhook) WebhookResource {
	return WebhookResource{
		Metadata: WebhookMetadata{
			ID:        hook.ID,
			ProjectID: hook.ProjectID,
			CreatedAt: hook.CreatedAt,
			UpdatedAt: hook.UpdatedAt,
		},
		Spec: WebhookSpec{
			Name:        hook.Name,
			Direction:   hook.Direction,
			Enabled:     hook.Enabled,
			ActorUserID: hook.ActorUserID,
			EventTypes:  hook.EventTypes,
			Engine:      engineFromService(hook.Engine),
		},
		Status: WebhookStatus{
			TokenSet:       hook.TokenSet,
			TokenRotatedAt: hook.TokenRotatedAt,
			LastError:      hook.LastError,
		},
	}
}

func createdWebhookResource(hook webhooks.CreatedWebhook) CreatedWebhookResource {
	resource := webhookResource(hook.Webhook)
	return CreatedWebhookResource{
		Metadata: resource.Metadata,
		Spec:     resource.Spec,
		Status: CreatedWebhookStatus{
			WebhookStatus: resource.Status,
			Token:         hook.Token,
		},
	}
}

func webhookResources(hooks []webhooks.Webhook) []WebhookResource {
	resources := make([]WebhookResource, 0, len(hooks))
	for _, hook := range hooks {
		resources = append(resources, webhookResource(hook))
	}
	return resources
}

func incomingWebhookResource(input IncomingWebhookSpec, result webhooks.IncomingResult) IncomingWebhookResource {
	return IncomingWebhookResource{
		Metadata: WebhookRunMetadata{
			ID:        result.Run.ID,
			CreatedAt: result.Run.CreatedAt,
		},
		Spec: input,
		Status: WebhookRunStatus{
			State:      result.Run.Status,
			Output:     result.Run.Output,
			Error:      result.Run.Error,
			StartedAt:  result.Run.StartedAt,
			FinishedAt: result.Run.FinishedAt,
		},
	}
}

func runResource(run automation.Run) WebhookRunResource {
	return WebhookRunResource{
		Metadata: WebhookRunMetadata{
			ID:        run.ID,
			CreatedAt: run.CreatedAt,
		},
		Spec: WebhookRunSpec{
			TriggerType: run.TriggerType,
			TriggerRef:  run.TriggerRef,
			ProjectID:   run.ProjectID,
			TicketID:    run.TicketID,
			Input:       run.Input,
		},
		Status: WebhookRunStatus{
			State:      run.Status,
			Output:     run.Output,
			Error:      run.Error,
			StartedAt:  run.StartedAt,
			FinishedAt: run.FinishedAt,
		},
	}
}

func runResources(runs []automation.Run) []WebhookRunResource {
	resources := make([]WebhookRunResource, 0, len(runs))
	for _, run := range runs {
		resources = append(resources, runResource(run))
	}
	return resources
}

func outgoingDeliveryResource(delivery webhooks.OutgoingDelivery) OutgoingWebhookDeliveryResource {
	return OutgoingWebhookDeliveryResource{
		Metadata: OutgoingWebhookDeliveryMetadata{
			ID:             delivery.ID,
			WebhookID:      delivery.WebhookID,
			WebhookName:    delivery.WebhookName,
			DomainEventID:  delivery.DomainEventID,
			IdempotencyKey: delivery.IdempotencyKey,
			ProjectID:      delivery.ProjectID,
			CreatedAt:      delivery.CreatedAt,
			UpdatedAt:      delivery.UpdatedAt,
		},
		Spec: OutgoingWebhookDeliverySpec{
			EventType:   delivery.EventType,
			SubjectType: delivery.SubjectType,
			SubjectID:   delivery.SubjectID,
			Payload:     delivery.Payload,
			MaxAttempts: delivery.MaxAttempts,
		},
		Status: OutgoingWebhookDeliveryStatus{
			State:         delivery.Status,
			AttemptCount:  delivery.AttemptCount,
			NextAttemptAt: delivery.NextAttemptAt,
			LastAttemptAt: delivery.LastAttemptAt,
			DeliveredAt:   delivery.DeliveredAt,
			LastError:     delivery.LastError,
		},
	}
}

func outgoingDeliveryResources(deliveries []webhooks.OutgoingDelivery) []OutgoingWebhookDeliveryResource {
	resources := make([]OutgoingWebhookDeliveryResource, 0, len(deliveries))
	for _, delivery := range deliveries {
		resources = append(resources, outgoingDeliveryResource(delivery))
	}
	return resources
}

func engineVariantSchema(title string, required []string, properties map[string]*huma.Schema) *huma.Schema {
	return &huma.Schema{
		Type:                 huma.TypeObject,
		Title:                title,
		Required:             required,
		Properties:           properties,
		AdditionalProperties: false,
	}
}
