package config

import (
	"testing"
	"time"
)

func TestParseValidConfig(t *testing.T) {
	raw := []byte(`
server:
  addr: ":9090"
connectors:
  - type: rest
    name: metrics
    url: https://api.example.com/metrics
    poll_interval: 5s
  - type: kafka
    name: orders
    brokers: ["localhost:9092"]
    topic: orders
ai:
  provider: openai
  model: gpt-4o
  api_key_env: OPENAI_API_KEY
`)
	cfg, err := parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if cfg.Server.Addr != ":9090" {
		t.Errorf("Addr = %q, want :9090", cfg.Server.Addr)
	}
	if len(cfg.Connectors) != 2 {
		t.Fatalf("got %d connectors, want 2", len(cfg.Connectors))
	}
	if cfg.Connectors[0].PollInterval.AsDuration() != 5*time.Second {
		t.Errorf("poll_interval = %v, want 5s", cfg.Connectors[0].PollInterval.AsDuration())
	}
	if cfg.Connectors[1].Topic != "orders" {
		t.Errorf("topic = %q, want orders", cfg.Connectors[1].Topic)
	}
	if cfg.AI.Provider != "openai" {
		t.Errorf("provider = %q, want openai", cfg.AI.Provider)
	}
}

func TestParseAppliesDefaultAddr(t *testing.T) {
	cfg, err := parse([]byte(`connectors: []`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if cfg.Server.Addr != ":8080" {
		t.Errorf("default Addr = %q, want :8080", cfg.Server.Addr)
	}
}

func TestParseRejectsConnectorWithoutName(t *testing.T) {
	_, err := parse([]byte(`
connectors:
  - type: rest
    url: https://example.com
`))
	if err == nil {
		t.Fatal("expected error for connector without name")
	}
}
