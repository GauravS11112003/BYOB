"use client";

import dynamic from "next/dynamic";

// The dashboard relies on the browser (EventSource) and registers CopilotKit
// actions, so it must run client-side only. Disabling SSR avoids prerendering
// the live data plane and generative actions at build time.
const Dashboard = dynamic(() => import("@/components/dashboard/Dashboard"), {
  ssr: false,
  loading: () => <div style={{ padding: 32 }}>Loading BYOB…</div>,
});

export default function Page() {
  return <Dashboard />;
}
