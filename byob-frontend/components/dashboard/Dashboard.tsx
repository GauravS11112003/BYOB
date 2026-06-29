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

export default function Dashboard() {
  const { records, connected, error } = useDataStream();

  // Register all generative UI actions so the AI can render charts/tables/etc.
  useByobActions();

  // Expose the latest data to the AI as readable context.
  const recentData = useMemo(() => records.slice(-50), [records]);
  useCopilotReadable({
    description:
      "The most recent records streaming from the connected data sources. Each record has a source, timestamp, and a data object of fields.",
    value: recentData,
  });

  // Derive columns for the live table from the most recent record.
  const columns = useMemo(() => {
    const last = records[records.length - 1];
    return last ? Object.keys(last.data) : [];
  }, [records]);

  const tableRows = useMemo(
    () => records.slice(-8).map((r) => r.data),
    [records]
  );

  // A simple live chart: count of records per source over the recent window.
  const chartData = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const r of records.slice(-100)) {
      counts[r.source] = (counts[r.source] ?? 0) + 1;
    }
    return Object.entries(counts).map(([source, count]) => ({ source, count }));
  }, [records]);

  return (
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
          {connected ? "Backend connected" : "Backend offline"}
        </div>
      </header>

      {error && <div className="byob-banner">{error}</div>}

      <section className="byob-grid">
        {records.length === 0 ? (
          <div className="byob-empty">
            <h2>No data yet</h2>
            <p>
              Start the Go backend and configure a connector in{" "}
              <code>byob-backend/config.yaml</code>. Records will stream in here,
              and you can ask the assistant to visualize them.
            </p>
          </div>
        ) : (
          <>
            <WidgetBoundary>
              <ChartWidget
                title="Records by source (live)"
                type="bar"
                xKey="source"
                yKey="count"
                data={chartData}
              />
            </WidgetBoundary>
            <WidgetBoundary>
              <DataTable
                title="Latest records"
                columns={columns}
                rows={tableRows}
              />
            </WidgetBoundary>
          </>
        )}
      </section>

      <ByobChat />
    </main>
  );
}
