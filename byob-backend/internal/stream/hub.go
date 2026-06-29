// Package stream fans out connector records to many subscribers (frontend
// clients) over the data plane.
package stream

import (
	"sync"

	"github.com/GauravS11112003/BYOB/byob-backend/internal/connectors"
)

// subscriberBuffer is the per-subscriber channel buffer. Slow subscribers that
// fill their buffer drop records rather than blocking the whole hub.
const subscriberBuffer = 256

// Hub broadcasts records to all active subscribers. It is safe for concurrent use.
type Hub struct {
	mu     sync.RWMutex
	nextID uint64
	subs   map[uint64]chan connectors.Record
}

// NewHub creates an empty Hub.
func NewHub() *Hub {
	return &Hub{subs: make(map[uint64]chan connectors.Record)}
}

// Subscribe registers a new subscriber and returns its id and receive channel.
// Call Unsubscribe with the id when done to release resources.
func (h *Hub) Subscribe() (uint64, <-chan connectors.Record) {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := h.nextID
	h.nextID++
	ch := make(chan connectors.Record, subscriberBuffer)
	h.subs[id] = ch
	return id, ch
}

// Unsubscribe removes a subscriber and closes its channel. It is idempotent.
func (h *Hub) Unsubscribe(id uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if ch, ok := h.subs[id]; ok {
		delete(h.subs, id)
		close(ch)
	}
}

// Broadcast sends a record to every subscriber. Subscribers whose buffer is full
// are skipped (the record is dropped for them) so one slow client cannot stall
// delivery to others.
func (h *Hub) Broadcast(r connectors.Record) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, ch := range h.subs {
		select {
		case ch <- r:
		default:
			// subscriber is slow; drop this record for them.
		}
	}
}

// SubscriberCount returns the number of active subscribers.
func (h *Hub) SubscriberCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs)
}
