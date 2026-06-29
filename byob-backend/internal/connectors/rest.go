package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RESTConnector polls an HTTP endpoint on a fixed interval and emits each JSON
// response as a Record. It handles both JSON objects (one record) and JSON
// arrays (one record per element).
type RESTConnector struct {
	name     string
	url      string
	interval time.Duration
	client   *http.Client
}

// RESTConfig configures a RESTConnector.
type RESTConfig struct {
	Name     string
	URL      string
	Interval time.Duration
}

// NewRESTConnector builds a RESTConnector. Interval defaults to 5s if zero.
func NewRESTConnector(cfg RESTConfig) *RESTConnector {
	interval := cfg.Interval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &RESTConnector{
		name:     cfg.Name,
		url:      cfg.URL,
		interval: interval,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *RESTConnector) Name() string { return c.name }

func (c *RESTConnector) Connect(ctx context.Context) error {
	if c.url == "" {
		return fmt.Errorf("rest connector %q: url is required", c.name)
	}
	return nil
}

func (c *RESTConnector) Schema() Schema {
	return Schema{Source: c.name}
}

func (c *RESTConnector) Close() error { return nil }

// Subscribe polls immediately, then every interval, until ctx is cancelled.
// Transient poll errors are skipped (the next tick retries) rather than fatal.
func (c *RESTConnector) Subscribe(ctx context.Context, out chan<- Record) error {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// poll once immediately so consumers see data without waiting a full interval.
	if err := c.pollOnce(ctx, out); err != nil && ctx.Err() != nil {
		return ctx.Err()
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := c.pollOnce(ctx, out); err != nil && ctx.Err() != nil {
				return ctx.Err()
			}
		}
	}
}

// pollOnce performs a single fetch and emits the resulting record(s).
func (c *RESTConnector) pollOnce(ctx context.Context, out chan<- Record) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("rest connector %q: status %d", c.name, resp.StatusCode)
	}

	return c.emit(ctx, body, out)
}

// emit decodes a JSON body and sends one Record per object.
func (c *RESTConnector) emit(ctx context.Context, body []byte, out chan<- Record) error {
	now := time.Now().UTC()

	// Try array first.
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err == nil {
		for _, item := range arr {
			if err := c.send(ctx, out, Record{Source: c.name, Timestamp: now, Data: item}); err != nil {
				return err
			}
		}
		return nil
	}

	// Fall back to a single object.
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		return fmt.Errorf("rest connector %q: decode body: %w", c.name, err)
	}
	return c.send(ctx, out, Record{Source: c.name, Timestamp: now, Data: obj})
}

func (c *RESTConnector) send(ctx context.Context, out chan<- Record, r Record) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case out <- r:
		return nil
	}
}
