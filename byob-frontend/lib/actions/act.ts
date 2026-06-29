"use client";

import { useCopilotAction } from "@copilotkit/react-core";
import React from "react";
import { ActionPanel, type SuggestedAction } from "@/components/generative/ActionPanel";

// useSuggestActionsAction renders a panel of AI-suggested next actions.
export function useSuggestActionsAction() {
  useCopilotAction({
    name: "suggestActions",
    description:
      "Render a panel of suggested next actions for the user based on the data.",
    parameters: [
      { name: "title", type: "string", description: "Panel title", required: true },
      {
        name: "actions",
        type: "object[]",
        description:
          "Array of actions; each item has 'label' (string) and optional 'description' (string)",
        required: true,
      },
    ],
    render: ({ args }) => {
      const { title, actions } = args as {
        title?: string;
        actions?: SuggestedAction[];
      };
      if (!title || !actions) {
        return React.createElement("div", { className: "byob-card" }, "Preparing actions…");
      }
      return React.createElement(ActionPanel, { title, actions });
    },
  });
}
