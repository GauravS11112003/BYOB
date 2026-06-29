# BYOB — Bring Your Own Backend

A generative UI dashboard. Plug in your own data sources (Kafka, REST APIs,
custom services) and get an AI-powered dashboard that renders charts, tables,
cards, conversational answers, and action panels on top of your data.

Built on [CopilotKit](https://github.com/CopilotKit/CopilotKit) (generative UI
primitives) with a Go backend as the data-connector hub.

## Structure

```
BYOB/
├── byob-backend/    # Go: connectors, stream hub, AI runtime adapter
├── byob-frontend/   # Next.js + CopilotKit: branded dashboard
└── docs/            # Design spec + implementation plan
```

## Prerequisites

- Go 1.26+
- Node 20+ and pnpm

## Running (dev)

### 1. Try it with the mock data source (no external services needed)

Terminal 1 — mock metrics source:

```bash
cd byob-backend
go run ./examples/mockdata          # serves http://localhost:7070/metrics
```

Terminal 2 — BYOB backend (data plane + AI runtime):

```bash
cd byob-backend
go run ./cmd/server -config config.local.yaml   # listens on :8080
```

Terminal 3 — frontend:

```bash
cd byob-frontend
cp .env.example .env.local          # set OPENAI_API_KEY for the chat
pnpm install
pnpm dev                            # http://localhost:3000
```

Open http://localhost:3000. You'll see live records streaming in. Open the
assistant and try: *"show me a bar chart of requests by region"*.

### 2. The data plane on its own

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/api/connectors
curl http://localhost:8080/stream      # live Server-Sent Events
```

## Architecture notes

- **Data plane (Go):** connectors normalize source data into records and fan
  them out over SSE. This is BYOB's core differentiator.
- **AI plane:** CopilotKit's runtime is hosted in the Next.js route
  `app/api/copilotkit/route.ts` (it speaks CopilotKit's own protocol). The Go
  backend also exposes a standalone provider-agnostic chat endpoint at
  `POST /api/chat` for non-CopilotKit clients.
- **Rebranding:** all CopilotKit imports are isolated to `components/byob/`.
  Edit `components/byob/theme.ts` to change the product name, colors, and copy.

## Testing

Backend:

```bash
cd byob-backend && go test ./...
```

Frontend:

```bash
cd byob-frontend && pnpm lint && pnpm build
```

## Configuration

Declare your data sources and AI provider in `byob-backend/config.yaml`:

```yaml
connectors:
  - type: rest
    name: metrics
    url: https://api.example.com/metrics
    poll_interval: 5s
ai:
  provider: openai
  api_key_env: OPENAI_API_KEY
```

See `docs/superpowers/specs/` for the full design and `docs/superpowers/plans/`
for the implementation plan.
