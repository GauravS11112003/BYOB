package stream

import (
	"testing"
	"time"

	"github.com/GauravS11112003/BYOB/byob-backend/internal/connectors"
)

func recv(t *testing.T, ch <-chan connectors.Record) (connectors.Record, bool) {
	t.Helper()
	select {
	case r, ok := <-ch:
		return r, ok
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for record")
		return connectors.Record{}, false
	}
}

func TestHubBroadcastsToAllSubscribers(t *testing.T) {
	h := NewHub()
	_, a := h.Subscribe()
	_, b := h.Subscribe()

	if got := h.SubscriberCount(); got != 2 {
		t.Fatalf("SubscriberCount = %d, want 2", got)
	}

	want := connectors.Record{Source: "s", Data: map[string]any{"v": 42}}
	h.Broadcast(want)

	ra, _ := recv(t, a)
	rb, _ := recv(t, b)
	if ra.Data["v"] != 42 || rb.Data["v"] != 42 {
		t.Fatalf("subscribers got %v / %v, want both 42", ra.Data["v"], rb.Data["v"])
	}
}

func TestHubUnsubscribeStopsDelivery(t *testing.T) {
	h := NewHub()
	idA, a := h.Subscribe()
	_, b := h.Subscribe()

	h.Unsubscribe(idA)
	if got := h.SubscriberCount(); got != 1 {
		t.Fatalf("SubscriberCount = %d, want 1", got)
	}

	// channel a should be closed by Unsubscribe.
	if _, ok := <-a; ok {
		t.Fatal("expected channel a to be closed after Unsubscribe")
	}

	h.Broadcast(connectors.Record{Source: "s", Data: map[string]any{"v": 7}})
	rb, _ := recv(t, b)
	if rb.Data["v"] != 7 {
		t.Fatalf("remaining subscriber got %v, want 7", rb.Data["v"])
	}
}

func TestHubUnsubscribeIdempotent(t *testing.T) {
	h := NewHub()
	id, _ := h.Subscribe()
	h.Unsubscribe(id)
	h.Unsubscribe(id) // must not panic
}
