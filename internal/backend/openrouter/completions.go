package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type CompletionInput struct {
	ProviderID string
	Prompt     string
}

type CompletionResult struct {
	ProviderID string
	Model      string
	ResponseID string
	Output     map[string]any
	Usage      map[string]any
}

func defaultHTTPClient() httpClient {
	return &http.Client{Timeout: 30 * time.Second}
}

func (s *Service) CompleteJSON(ctx context.Context, input CompletionInput) (CompletionResult, error) {
	if s == nil || s.db == nil {
		return CompletionResult{}, errorsf("OpenRouter service is required")
	}
	provider, err := s.getExecutionProvider(ctx, input.ProviderID)
	if err != nil {
		return CompletionResult{}, err
	}
	if !provider.Enabled {
		return CompletionResult{}, fmt.Errorf("%w: OpenRouter provider is disabled", ErrValidation)
	}
	if strings.TrimSpace(provider.APIKey) == "" {
		return CompletionResult{}, fmt.Errorf("%w: OpenRouter provider API key is not configured", ErrValidation)
	}
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return CompletionResult{}, fmt.Errorf("%w: prompt is required", ErrValidation)
	}

	timeout := time.Duration(provider.DefaultTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	requestBody := chatCompletionRequest{
		Model: provider.DefaultModel,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: "You are Rayboard automation. Return only a valid JSON object. Do not include markdown, prose, or code fences.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: provider.MaxOutputTokens,
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return CompletionResult{}, fmt.Errorf("encode OpenRouter request: %w", err)
	}

	url := strings.TrimRight(s.baseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(callCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return CompletionResult{}, fmt.Errorf("create OpenRouter request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-OpenRouter-Title", "Rayboard")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return CompletionResult{}, fmt.Errorf("call OpenRouter: %w", err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return CompletionResult{}, fmt.Errorf("read OpenRouter response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CompletionResult{}, fmt.Errorf("%w: OpenRouter returned %d: %s", ErrValidation, resp.StatusCode, openrouterErrorMessage(payload))
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(payload, &completion); err != nil {
		return CompletionResult{}, fmt.Errorf("decode OpenRouter response: %w", err)
	}
	if len(completion.Choices) == 0 {
		return CompletionResult{}, fmt.Errorf("%w: OpenRouter response has no choices", ErrValidation)
	}
	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	if content == "" {
		return CompletionResult{}, fmt.Errorf("%w: OpenRouter response content is empty", ErrValidation)
	}
	var output map[string]any
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return CompletionResult{}, fmt.Errorf("%w: OpenRouter response must be a JSON object: %v", ErrValidation, err)
	}
	if output == nil {
		output = map[string]any{}
	}
	return CompletionResult{
		ProviderID: provider.ID,
		Model:      provider.DefaultModel,
		ResponseID: completion.ID,
		Output:     output,
		Usage:      completion.Usage,
	}, nil
}

type chatCompletionRequest struct {
	Model          string            `json:"model"`
	Messages       []chatMessage     `json:"messages"`
	MaxTokens      int               `json:"max_tokens,omitempty"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	ID      string                 `json:"id"`
	Choices []chatCompletionChoice `json:"choices"`
	Usage   map[string]any         `json:"usage,omitempty"`
}

type chatCompletionChoice struct {
	Message chatMessage `json:"message"`
}

func openrouterErrorMessage(payload []byte) string {
	var body struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(payload, &body); err == nil && strings.TrimSpace(body.Error.Message) != "" {
		return body.Error.Message
	}
	return strings.TrimSpace(string(payload))
}

func errorsf(message string) error {
	return fmt.Errorf("%w: %s", ErrValidation, message)
}
