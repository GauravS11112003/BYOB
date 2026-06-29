// Command server is the BYOB backend entrypoint. It loads configuration, starts
// the configured data-source connectors, fans their records out over SSE, and
// exposes the AI chat runtime.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/GauravS11112003/BYOB/byob-backend/internal/ai"
	"github.com/GauravS11112003/BYOB/byob-backend/internal/ai/providers"
	"github.com/GauravS11112003/BYOB/byob-backend/internal/api"
	"github.com/GauravS11112003/BYOB/byob-backend/internal/config"
	"github.com/GauravS11112003/BYOB/byob-backend/internal/connectors"
	"github.com/GauravS11112003/BYOB/byob-backend/internal/stream"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Data plane: build connectors, start them, fan out to the hub.
	conns, err := connectors.BuildAll(cfg.Connectors)
	if err != nil {
		log.Fatalf("build connectors: %v", err)
	}
	hub := stream.NewHub()
	startConnectors(ctx, conns, hub)

	// AI plane: build provider + runtime.
	provider, err := providers.Build(cfg.AI)
	if err != nil {
		log.Fatalf("build ai provider: %v", err)
	}
	runtime := ai.NewRuntime(provider)

	mux := http.NewServeMux()
	mux.Handle("/stream", stream.SSEHandler(hub))
	mux.Handle("/api/chat", runtime.Handler())
	mux.Handle("/api/connectors", api.ListConnectors(conns))
	mux.Handle("/healthz", api.Health())

	srv := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: withCORS(mux),
	}

	go func() {
		log.Printf("BYOB backend listening on %s", cfg.Server.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	for _, c := range conns {
		_ = c.Close()
	}
}

// startConnectors connects each connector and streams its records into the hub.
// A failing connector logs and is skipped so others keep running.
func startConnectors(ctx context.Context, conns []connectors.Connector, hub *stream.Hub) {
	for _, c := range conns {
		c := c
		if err := c.Connect(ctx); err != nil {
			log.Printf("connector %q: connect failed: %v", c.Name(), err)
			continue
		}
		go func() {
			records := make(chan connectors.Record, 64)
			go func() {
				if err := c.Subscribe(ctx, records); err != nil && ctx.Err() == nil {
					log.Printf("connector %q: subscribe ended: %v", c.Name(), err)
				}
				close(records)
			}()
			for r := range records {
				hub.Broadcast(r)
			}
		}()
	}
}

// withCORS allows browser clients (the Next.js frontend) to call the backend.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
