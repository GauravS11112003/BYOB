"use client";

import { useCopilotAction } from "@copilotkit/react-core";
import React from "react";
import { DataTable } from "@/components/generative/DataTable";
import { SummaryCard, type SummaryMetric } from "@/components/generative/SummaryCard";

// useTableAction renders structured data as a table.
export function useTableAction() {
  useCopilotAction({
    name: "showTable",
    description:
      "Render data as a table. Use when the user wants to see structured rows.",
    parameters: [
      { name: "title", type: "string", description: "Table title", required: false },
      {
        name: "columns",
        type: "string[]",
        description: "Column field names to display",
        required: true,
      },
      {
        name: "rows",
        type: "object[]",
        description: "Array of row objects",
        required: true,
      },
    ],
    render: ({ args }) => {
      const { title, columns, rows } = args as {
        title?: string;
        columns?: string[];
        rows?: Array<Record<string, unknown>>;
      };
      if (!columns || !rows) {
        return React.createElement("div", { className: "byob-card" }, "Preparing table…");
      }
      return React.createElement(DataTable, { title, columns, rows });
    },
  });
}

// useSummarizeAction renders an AI interpretation of data as a card.
export function useSummarizeAction() {
  useCopilotAction({
    name: "summarize",
    description:
      "Render a summary card interpreting the data, with a headline, prose, and key metrics.",
    parameters: [
      { name: "title", type: "string", description: "Card title", required: true },
      { name: "summary", type: "string", description: "Prose summary", required: true },
      {
        name: "metrics",
        type: "object[]",
        description: "Optional key metrics: each item has 'label' and 'value' strings",
        required: false,
      },
    ],
    render: ({ args }) => {
      const { title, summary, metrics } = args as {
        title?: string;
        summary?: string;
        metrics?: SummaryMetric[];
      };
      if (!title || !summary) {
        return React.createElement("div", { className: "byob-card" }, "Summarizing…");
      }
      return React.createElement(SummaryCard, { title, summary, metrics });
    },
  });
}
