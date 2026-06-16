package openrouterapi

import (
	"time"

	"github.com/timo-42/rayboard/internal/backend/httpapi/shared"
	"github.com/timo-42/rayboard/internal/backend/openrouter"
)

type ProviderIDInput struct {
	shared.AuthInput
	ProviderID string `path:"provider_id" doc:"OpenRouter provider ID."`
}

type CreateProviderInput struct {
	shared.AuthInput
	Body shared.ResourceInput[CreateProviderSpec]
}

type UpdateProviderInput struct {
	shared.AuthInput
	ProviderID string `path:"provider_id" doc:"OpenRouter provider ID."`
	Body       shared.ResourceInput[UpdateProviderSpec]
}

type ProviderMetadata struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProviderSpec struct {
	Name                  string   `json:"name,omitempty"`
	DefaultModel          string   `json:"default_model,omitempty"`
	AllowedModels         []string `json:"allowed_models,omitempty"`
	DefaultTimeoutSeconds int      `json:"default_timeout_seconds,omitempty"`
	MaxOutputTokens       int      `json:"max_output_tokens,omitempty"`
	Enabled               bool     `json:"enabled,omitempty"`
}

type CreateProviderSpec struct {
	Name                  string   `json:"name,omitempty"`
	DefaultModel          string   `json:"default_model,omitempty"`
	APIKey                string   `json:"api_key,omitempty" doc:"OpenRouter API key. Write-only; never returned in responses."`
	AllowedModels         []string `json:"allowed_models,omitempty"`
	DefaultTimeoutSeconds int      `json:"default_timeout_seconds,omitempty"`
	MaxOutputTokens       int      `json:"max_output_tokens,omitempty"`
	Enabled               bool     `json:"enabled,omitempty"`
}

type UpdateProviderSpec struct {
	Name                  *string   `json:"name,omitempty"`
	DefaultModel          *string   `json:"default_model,omitempty"`
	APIKey                *string   `json:"api_key,omitempty" doc:"OpenRouter API key. Omit to leave unchanged; empty string is rejected."`
	AllowedModels         *[]string `json:"allowed_models,omitempty"`
	DefaultTimeoutSeconds *int      `json:"default_timeout_seconds,omitempty"`
	MaxOutputTokens       *int      `json:"max_output_tokens,omitempty"`
	Enabled               *bool     `json:"enabled,omitempty"`
}

type ProviderStatus struct {
	APIKeySet bool `json:"api_key_set"`
	Deleted   bool `json:"deleted"`
}

type ProviderResource = shared.Resource[ProviderMetadata, ProviderSpec, ProviderStatus]

type ListProvidersOutput = shared.ListOutput[ProviderResource]
type CreateProviderOutput = shared.CreatedOutput[ProviderResource]

type ProviderOutput struct {
	Body ProviderResource
}

func (spec CreateProviderSpec) createInput() openrouter.CreateProviderInput {
	return openrouter.CreateProviderInput{
		Name:                  spec.Name,
		DefaultModel:          spec.DefaultModel,
		APIKey:                spec.APIKey,
		AllowedModels:         spec.AllowedModels,
		DefaultTimeoutSeconds: spec.DefaultTimeoutSeconds,
		MaxOutputTokens:       spec.MaxOutputTokens,
		Enabled:               spec.Enabled,
	}
}

func (spec UpdateProviderSpec) updateInput() openrouter.UpdateProviderInput {
	return openrouter.UpdateProviderInput{
		Name:                  spec.Name,
		DefaultModel:          spec.DefaultModel,
		APIKey:                spec.APIKey,
		AllowedModels:         spec.AllowedModels,
		DefaultTimeoutSeconds: spec.DefaultTimeoutSeconds,
		MaxOutputTokens:       spec.MaxOutputTokens,
		Enabled:               spec.Enabled,
	}
}

func providerResource(provider openrouter.Provider) ProviderResource {
	return ProviderResource{
		Metadata: ProviderMetadata{
			ID:        provider.ID,
			CreatedAt: provider.CreatedAt,
			UpdatedAt: provider.UpdatedAt,
		},
		Spec: ProviderSpec{
			Name:                  provider.Name,
			DefaultModel:          provider.DefaultModel,
			AllowedModels:         provider.AllowedModels,
			DefaultTimeoutSeconds: provider.DefaultTimeoutSeconds,
			MaxOutputTokens:       provider.MaxOutputTokens,
			Enabled:               provider.Enabled,
		},
		Status: ProviderStatus{
			APIKeySet: provider.APIKeySet,
			Deleted:   false,
		},
	}
}

func providerResources(providers []openrouter.Provider) []ProviderResource {
	resources := make([]ProviderResource, 0, len(providers))
	for _, provider := range providers {
		resources = append(resources, providerResource(provider))
	}
	return resources
}
