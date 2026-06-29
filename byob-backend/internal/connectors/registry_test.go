package connectors

import (
	"testing"

	"github.com/GauravS11112003/BYOB/byob-backend/internal/config"
)

func TestBuildAllConstructsConnectorsByType(t *testing.T) {
	cfgs := []config.ConnectorConfig{
		{Type: "rest", Name: "metrics", URL: "https://example.com"},
		{Type: "kafka", Name: "orders", Brokers: []string{"localhost:9092"}, Topic: "orders"},
	}

	conns, err := BuildAll(cfgs)
	if err != nil {
		t.Fatalf("BuildAll: %v", err)
	}
	if len(conns) != 2 {
		t.Fatalf("got %d connectors, want 2", len(conns))
	}

	if _, ok := conns[0].(*RESTConnector); !ok {
		t.Errorf("connector 0 = %T, want *RESTConnector", conns[0])
	}
	if _, ok := conns[1].(*KafkaConnector); !ok {
		t.Errorf("connector 1 = %T, want *KafkaConnector", conns[1])
	}
	if conns[0].Name() != "metrics" {
		t.Errorf("name = %q, want metrics", conns[0].Name())
	}
}

func TestBuildRejectsUnknownType(t *testing.T) {
	_, err := Build(config.ConnectorConfig{Type: "mystery", Name: "x"})
	if err == nil {
		t.Fatal("expected error for unknown connector type")
	}
}
