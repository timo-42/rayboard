package tickethooks

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type ProjectHooksInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
	Event     string `query:"event" doc:"Filter by hook event."`
	Phase     string `query:"phase" doc:"Filter by hook phase."`
	Limit     int    `query:"limit" doc:"Maximum number of hooks to return."`
	Offset    int    `query:"offset" doc:"Number of hooks to skip."`
}

type CreateProjectHookInput struct {
	shared.AuthInput
	ProjectID string `path:"project_id" doc:"Project ID."`
	Body      shared.ResourceInput[HookSpec]
}

type HookIDInput struct {
	shared.AuthInput
	HookID string `path:"hook_id" doc:"Ticket hook ID."`
}

type UpdateHookInput struct {
	shared.AuthInput
	HookID string `path:"hook_id" doc:"Ticket hook ID."`
	Body   shared.ResourceInput[UpdateHookSpec]
}

type PreviewHookInput struct {
	shared.AuthInput
	HookID string `path:"hook_id" doc:"Ticket hook ID."`
	Body   shared.ResourceInput[PreviewHookSpec]
}

type HookMetadata struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HookEngineSpec struct {
	Type       string `json:"type"`
	Script     string `json:"script,omitempty"`
	Prompt     string `json:"prompt,omitempty"`
	ProviderID string `json:"provider_id,omitempty"`
}

func (HookEngineSpec) Schema(_ huma.Registry) *huma.Schema {
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

type HookSpec struct {
	Name     string         `json:"name,omitempty"`
	Event    string         `json:"event,omitempty"`
	Phase    string         `json:"phase,omitempty"`
	Enabled  bool           `json:"enabled,omitempty"`
	Position int            `json:"position,omitempty"`
	Engine   HookEngineSpec `json:"engine"`
}

type UpdateHookSpec struct {
	Name     *string         `json:"name,omitempty"`
	Event    *string         `json:"event,omitempty"`
	Phase    *string         `json:"phase,omitempty"`
	Enabled  *bool           `json:"enabled,omitempty"`
	Position *int            `json:"position,omitempty"`
	Engine   *HookEngineSpec `json:"engine,omitempty"`
}

type HookStatus struct {
	LastError string `json:"last_error,omitempty"`
}

type PreviewHookSpec struct {
	Ticket  map[string]any `json:"ticket"`
	Current map[string]any `json:"current,omitempty"`
}

type PreviewHookMetadata struct {
	HookID    string `json:"hook_id"`
	ProjectID string `json:"project_id"`
}

type PreviewHookStatus struct {
	Output map[string]any `json:"output"`
	Ticket map[string]any `json:"ticket,omitempty"`
	Logs   []string       `json:"logs"`
	Error  string         `json:"error,omitempty"`
}

type HookResource = shared.Resource[HookMetadata, HookSpec, HookStatus]
type PreviewHookResource = shared.Resource[PreviewHookMetadata, PreviewHookSpec, PreviewHookStatus]

type ListHooksOutput = shared.ListOutput[HookResource]
type CreateHookOutput = shared.CreatedOutput[HookResource]

type HookOutput struct {
	Body HookResource
}

type PreviewHookOutput struct {
	Body PreviewHookResource
}

func (spec HookSpec) createInput(projectID string) tracker.CreateHookInput {
	return tracker.CreateHookInput{
		ProjectID: projectID,
		Name:      spec.Name,
		Event:     spec.Event,
		Phase:     spec.Phase,
		Enabled:   spec.Enabled,
		Position:  spec.Position,
		Engine:    spec.Engine.toService(),
	}
}

func (spec UpdateHookSpec) updateInput() tracker.UpdateHookInput {
	return tracker.UpdateHookInput{
		Name:     spec.Name,
		Event:    spec.Event,
		Phase:    spec.Phase,
		Enabled:  spec.Enabled,
		Position: spec.Position,
		Engine:   optionalEngineSpec(spec.Engine),
	}
}

func (spec PreviewHookSpec) previewInput() tracker.PreviewHookInput {
	return tracker.PreviewHookInput{
		Ticket:  spec.Ticket,
		Current: spec.Current,
	}
}

func (spec HookEngineSpec) toService() tracker.HookEngineSpec {
	return tracker.HookEngineSpec{
		Type:       spec.Type,
		Script:     spec.Script,
		Prompt:     spec.Prompt,
		ProviderID: spec.ProviderID,
	}
}

func optionalEngineSpec(spec *HookEngineSpec) *tracker.HookEngineSpec {
	if spec == nil {
		return nil
	}
	serviceSpec := spec.toService()
	return &serviceSpec
}

func hookEngineFromService(spec tracker.HookEngineSpec) HookEngineSpec {
	return HookEngineSpec{
		Type:       spec.Type,
		Script:     spec.Script,
		Prompt:     spec.Prompt,
		ProviderID: spec.ProviderID,
	}
}

func hookResource(hook tracker.Hook) HookResource {
	return HookResource{
		Metadata: HookMetadata{
			ID:        hook.ID,
			ProjectID: hook.ProjectID,
			CreatedAt: hook.CreatedAt,
			UpdatedAt: hook.UpdatedAt,
		},
		Spec: HookSpec{
			Name:     hook.Name,
			Event:    hook.Event,
			Phase:    hook.Phase,
			Enabled:  hook.Enabled,
			Position: hook.Position,
			Engine:   hookEngineFromService(hook.Engine),
		},
		Status: HookStatus{
			LastError: hook.LastError,
		},
	}
}

func hookResources(hooks []tracker.Hook) []HookResource {
	resources := make([]HookResource, 0, len(hooks))
	for _, hook := range hooks {
		resources = append(resources, hookResource(hook))
	}
	return resources
}

func previewResource(preview tracker.HookPreview) PreviewHookResource {
	ticket, _ := preview.Output["ticket"].(map[string]any)
	return PreviewHookResource{
		Metadata: PreviewHookMetadata{
			HookID:    preview.Hook.ID,
			ProjectID: preview.Hook.ProjectID,
		},
		Spec: PreviewHookSpec{
			Ticket:  objectFromPreviewInput(preview.Input, "ticket"),
			Current: objectFromPreviewInput(preview.Input, "current"),
		},
		Status: PreviewHookStatus{
			Output: preview.Output,
			Ticket: ticket,
			Logs:   preview.Logs,
			Error:  preview.Error,
		},
	}
}

func objectFromPreviewInput(input map[string]any, key string) map[string]any {
	value, ok := input[key]
	if !ok {
		return nil
	}
	object, _ := value.(map[string]any)
	return object
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
