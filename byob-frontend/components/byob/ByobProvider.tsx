"use client";

// ByobProvider is the ONLY place that imports CopilotKit's provider. Wrapping it
// here isolates the dependency so the rest of the app stays vendor-neutral and
// rebranding lives in one spot.

import { CopilotKit } from "@copilotkit/react-core";
import "@copilotkit/react-ui/styles.css";
import type { ReactNode } from "react";

export function ByobProvider({ children }: { children: ReactNode }) {
  return (
    <CopilotKit runtimeUrl="/api/copilotkit" showDevConsole={false}>
      {children}
    </CopilotKit>
  );
}
