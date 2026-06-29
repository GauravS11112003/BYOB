// Package connectors defines the core abstraction for BYOB data sources.
//
// Every data source (Kafka, REST, custom service, database) implements the
// Connector interface and emits normalized Record values. This is the heart of
// "Bring Your Own Backend": adding a new source means implementing one interface.
package connectors

import (
	"context"
	"time"
)

// FieldType is a coarse classification of a field's values, used so the AI and
// generative UI know how to render or aggregate a field.
type FieldType string

const (
	FieldString FieldType = "string"
	FieldNumber FieldType = "number"
	FieldBool   FieldType = "bool"
	FieldTime   FieldType = "time"
	FieldObject FieldType = "object"
)

// Field describes a single attribute of a Record's payload.
type Field struct {
	Name string    `json:"name"`
	Type FieldType `json:"type"`
}

// Schema describes the shape of records a connector emits. It may be partial or
// empty for schema-less sources; consumers should treat it as a hint.
type Schema struct {
	Source string  `json:"source"`
	Fields []Field `json:"fields"`
}

// Record is a single normalized data point flowing through the system.
type Record struct {
	Source    string         `json:"source"`
	Timestamp time.Time      `json:"timestamp"`
	Data      map[string]any `json:"data"`
}

// Connector is the interface every data source must implement.
//
// Lifecycle: Connect once, then Subscribe to stream records until the context is
// cancelled, then Close to release resources.
type Connector interface {
	// Name returns the unique connector name from configuration (e.g. "orders").
	Name() string

	// Connect establishes the underlying connection. It must be safe to call
	// before Subscribe. Implementations should honor context cancellation.
	Connect(ctx context.Context) error

	// Subscribe streams normalized records onto out until ctx is cancelled or an
	// unrecoverable error occurs. Subscribe must not close out; the caller owns it.
	Subscribe(ctx context.Context, out chan<- Record) error

	// Schema returns the (possibly partial) shape of emitted records.
	Schema() Schema

	// Close releases resources. It must be idempotent.
	Close() error
}
