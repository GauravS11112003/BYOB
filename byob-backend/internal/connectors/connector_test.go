package connectors

import (
	"context"
	"testing"
	"time"
)

// mockConnector emits a fixed set of records, one per Subscribe, for testing the
// interface contract and any code that consumes a Connector.
type mockConnector struct {
	name    string
	records []Record
	closed  bool
}

func (m *mockConnector) Name() string { return m.name }

func (m *mockConnector) Connect(ctx context.Context) error { return nil }

func (m *mockConnector) Subscribe(ctx context.Context, out chan<- Record) error {
	for _, r := range m.records {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- r:
		}
	}
	return nil
}

func (m *mockConnector) Schema() Schema {
	return Schema{Source: m.name, Fields: []Field{{Name: "value", Type: FieldNumber}}}
}

func (m *mockConnector) Close() error {
	m.closed = true
	return nil
}

// ensure mockConnector satisfies the interface at compile time.
var _ Connector = (*mockConnector)(nil)

func TestConnectorDeliversRecordsInOrder(t *testing.T) {
	want := []Record{
		{Source: "test", Timestamp: time.Unix(1, 0), Data: map[string]any{"value": 1}},
		{Source: "test", Timestamp: time.Unix(2, 0), Data: map[string]any{"value": 2}},
		{Source: "test", Timestamp: time.Unix(3, 0), Data: map[string]any{"value": 3}},
	}
	c := &mockConnector{name: "test", records: want}

	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	out := make(chan Record, len(want))
	if err := c.Subscribe(context.Background(), out); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	close(out)

	var got []Record
	for r := range out {
		got = append(got, r)
	}

	if len(got) != len(want) {
		t.Fatalf("got %d records, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Data["value"] != want[i].Data["value"] {
			t.Errorf("record %d: got %v, want %v", i, got[i].Data["value"], want[i].Data["value"])
		}
	}
}

func TestConnectorSubscribeRespectsContext(t *testing.T) {
	c := &mockConnector{
		name:    "test",
		records: []Record{{Source: "test", Data: map[string]any{"value": 1}}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before subscribing

	out := make(chan Record) // unbuffered: send would block without cancellation
	err := c.Subscribe(ctx, out)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}
