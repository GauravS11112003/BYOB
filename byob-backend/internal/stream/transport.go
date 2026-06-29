package stream

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSEHandler returns an http.Handler that streams hub records to the client as
// Server-Sent Events. Each record is written as a single `data:` line of JSON.
//
// The connection lives until the client disconnects (request context cancelled).
func SSEHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		id, ch := hub.Subscribe()
		defer hub.Unsubscribe(id)

		// Initial comment so clients know the stream is open.
		fmt.Fprintf(w, ": connected\n\n")
		flusher.Flush()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case rec, ok := <-ch:
				if !ok {
					return
				}
				payload, err := json.Marshal(rec)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", payload)
				flusher.Flush()
			}
		}
	}
}
