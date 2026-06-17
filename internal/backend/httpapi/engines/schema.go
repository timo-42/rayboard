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
	Context     map[string]any `json:"context,omitempty"`
	Input       map[string]any `json:"input,omitempty"`
	DryRun      bool           `json:"dry_run,omitempty"`
}

type TestEngineMetadata struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type TestEngineStatus struct {
	State          string           `json:"state"`
	Output         map[string]any   `json:"output"`
	ActionPreviews []map[string]any `json:"action_previews,omitempty"`
	Logs           []string         `json:"logs,omitempty"`
	DurationMillis int64            `json:"duration_millis,omitempty"`
	Engine         map[string]any   `json:"engine,omitempty"`
	Error          string           `json:"error,omitempty"`
	StartedAt      *time.Time       `json:"started_at,omitempty"`
	FinishedAt     *time.Time       `json:"finished_at,omitempty"`
}

type TestEngineResource = shared.Resource[TestEngineMetadata, TestEngineSpec, TestEngineStatus]

func (spec TestEngineSpec) testInput() engines.TestInput {
	return engines.TestInput{
		ProjectID:   spec.ProjectID,
		ActorUserID: spec.ActorUserID,
		Surface:     spec.Surface,
		Context:     spec.Context,
		Input:       spec.Input,
		DryRun:      true,
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
	if sanitized.Context == nil {
		sanitized.Context = map[string]any{}
	}
	sanitized.DryRun = true
	return TestEngineResource{
		Metadata: TestEngineMetadata{
			ID:        run.ID,
			CreatedAt: run.CreatedAt,
		},
		Spec: sanitized,
		Status: TestEngineStatus{
			State:          run.Status,
			Output:         runOutput(run.Output),
			ActionPreviews: actionPreviews(run.Output),
			Logs:           runLogs(run.Output),
			DurationMillis: runDurationMillis(run),
			Engine:         engineMetadata(run),
			Error:          run.Error,
			StartedAt:      run.StartedAt,
			FinishedAt:     run.FinishedAt,
		},
	}
}

func runOutput(output map[string]any) map[string]any {
	if value, ok := output["output"].(map[string]any); ok {
		return value
	}
	return map[string]any{}
}

func actionPreviews(output map[string]any) []map[string]any {
	raw, ok := output["action_previews"].([]any)
	if !ok {
		return nil
	}
	previews := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if preview, ok := item.(map[string]any); ok {
			previews = append(previews, preview)
		}
	}
	return previews
}

func runLogs(output map[string]any) []string {
	raw, ok := output["logs"].([]any)
	if !ok {
		return nil
	}
	logs := make([]string, 0, len(raw))
	for _, item := range raw {
		if logLine, ok := item.(string); ok {
			logs = append(logs, logLine)
		}
	}
	return logs
}

func runDurationMillis(run automation.Run) int64 {
	if run.StartedAt == nil || run.FinishedAt == nil {
		return 0
	}
	return run.FinishedAt.Sub(*run.StartedAt).Milliseconds()
}

func engineMetadata(run automation.Run) map[string]any {
	metadata := map[string]any{}
	if engine, ok := run.Input["engine"].(string); ok && engine != "" {
		metadata["type"] = engine
	}
	if input, ok := run.Input["input"].(map[string]any); ok {
		if context, ok := input["context"].(map[string]any); ok {
			if dryRun, ok := context["dry_run"].(bool); ok {
				metadata["dry_run"] = dryRun
			}
			if surface, ok := context["surface"].(string); ok && surface != "" {
				metadata["surface"] = surface
			}
		}
	}
	return metadata
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
