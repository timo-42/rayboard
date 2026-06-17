package enginesapi

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/engines"
	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
)

type TestEngineInput struct {
	shared.AuthInput
	Body shared.ResourceInput[TestEngineSpec]
}

type TestEngineOutput struct {
	Body TestEngineResource
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

type TestEngineSpec struct {
	Surface     string         `json:"surface,omitempty"`
	ProjectID   string         `json:"project_id,omitempty"`
	ActorUserID string         `json:"actor_user_id,omitempty"`
	Engine      EngineSpec     `json:"engine"`
	Input       map[string]any `json:"input,omitempty"`
}

type TestEngineMetadata struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type TestEngineStatus struct {
	State      string         `json:"state"`
	Output     map[string]any `json:"output"`
	Error      string         `json:"error,omitempty"`
	StartedAt  *time.Time     `json:"started_at,omitempty"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
}

type TestEngineResource = shared.Resource[TestEngineMetadata, TestEngineSpec, TestEngineStatus]

func (spec TestEngineSpec) testInput() engines.TestInput {
	return engines.TestInput{
		ProjectID:   spec.ProjectID,
		ActorUserID: spec.ActorUserID,
		Surface:     spec.Surface,
		Input:       spec.Input,
		Engine: engines.EngineSpec{
			Type:       spec.Engine.Type,
			Script:     spec.Engine.Script,
			Prompt:     spec.Engine.Prompt,
			ProviderID: spec.Engine.ProviderID,
		},
	}
}

func testEngineResource(run automation.Run, spec TestEngineSpec) TestEngineResource {
	sanitized := spec
	sanitized.Engine.Script = ""
	sanitized.Engine.Prompt = ""
	if sanitized.Input == nil {
		sanitized.Input = map[string]any{}
	}
	return TestEngineResource{
		Metadata: TestEngineMetadata{
			ID:        run.ID,
			CreatedAt: run.CreatedAt,
		},
		Spec: sanitized,
		Status: TestEngineStatus{
			State:      run.Status,
			Output:     run.Output,
			Error:      run.Error,
			StartedAt:  run.StartedAt,
			FinishedAt: run.FinishedAt,
		},
	}
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
