"use client";

export interface DataTableProps {
  title?: string;
  columns: string[];
  rows: Array<Record<string, unknown>>;
}

// DataTable renders structured tabular data. Columns are explicit so the AI can
// choose which fields to surface.
export function DataTable({ title, columns, rows }: DataTableProps) {
  return (
    <div className="byob-card">
      {title && <h3 className="byob-card-title">{title}</h3>}
      <div className="byob-table-wrap">
        <table className="byob-table">
          <thead>
            <tr>
              {columns.map((c) => (
                <th key={c}>{c}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => (
              <tr key={i}>
                {columns.map((c) => (
                  <td key={c}>{formatCell(row[c])}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function formatCell(value: unknown): string {
  if (value === null || value === undefined) return "—";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}
