# BYOB Generative UI Dashboard Implementation Plan

> **For agentic workers:** Implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a generative UI dashboard where developers connect their own data sources (Kafka, REST, custom) and get AI-rendered charts, tables, cards, chat, and action panels — CopilotKit as a library, Go backend as the data hub.

**Architecture:** Two independently-failing planes. Data plane: Go connectors normalize source data into `Record`s and fan out over WebSocket/SSE. AI plane: Next.js + CopilotKit chat calls the Go AI runtime adapter which proxies to a configured provider. A BYOB wrapper layer isolates all CopilotKit imports for rebranding.

**Tech Stack:** Go 1.26 (backend), Next.js 15 + React 19 + TypeScript (frontend), CopilotKit (npm), pnpm, Recharts (charts).

---

## Phase 0: Repo Scaffold

### Task 0: Monorepo layout

**Files:**
- Create: `byob-backend/go.mod`
- Create: `byob-frontend/` (via create-next-app)
- Create: `README.md`

- [ ] Init Go module: `cd byob-backend && go mod init github.com/GauravS11112003/BYOB/byob-backend`
- [ ] Scaffold Next.js: `pnpm create next-app@latest byob-frontend --ts --eslint --app --tailwind --src-dir=false --import-alias "@/*" --no-turbopack`
- [ ] Add CopilotKit deps: `cd byob-frontend && pnpm add @copilotkit/react-core @copilotkit/react-ui recharts`
- [ ] Commit scaffold.

---

## Phase 1: Go Data Plane

### Task 1: Connector interface + Record type

**Files:**
- Create: `byob-backend/internal/connectors/connector.go`
- Test: `byob-backend/internal/connectors/connector_test.go`

- [ ] Define `Record`, `Schema`, `Field`, and `Connector` interface.
- [ ] Add a `mockConnector` in the test that emits N records, assert `Subscribe` delivers them in order.
- [ ] Run `go test ./...`, expect PASS.
- [ ] Commit.

### Task 2: Stream hub (fan-out)

**Files:**
- Create: `byob-backend/internal/stream/hub.go`
- Test: `byob-backend/internal/stream/hub_test.go`

- [ ] `Hub` with `Subscribe() (id, <-chan Record)`, `Unsubscribe(id)`, `Broadcast(Record)`.
- [ ] Test: two subscribers both receive a broadcast; after unsubscribe one stops receiving.
- [ ] Run tests, PASS. Commit.

### Task 3: REST polling connector

**Files:**
- Create: `byob-backend/internal/connectors/rest.go`
- Test: `byob-backend/internal/connectors/rest_test.go`

- [ ] `RESTConnector` polls a URL on an interval, emits each JSON response as a `Record`.
- [ ] Test with `httptest.Server` returning JSON; assert a `Record` is emitted.
- [ ] Run tests, PASS. Commit.

### Task 4: Kafka connector

**Files:**
- Create: `byob-backend/internal/connectors/kafka.go`
- Test: `byob-backend/internal/connectors/kafka_test.go`

- [ ] Add `github.com/segmentio/kafka-go`.
- [ ] `KafkaConnector` consumes a topic, emits each message value (JSON-decoded) as a `Record`.
- [ ] Test the message→Record transform with a fake reader interface (no live broker).
- [ ] Run tests, PASS. Commit.

### Task 5: Registry + config

**Files:**
- Create: `byob-backend/internal/config/config.go`
- Create: `byob-backend/internal/connectors/registry.go`
- Create: `byob-backend/config.yaml`
- Test: `byob-backend/internal/config/config_test.go`

- [ ] Parse `config.yaml` (connectors + ai) with `gopkg.in/yaml.v3`.
- [ ] Registry builds connectors from config by `type`.
- [ ] Test: load sample yaml → registry returns expected connector types.
- [ ] Run tests, PASS. Commit.

---

## Phase 2: Transport + AI Plane

### Task 6: WS/SSE transport

**Files:**
- Create: `byob-backend/internal/stream/transport.go`
- Test: `byob-backend/internal/stream/transport_test.go`

- [ ] SSE handler `/stream` subscribes to hub, writes `data:` events.
- [ ] Test with `httptest` that a broadcast record appears on the SSE response.
- [ ] Run tests, PASS. Commit.

### Task 7: AI runtime adapter + providers

**Files:**
- Create: `byob-backend/internal/ai/runtime.go`
- Create: `byob-backend/internal/ai/providers/openai.go`
- Test: `byob-backend/internal/ai/runtime_test.go`

- [ ] `Provider` interface: `Chat(ctx, messages, tools) (stream)`.
- [ ] OpenAI provider implementation (configurable base URL + key).
- [ ] Runtime handler accepts chat requests, calls provider, streams back.
- [ ] Test runtime with a mock provider; assert streamed response shape.
- [ ] Run tests, PASS. Commit.

### Task 8: Server wiring

**Files:**
- Create: `byob-backend/cmd/server/main.go`
- Create: `byob-backend/internal/api/handlers.go`

- [ ] `main.go`: load config, build registry, start connectors → hub, mount `/stream`, `/api/copilotkit`, `/api/connectors`, `/healthz`.
- [ ] Manual run: `go run ./cmd/server` boots and `/healthz` returns 200.
- [ ] Commit.

---

## Phase 3: Frontend BYOB Wrapper + Generative UI

### Task 9: BYOB wrapper layer + theme

**Files:**
- Create: `byob-frontend/components/byob/theme.ts`
- Create: `byob-frontend/components/byob/ByobProvider.tsx`
- Create: `byob-frontend/components/byob/ByobChat.tsx`

- [ ] `theme.ts` brand tokens. `ByobProvider` wraps `<CopilotKit runtimeUrl>`. `ByobChat` wraps `<CopilotChat>` with branding. CopilotKit imported only here.
- [ ] Commit.

### Task 10: Data stream hook

**Files:**
- Create: `byob-frontend/lib/connectors/useDataStream.ts`

- [ ] `useDataStream(source)` opens EventSource to Go `/stream`, buffers records, exposes latest data.
- [ ] Commit.

### Task 11: Generative components + actions

**Files:**
- Create: `byob-frontend/components/generative/{ChartWidget,DataTable,SummaryCard,ActionPanel}.tsx`
- Create: `byob-frontend/lib/actions/{visualize,query,act}.ts`

- [ ] Components render from props. Actions are `useCopilotAction` hooks that render the components.
- [ ] Commit.

### Task 12: Dashboard page + rebrand

**Files:**
- Modify: `byob-frontend/app/layout.tsx`, `byob-frontend/app/page.tsx`

- [ ] Wrap app in `ByobProvider`, mount `ByobChat`, register actions, feed `useDataStream` to copilot-readable state. Set BYOB metadata/favicon.
- [ ] Manual run: `pnpm dev`, dashboard renders, chat opens.
- [ ] Commit.

---

## Phase 4: Integration

### Task 13: End-to-end smoke

- [ ] Run Go backend + Next.js frontend with a REST connector pointed at a sample endpoint.
- [ ] Ask the chat to visualize data → ChartWidget renders.
- [ ] Document run steps in `README.md`. Commit.

---

## Notes / Risks

- CopilotKit runtime is normally Node/TS. We implement its wire protocol in Go (`ai/runtime.go`). If the protocol proves unstable, fall back to a Next.js `app/api/copilotkit/route.ts` for the AI plane only; Go keeps the data plane.
- Kafka tests avoid a live broker by testing the transform layer.
