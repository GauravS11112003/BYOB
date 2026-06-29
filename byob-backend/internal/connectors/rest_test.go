package connectors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRESTConnectorEmitsObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"region":"us-east","count":12}`))
	}))
	defer srv.Close()

	c := NewRESTConnector(RESTConfig{Name: "metrics", URL: srv.URL, Interval: time.Hour})
	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan Record, 1)
	go func() { _ = c.Subscribe(ctx, out) }()

	select {
	case r := <-out:
		if r.Source != "metrics" {
			t.Errorf("Source = %q, want metrics", r.Source)
		}
		if r.Data["region"] != "us-east" {
			t.Errorf("region = %v, want us-east", r.Data["region"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for record")
	}
}

func TestRESTConnectorEmitsArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"v":1},{"v":2}]`))
	}))
	defer srv.Close()

	c := NewRESTConnector(RESTConfig{Name: "metrics", URL: srv.URL, Interval: time.Hour})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out := make(chan Record, 2)
	go func() { _ = c.Subscribe(ctx, out) }()

	got := map[float64]bool{}
	for i := 0; i < 2; i++ {
		select {
		case r := <-out:
			got[r.Data["v"].(float64)] = true
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for record")
		}
	}
	if !got[1] || !got[2] {
		t.Errorf("expected records v=1 and v=2, got %v", got)
	}
}

func TestRESTConnectorRequiresURL(t *testing.T) {
	c := NewRESTConnector(RESTConfig{Name: "bad"})
	if err := c.Connect(context.Background()); err == nil {
		t.Fatal("expected error for missing url")
	}
}
