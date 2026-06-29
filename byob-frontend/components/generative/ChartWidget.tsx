"use client";

import {
  Bar,
  BarChart,
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { brand } from "@/components/byob/theme";

export type ChartType = "bar" | "line";

export interface ChartWidgetProps {
  title?: string;
  type: ChartType;
  xKey: string;
  yKey: string;
  data: Array<Record<string, unknown>>;
}

// ChartWidget renders bar/line charts from arbitrary data rows. Used by the AI
// "visualize" action to render directly into chat or the dashboard.
export function ChartWidget({ title, type, xKey, yKey, data }: ChartWidgetProps) {
  return (
    <div className="byob-card">
      {title && <h3 className="byob-card-title">{title}</h3>}
      <ResponsiveContainer width="100%" height={260}>
        {type === "line" ? (
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke={brand.colors.border} />
            <XAxis dataKey={xKey} stroke={brand.colors.textMuted} fontSize={12} />
            <YAxis stroke={brand.colors.textMuted} fontSize={12} />
            <Tooltip
              contentStyle={{
                background: brand.colors.surfaceAlt,
                border: `1px solid ${brand.colors.border}`,
                borderRadius: 8,
                color: brand.colors.text,
              }}
            />
            <Line
              type="monotone"
              dataKey={yKey}
              stroke={brand.colors.primary}
              strokeWidth={2}
              dot={false}
            />
          </LineChart>
        ) : (
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke={brand.colors.border} />
            <XAxis dataKey={xKey} stroke={brand.colors.textMuted} fontSize={12} />
            <YAxis stroke={brand.colors.textMuted} fontSize={12} />
            <Tooltip
              cursor={{ fill: "rgba(255,255,255,0.04)" }}
              contentStyle={{
                background: brand.colors.surfaceAlt,
                border: `1px solid ${brand.colors.border}`,
                borderRadius: 8,
                color: brand.colors.text,
              }}
            />
            <Bar dataKey={yKey} fill={brand.colors.primary} radius={[4, 4, 0, 0]} />
          </BarChart>
        )}
      </ResponsiveContainer>
    </div>
  );
}
