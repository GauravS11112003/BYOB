// Command mockdata serves random metric JSON for trying out BYOB locally.
// Point a REST connector at http://localhost:7070/metrics.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	addr := flag.String("addr", ":7070", "listen address")
	flag.Parse()

	regions := []string{"us-east", "us-west", "eu-central", "ap-south"}

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		region := regions[rand.Intn(len(regions))]
		payload := map[string]any{
			"region":    region,
			"requests":  rand.Intn(500),
			"latencyMs": 20 + rand.Intn(180),
			"errors":    rand.Intn(10),
			"ts":        time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})

	log.Printf("mockdata serving metrics on %s/metrics", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}
