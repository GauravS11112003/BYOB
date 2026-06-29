// Package api exposes BYOB's HTTP control-plane endpoints (health, connector
// listing) alongside the streaming and AI handlers wired in main.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/GauravS11112003/BYOB/byob-backend/internal/connectors"
)

// ConnectorInfo is the JSON shape returned by the connectors listing endpoint.
type ConnectorInfo struct {
	Name   string            `json:"name"`
	Schema connectors.Schema `json:"schema"`
}

// Health returns a simple liveness handler.
func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// ListConnectors returns metadata about the active connectors.
func ListConnectors(conns []connectors.Connector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		infos := make([]ConnectorInfo, 0, len(conns))
		for _, c := range conns {
			infos = append(infos, ConnectorInfo{Name: c.Name(), Schema: c.Schema()})
		}
		writeJSON(w, http.StatusOK, infos)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
