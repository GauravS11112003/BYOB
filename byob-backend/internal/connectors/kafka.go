package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

// messageReader is the subset of kafka.Reader behavior we depend on. Abstracting
// it lets us unit-test the message→Record transform without a live broker.
type messageReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

// KafkaConnector consumes a Kafka topic and emits each message value as a Record.
// Message values are JSON-decoded when possible; otherwise the raw string is kept
// under the "value" key.
type KafkaConnector struct {
	name   string
	topic  string
	reader messageReader
}

// KafkaConfig configures a KafkaConnector.
type KafkaConfig struct {
	Name    string
	Brokers []string
	Topic   string
	GroupID string
}

// NewKafkaConnector builds a KafkaConnector backed by a real kafka-go reader.
func NewKafkaConnector(cfg KafkaConfig) *KafkaConnector {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	})
	return &KafkaConnector{name: cfg.Name, topic: cfg.Topic, reader: reader}
}

// newKafkaConnectorWithReader is used by tests to inject a fake reader.
func newKafkaConnectorWithReader(name, topic string, reader messageReader) *KafkaConnector {
	return &KafkaConnector{name: name, topic: topic, reader: reader}
}

func (c *KafkaConnector) Name() string { return c.name }

func (c *KafkaConnector) Connect(ctx context.Context) error {
	if c.reader == nil {
		return fmt.Errorf("kafka connector %q: reader not initialized", c.name)
	}
	return nil
}

func (c *KafkaConnector) Schema() Schema {
	return Schema{Source: c.name}
}

func (c *KafkaConnector) Close() error {
	if c.reader == nil {
		return nil
	}
	return c.reader.Close()
}

// Subscribe reads messages until ctx is cancelled. A read error after
// cancellation is treated as a clean shutdown.
func (c *KafkaConnector) Subscribe(ctx context.Context, out chan<- Record) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("kafka connector %q: read: %w", c.name, err)
		}
		rec := c.toRecord(msg)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- rec:
		}
	}
}

// toRecord converts a Kafka message into a normalized Record.
func (c *KafkaConnector) toRecord(msg kafka.Message) Record {
	ts := msg.Time
	if ts.IsZero() {
		ts = time.Now()
	}

	data := map[string]any{}
	if err := json.Unmarshal(msg.Value, &data); err != nil {
		// not JSON: keep the raw value.
		data = map[string]any{"value": string(msg.Value)}
	}
	if len(msg.Key) > 0 {
		data["_key"] = string(msg.Key)
	}

	return Record{Source: c.name, Timestamp: ts.UTC(), Data: data}
}
