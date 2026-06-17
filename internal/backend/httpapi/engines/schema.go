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
	Type         string `json:"type"`
	Script       string `json:"script,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	ProviderID   string `json:"provider_id,omitempty"`
	ModuleBase64 string `json:"module_base64,omitempty"`
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
			engineVariantSchema("wasm", []string{"type", "module_base64"}, map[string]*huma.Schema{
				"type":          {Type: huma.TypeString, Enum: []any{"wasm"}},
				"module_base64": {Type: huma.TypeString, Description: "Base64-encoded WebAssembly module using the Rayboard WASI stdin/stdout JSON contract."},
			}),
		},
		Discriminator: &huma.Discriminator{PropertyName: "type"},
	}
}

type TestEngineSpec struct {
	Surface      string         `json:"surface,omitempty" enum:"scratch,cron,ticket_hook_before,ticket_hook_after,custom_create_page,incoming_webhook,outgoing_webhook,notification_hook" default:"scratch" doc:"Automation surface contract to test; scratch is a generic playground."`
	ProjectID    string         `json:"project_id,omitempty"`
	ActorUserID  string         `json:"actor_user_id,omitempty"`
	Engine       EngineSpec     `json:"engine"`
	Context      map[string]any `json:"context,omitempty"`
	Input        map[string]any `json:"input,omitempty"`
	DryRun       bool           `json:"dry_run,omitempty"`
	ValidateOnly bool           `json:"validate_only,omitempty" doc:"Validate engine shape, permissions, provider/runtime availability, and surface compatibility without executing code."`
}

type TestEngineMetadata struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type TestEngineStatus struct {
	State          string           `json:"state"`
	Mode           string           `json:"mode,omitempty"`
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
		ProjectID:    spec.ProjectID,
		ActorUserID:  spec.ActorUserID,
		Surface:      spec.Surface,
		Context:      spec.Context,
		Input:        spec.Input,
		DryRun:       true,
		ValidateOnly: spec.ValidateOnly,
		Engine: engines.EngineSpec{
			Type:         spec.Engine.Type,
			Script:       spec.Engine.Script,
			Prompt:       spec.Engine.Prompt,
			ProviderID:   spec.Engine.ProviderID,
			ModuleBase64: spec.Engine.ModuleBase64,
		},
	}
}

func testEngineResource(run automation.Run, spec TestEngineSpec) TestEngineResource {
	sanitized := spec
	if sanitized.Surface == "" {
		sanitized.Surface = "scratch"
	}
	if sanitized.ProjectID == "" {
		sanitized.ProjectID = run.ProjectID
	}
	if sanitized.ActorUserID == "" {
		if actorUserID, ok := run.Input["actor_user_id"].(string); ok {
			sanitized.ActorUserID = actorUserID
		}
	}
	sanitized.Engine.Script = ""
	sanitized.Engine.Prompt = ""
	sanitized.Engine.ModuleBase64 = ""
	if sanitized.Input == nil {
		sanitized.Input = map[string]any{}
	}
	if sanitized.Context == nil {
		sanitized.Context = map[string]any{}
	}
	sanitized.Context["surface"] = sanitized.Surface
	sanitized.Context["project_id"] = sanitized.ProjectID
	sanitized.Context["actor_user_id"] = sanitized.ActorUserID
	sanitized.Context["dry_run"] = true
	sanitized.Context["validate_only"] = sanitized.ValidateOnly
	sanitized.DryRun = true
	return TestEngineResource{
		Metadata: TestEngineMetadata{
			ID:        run.ID,
			CreatedAt: run.CreatedAt,
		},
		Spec: sanitized,
		Status: TestEngineStatus{
			State:          run.Status,
			Mode:           runMode(run),
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

func runMode(run automation.Run) string {
	if input, ok := run.Input["input"].(map[string]any); ok {
		if validateOnly, ok := input["validate_only"].(bool); ok && validateOnly {
			return "validated"
		}
	}
	return "executed"
}

func runOutput(output map[string]any) map[string]any {
	if value, ok := output["output"].(map[string]any); ok {
		return value
	}
	return map[string]any{}
}

func actionPreviews(output map[string]any) []map[string]any {
	rawOutput, ok := output["output"].(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := rawOutput["action_previews"].([]any)
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
			if validateOnly, ok := context["validate_only"].(bool); ok {
				metadata["validate_only"] = validateOnly
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
