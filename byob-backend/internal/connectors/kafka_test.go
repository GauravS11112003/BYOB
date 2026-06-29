package connectors

import (
	"context"
	"errors"
	"testing"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

// fakeReader returns a queued set of messages, then an error to end the loop.
type fakeReader struct {
	msgs []kafka.Message
	i    int
}

func (f *fakeReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	if f.i >= len(f.msgs) {
		return kafka.Message{}, errors.New("EOF")
	}
	m := f.msgs[f.i]
	f.i++
	return m, nil
}

func (f *fakeReader) Close() error { return nil }

func TestKafkaConnectorDecodesJSON(t *testing.T) {
	reader := &fakeReader{msgs: []kafka.Message{
		{Key: []byte("k1"), Value: []byte(`{"region":"eu","count":5}`), Time: time.Unix(10, 0)},
	}}
	c := newKafkaConnectorWithReader("orders", "orders", reader)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan Record, 1)
	go func() { _ = c.Subscribe(ctx, out) }()

	select {
	case r := <-out:
		if r.Source != "orders" {
			t.Errorf("Source = %q, want orders", r.Source)
		}
		if r.Data["region"] != "eu" {
			t.Errorf("region = %v, want eu", r.Data["region"])
		}
		if r.Data["_key"] != "k1" {
			t.Errorf("_key = %v, want k1", r.Data["_key"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for record")
	}
}

func TestKafkaConnectorHandlesNonJSON(t *testing.T) {
	reader := &fakeReader{msgs: []kafka.Message{
		{Value: []byte("plain text"), Time: time.Unix(20, 0)},
	}}
	c := newKafkaConnectorWithReader("logs", "logs", reader)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan Record, 1)
	go func() { _ = c.Subscribe(ctx, out) }()

	select {
	case r := <-out:
		if r.Data["value"] != "plain text" {
			t.Errorf("value = %v, want 'plain text'", r.Data["value"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for record")
	}
}
