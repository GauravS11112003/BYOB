# BYOB — Bring Your Own Backend: Generative UI Dashboard

**Date:** 2026-06-28
**Status:** Approved design

## Summary

BYOB is a developer-facing generative UI dashboard. Developers plug in their own
data sources (Kafka streams, REST APIs, custom services, databases) and get an
AI-powered dashboard that renders charts, tables, cards, conversational answers,
and action panels on top of that data. CopilotKit provides the generative UI
primitives (used as a library); a Go backend is the data-connector hub.

## Goals

- Provider-agnostic **data** layer: connect any backend behind one interface.
- Generative UI: chat panel, AI-triggered actions, and custom AI-rendered components.
- Rebrandable: nothing user-facing references CopilotKit.
- AI provider is pluggable via config (OpenAI / Anthropic / Gemini / custom endpoint).

## Non-Goals (YAGNI for v1)

- Multi-tenant SaaS / billing.
- White-label resale tooling.
- Building generative UI primitives from scratch (we wrap CopilotKit).

## Chosen Approach

**Approach A — Library Wrapper.** Use CopilotKit as an npm dependency and build
BYOB (Next.js frontend + Go backend) on top. Rationale: ship fast on battle-tested
primitives; invest engineering effort in the differentiating data-connector layer;
keep the option to fork later open.

## Architecture

```
Data Sources            BYOB Go Backend            BYOB Next.js Frontend
─────────────           ────────────────           ─────────────────────
Kafka Stream    ───▶     Connector Registry  ───▶   BYOB Component Library
REST API                 WebSocket / SSE Hub        CopilotKit (npm dep)
Custom Service           AI Runtime Adapter         Chat Panel
Database                 Auth / Config API          Generative Dashboard
                              ▲
                         AI Provider (pluggable: OpenAI/Anthropic/Gemini/custom)
```

Two planes that fail independently:
- **Data plane:** connectors → normalized `Record` → stream hub → frontend.
- **AI plane:** frontend chat → Go AI runtime adapter → configured provider → stream back.

## Frontend Structure (Next.js)

```
byob-frontend/
├── app/
│   ├── layout.tsx              # wraps app in <ByobProvider>
│   ├── page.tsx                # Dashboard home
│   ├── api/copilotkit/route.ts # proxy → Go AI runtime (if needed)
│   └── (dashboard)/[view]/page.tsx
├── components/
│   ├── byob/                   # branded wrapper layer (only place importing CopilotKit)
│   │   ├── ByobProvider.tsx
│   │   ├── ByobChat.tsx
│   │   ├── ByobSidebar.tsx
│   │   └── theme.ts            # brand tokens
│   └── generative/             # AI-rendered components
│       ├── ChartWidget.tsx
│       ├── DataTable.tsx
│       ├── SummaryCard.tsx
│       └── ActionPanel.tsx
└── lib/
    ├── actions/                # useCopilotAction definitions (visualize, query, act)
    └── connectors/useDataStream.ts  # WS/SSE subscription hook
```

**Wrapper-layer rule:** CopilotKit is imported only inside `components/byob/`.
The rest of the app imports `ByobChat`, `ByobProvider`, etc. This isolates
rebranding and keeps CopilotKit swappable.

## Go Backend Structure

```
byob-backend/
├── cmd/server/main.go
├── internal/
│   ├── connectors/   # connector.go (interface), kafka.go, rest.go, http_custom.go, registry.go
│   ├── stream/       # hub.go (fan-out), transport.go (WS + SSE)
│   ├── ai/           # runtime.go (CopilotKit protocol), providers/{openai,anthropic,gemini,custom}.go
│   ├── config/       # config.go
│   └── api/          # handlers.go (list connectors, health, config)
├── go.mod
└── config.yaml
```

### Core abstraction

```go
type Connector interface {
    Connect(ctx context.Context) error
    Subscribe(ctx context.Context, out chan<- Record) error
    Schema() Schema
    Close() error
}

type Record struct {
    Source    string
    Timestamp time.Time
    Data      map[string]any
}
```

New data source = implement `Connector` + register it.

### Config-driven setup

```yaml
connectors:
  - type: kafka
    name: orders
    brokers: ["localhost:9092"]
    topic: orders
  - type: rest
    name: metrics
    url: https://api.example.com/metrics
    poll_interval: 5s
ai:
  provider: openai
  api_key_env: OPENAI_API_KEY
```

### Risk / fallback

CopilotKit's runtime is normally Node/TS. Implementing its wire protocol in Go
(`ai/runtime.go`) is the main added effort. Fallback: a thin Next.js API route
owns the AI plane while Go keeps owning the data plane.

## Data Flow (example: "show me order volume by region")

1. Connectors stream data → Go normalizes to `Record` → Hub fans out over WS/SSE.
2. Frontend `useDataStream()` feeds copilot-readable state.
3. User asks in `ByobChat`.
4. Request → Go AI runtime → provider (with data context + action schemas).
5. AI calls `useCopilotAction("visualize", {...})`.
6. Frontend renders `<ChartWidget>` inline.
7. Widget updates live as new records stream in.

## Error Handling

| Layer | Failure | Handling |
|-------|---------|----------|
| Connector | source down | retry w/ backoff; mark `degraded`; expose via `/api/connectors` |
| Stream Hub | client disconnect | cleanup subscription; frontend auto-reconnects |
| AI runtime | provider error / bad key | structured error to chat; never crash stream |
| AI action | invalid args | validate against schema; reject gracefully |
| Frontend | widget render fails | per-widget error boundary |

Principle: data plane and AI plane fail independently.

## Testing

- **Go:** connector unit tests w/ mocks; `Connector` conformance suite; hub fan-out;
  AI runtime w/ mock provider.
- **Frontend:** generative widget component tests; action handler tests;
  `useDataStream` hook test w/ mock WS.
- **Integration:** one e2e — mock connector → Go → frontend → mock AI → rendered widget.

## Rebranding Checklist

- Brand tokens in `components/byob/theme.ts`.
- All CopilotKit imports isolated to `components/byob/`.
- Custom chat copy, empty/loading states.
- Own favicon, app name, metadata in `app/layout.tsx`.
