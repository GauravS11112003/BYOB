"use client";

import { useCopilotAction } from "@copilotkit/react-core";
import React from "react";
import { ChartWidget, type ChartType } from "@/components/generative/ChartWidget";

// useVisualizeAction lets the AI render a chart from data rows it provides.
export function useVisualizeAction() {
  useCopilotAction({
    name: "visualize",
    description:
      "Render a bar or line chart from data rows. Use when the user wants to see data visually.",
    parameters: [
      { name: "title", type: "string", description: "Chart title", required: false },
      {
        name: "chartType",
        type: "string",
        description: "Either 'bar' or 'line'",
        required: true,
      },
      { name: "xKey", type: "string", description: "Field name for the X axis", required: true },
      { name: "yKey", type: "string", description: "Field name for the Y axis", required: true },
      {
        name: "data",
        type: "object[]",
        description: "Array of row objects to plot",
        required: true,
      },
    ],
    handler: async () => "Chart rendered.",
    render: ({ args }) => {
      const { title, chartType, xKey, yKey, data } = args as {
        title?: string;
        chartType?: string;
        xKey?: string;
        yKey?: string;
        data?: Array<Record<string, unknown>>;
      };
      if (!xKey || !yKey || !data) {
        return React.createElement("div", { className: "byob-card" }, "Preparing chart…");
      }
      return React.createElement(ChartWidget, {
        title,
        type: (chartType === "line" ? "line" : "bar") as ChartType,
        xKey,
        yKey,
        data,
      });
    },
  });
}
