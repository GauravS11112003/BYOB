"use client";

export interface SummaryMetric {
  label: string;
  value: string;
}

export interface SummaryCardProps {
  title: string;
  summary: string;
  metrics?: SummaryMetric[];
}

// SummaryCard is an AI-generated interpretation of data: a headline, a prose
// summary, and optional key metrics.
export function SummaryCard({ title, summary, metrics }: SummaryCardProps) {
  return (
    <div className="byob-card">
      <h3 className="byob-card-title">{title}</h3>
      <p className="byob-summary">{summary}</p>
      {metrics && metrics.length > 0 && (
        <div className="byob-metrics">
          {metrics.map((m) => (
            <div key={m.label} className="byob-metric">
              <span className="byob-metric-value">{m.value}</span>
              <span className="byob-metric-label">{m.label}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
