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

Backend:

```bash
cd byob-backend
go run ./cmd/server
```

Frontend:

```bash
cd byob-frontend
pnpm install
pnpm dev
```

Open http://localhost:3000.

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
