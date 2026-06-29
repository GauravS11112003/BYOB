package ai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockProvider streams a fixed set of chunks.
type mockProvider struct {
	deltas []string
	err    error
}

func (m *mockProvider) Name() string { return "mock" }

func (m *mockProvider) Chat(ctx context.Context, messages []Message, tools []Tool) (<-chan Chunk, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make(chan Chunk)
	go func() {
		defer close(out)
		for _, d := range m.deltas {
			out <- Chunk{Delta: d}
		}
		out <- Chunk{Done: true}
	}()
	return out, nil
}

func TestRuntimeStreamsProviderChunks(t *testing.T) {
	rt := NewRuntime(&mockProvider{deltas: []string{"Hello", " world"}})
	srv := httptest.NewServer(rt.Handler())
	defer srv.Close()

	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(`{"messages":[{"role":"user","content":"hi"}]}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	body := string(raw)

	if !strings.Contains(body, `"delta":"Hello"`) {
		t.Errorf("response missing Hello delta: %q", body)
	}
	if !strings.Contains(body, `"delta":" world"`) {
		t.Errorf("response missing world delta: %q", body)
	}
	if !strings.Contains(body, `"done":true`) {
		t.Errorf("response missing done chunk: %q", body)
	}
}

func TestRuntimeStreamsErrorAsChunk(t *testing.T) {
	rt := NewRuntime(&mockProvider{err: context.DeadlineExceeded})
	srv := httptest.NewServer(rt.Handler())
	defer srv.Close()

	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(`{"messages":[]}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(raw), "AI provider error") {
		t.Errorf("expected structured error chunk, got %q", string(raw))
	}
}

func TestRuntimeRejectsGET(t *testing.T) {
	rt := NewRuntime(&mockProvider{})
	srv := httptest.NewServer(rt.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}
