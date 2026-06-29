// Package providers contains concrete AI Provider implementations.
package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/GauravS11112003/BYOB/byob-backend/internal/ai"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

// OpenAI is a Provider backed by the OpenAI Chat Completions API (or any
// OpenAI-compatible endpoint via BaseURL).
type OpenAI struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// OpenAIConfig configures an OpenAI provider.
type OpenAIConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

// NewOpenAI builds an OpenAI provider, applying defaults for model and base URL.
func NewOpenAI(cfg OpenAIConfig) *OpenAI {
	model := cfg.Model
	if model == "" {
		model = "gpt-4o"
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	return &OpenAI{
		apiKey:  cfg.APIKey,
		model:   model,
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (o *OpenAI) Name() string { return "openai" }

// chatCompletionRequest is the OpenAI request body (streaming).
type chatCompletionRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
	Tools    []openAITool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAITool struct {
	Type     string         `json:"type"`
	Function openAIToolFunc `json:"function"`
}

type openAIToolFunc struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// streamChunk is a single SSE delta from OpenAI.
type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// Chat streams a completion from OpenAI, translating its SSE deltas into ai.Chunk.
func (o *OpenAI) Chat(ctx context.Context, messages []ai.Message, tools []ai.Tool) (<-chan ai.Chunk, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("openai: api key not set")
	}

	body := chatCompletionRequest{Model: o.model, Stream: true}
	for _, m := range messages {
		body.Messages = append(body.Messages, openAIMessage{Role: string(m.Role), Content: m.Content})
	}
	for _, t := range tools {
		body.Tools = append(body.Tools, openAITool{
			Type:     "function",
			Function: openAIToolFunc{Name: t.Name, Description: t.Description, Parameters: t.Parameters},
		})
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("openai: status %d", resp.StatusCode)
	}

	out := make(chan ai.Chunk)
	go func() {
		defer close(out)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				break
			}
			var sc streamChunk
			if err := json.Unmarshal([]byte(data), &sc); err != nil {
				continue
			}
			for _, choice := range sc.Choices {
				if choice.Delta.Content != "" {
					select {
					case <-ctx.Done():
						return
					case out <- ai.Chunk{Delta: choice.Delta.Content}:
					}
				}
			}
		}
		select {
		case <-ctx.Done():
		case out <- ai.Chunk{Done: true}:
		}
	}()

	return out, nil
}
