// Package ai is the AI plane: a provider-agnostic chat runtime that the BYOB
// frontend calls. Providers (OpenAI, Anthropic, etc.) implement the Provider
// interface; the runtime exposes an HTTP endpoint that streams responses.
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Role identifies who produced a chat message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is a single chat message.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Tool describes an action the model may call. Parameters is a JSON-schema object.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// Chunk is a streamed fragment of a provider response.
type Chunk struct {
	// Delta is incremental assistant text.
	Delta string `json:"delta,omitempty"`
	// Done marks the final chunk.
	Done bool `json:"done,omitempty"`
}

// Provider is implemented by each AI backend.
type Provider interface {
	// Name returns the provider identifier (e.g. "openai").
	Name() string
	// Chat streams response chunks for the given conversation. Implementations
	// must close the returned channel when finished and honor ctx cancellation.
	Chat(ctx context.Context, messages []Message, tools []Tool) (<-chan Chunk, error)
}

// Runtime wires a Provider to an HTTP handler.
type Runtime struct {
	provider Provider
}

// NewRuntime creates a Runtime backed by the given provider.
func NewRuntime(p Provider) *Runtime {
	return &Runtime{provider: p}
}

// chatRequest is the JSON body accepted by the chat endpoint.
type chatRequest struct {
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools"`
}

// Handler returns an http.Handler that accepts a chat request and streams the
// provider response back as Server-Sent Events.
func (rt *Runtime) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		chunks, err := rt.provider.Chat(r.Context(), req.Messages, req.Tools)
		if err != nil {
			// Stream a structured error so the chat UI can show it without the
			// data plane being affected.
			writeChunk(w, flusher, Chunk{Delta: fmt.Sprintf("AI provider error: %v", err), Done: true})
			return
		}

		for chunk := range chunks {
			writeChunk(w, flusher, chunk)
		}
	}
}

func writeChunk(w http.ResponseWriter, flusher http.Flusher, c Chunk) {
	payload, err := json.Marshal(c)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", payload)
	flusher.Flush()
}
