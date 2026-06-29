"use client";

import { useVisualizeAction } from "./visualize";
import { useSummarizeAction, useTableAction } from "./query";
import { useSuggestActionsAction } from "./act";

// useByobActions registers all generative UI actions in one call. Mount it once
// inside a component under ByobProvider.
export function useByobActions() {
  useVisualizeAction();
  useTableAction();
  useSummarizeAction();
  useSuggestActionsAction();
}
