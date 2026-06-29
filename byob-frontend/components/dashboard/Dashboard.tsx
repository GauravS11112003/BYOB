"use client";

import { useMemo } from "react";
import { useCopilotReadable } from "@copilotkit/react-core";
import { ByobChat } from "@/components/byob/ByobChat";
import { brand } from "@/components/byob/theme";
import { useByobActions } from "@/lib/actions";
import { useDataStream } from "@/lib/connectors/useDataStream";
import { ChartWidget } from "@/components/generative/ChartWidget";
import { DataTable } from "@/components/generative/DataTable";
import { WidgetBoundary } from "@/components/generative/WidgetBoundary";

function num(v: unknown): number {
  const n = typeof v === "number" ? v : Number(v);
  return Number.isFinite(n) ? n : 0;
}

export default function Dashboard() {
  const { records, connected, error } = useDataStream();

  useByobActions();

  const recentData = useMemo(() => records.slice(-50), [records]);
  useCopilotReadable({
    description:
      "The most recent records streaming from the connected data sources. Each record has a source, timestamp, and a data object of fields.",
    value: recentData,
  });

  // KPIs derived from the live stream.
  const kpis = useMemo(() => {
    const window = records.slice(-100);
    const totalRequests = window.reduce((s, r) => s + num(r.data.requests), 0);
    const totalErrors = window.reduce((s, r) => s + num(r.data.errors), 0);
    const avgLatency =
      window.length > 0
        ? Math.round(
            window.reduce((s, r) => s + num(r.data.latencyMs), 0) / window.length
          )
        : 0;
    const regions = new Set(window.map((r) => String(r.data.region ?? r.source)));
    const errorRate =
      totalRequests > 0
        ? ((totalErrors / totalRequests) * 100).toFixed(1)
        : "0.0";
    return {
      totalRequests,
      avgLatency,
      errorRate,
      regions: regions.size,
      events: records.length,
    };
  }, [records]);

  // Time-series for the main chart (requests + latency over recent events).
  const timeSeries = useMemo(
    () =>
      records.slice(-24).map((r, i) => ({
        t: new Date(r.timestamp).toLocaleTimeString([], {
          minute: "2-digit",
          second: "2-digit",
        }),
        requests: num(r.data.requests),
        latencyMs: num(r.data.latencyMs),
        i,
      })),
    [records]
  );

  // Aggregate requests by region for the breakdown chart.
  const byRegion = useMemo(() => {
    const agg: Record<string, number> = {};
    for (const r of records.slice(-100)) {
      const region = String(r.data.region ?? r.source);
      agg[region] = (agg[region] ?? 0) + num(r.data.requests);
    }
    return Object.entries(agg)
      .map(([region, requests]) => ({ region, requests }))
      .sort((a, b) => b.requests - a.requests);
  }, [records]);

  const tableColumns = useMemo(() => {
    const last = records[records.length - 1];
    return last ? Object.keys(last.data) : [];
  }, [records]);

  const tableRows = useMemo(() => records.slice(-6).map((r) => r.data), [records]);

  const hasData = records.length > 0;

  return (
    <div className="byob-layout">
      <main className="byob-shell">
        <header className="byob-header">
          <div className="byob-brand">
            <span className="byob-logo">◆</span>
            <div>
              <h1>{brand.name}</h1>
              <p>{brand.tagline}</p>
            </div>
          </div>
          <div className={`byob-status ${connected ? "is-online" : "is-offline"}`}>
            <span className="byob-dot" />
            {connected ? "Live" : "Offline"}
          </div>
        </header>

        {error && <div className="byob-banner">{error}</div>}

        {!hasData ? (
          <div className="byob-empty">
            <h2>Waiting for data…</h2>
            <p>
              Start the Go backend and configure a connector in{" "}
              <code>byob-backend/config.yaml</code>. Records will stream in here
              and you can ask the assistant to visualize them.
            </p>
          </div>
        ) : (
          <>
            <section className="byob-kpis">
              <Kpi label="Total requests" value={kpis.totalRequests.toLocaleString()} accent="primary" />
              <Kpi label="Avg latency" value={`${kpis.avgLatency} ms`} accent="accent" />
              <Kpi label="Error rate" value={`${kpis.errorRate}%`} accent={Number(kpis.errorRate) > 5 ? "danger" : "success"} />
              <Kpi label="Regions" value={String(kpis.regions)} accent="muted" />
              <Kpi label="Events" value={kpis.events.toLocaleString()} accent="muted" />
            </section>

            <section className="byob-grid">
              <div className="byob-col-2">
                <WidgetBoundary>
                  <ChartWidget
                    title="Requests over time"
                    type="line"
                    xKey="t"
                    yKey="requests"
                    data={timeSeries}
                  />
                </WidgetBoundary>
              </div>
              <div className="byob-col-1">
                <WidgetBoundary>
                  <ChartWidget
                    title="Requests by region"
                    type="bar"
                    xKey="region"
                    yKey="requests"
                    data={byRegion}
                  />
                </WidgetBoundary>
              </div>
              <div className="byob-col-3">
                <WidgetBoundary>
                  <DataTable
                    title="Latest records"
                    columns={tableColumns}
                    rows={tableRows}
                  />
                </WidgetBoundary>
              </div>
            </section>
          </>
        )}
      </main>

      <ByobChat />
    </div>
  );
}

function Kpi({
  label,
  value,
  accent,
}: {
  label: string;
  value: string;
  accent: "primary" | "accent" | "success" | "danger" | "muted";
}) {
  return (
    <div className={`byob-kpi accent-${accent}`}>
      <span className="byob-kpi-value">{value}</span>
      <span className="byob-kpi-label">{label}</span>
    </div>
  );
}
