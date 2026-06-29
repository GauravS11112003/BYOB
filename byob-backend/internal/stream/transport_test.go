package stream

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GauravS11112003/BYOB/byob-backend/internal/connectors"
)

func TestSSEHandlerStreamsBroadcastRecords(t *testing.T) {
	hub := NewHub()
	srv := httptest.NewServer(SSEHandler(hub))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}

	reader := bufio.NewReader(resp.Body)
	// Read the initial ": connected" comment line.
	if _, err := reader.ReadString('\n'); err != nil {
		t.Fatalf("read connected line: %v", err)
	}

	// Give the handler a moment to register its subscription, then broadcast.
	waitForSubscriber(t, hub)
	hub.Broadcast(connectors.Record{Source: "test", Data: map[string]any{"v": 99}})

	// Read until we find a data line.
	dataLine := readDataLine(t, reader)
	if !strings.Contains(dataLine, `"v":99`) {
		t.Fatalf("data line = %q, want it to contain v:99", dataLine)
	}
}

func waitForSubscriber(t *testing.T, hub *Hub) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hub.SubscriberCount() > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("no subscriber registered in time")
}

func readDataLine(t *testing.T, reader *bufio.Reader) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if strings.HasPrefix(line, "data: ") {
			return line
		}
	}
	t.Fatal("no data line received in time")
	return ""
}
