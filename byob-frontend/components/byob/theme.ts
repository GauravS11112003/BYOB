// BYOB brand tokens. Change these to rebrand the entire app.
// This is the single source of truth for product name, colors, and copy.

export const brand = {
  name: "BYOB",
  tagline: "Bring Your Own Backend",
  description:
    "Connect your data sources and let AI render your dashboard.",

  colors: {
    primary: "#6366f1", // indigo-500
    primaryDark: "#4f46e5", // indigo-600
    accent: "#8b5cf6", // violet-500
    bg: "#0b0f1a",
    surface: "#111827",
    surfaceAlt: "#1f2937",
    border: "#374151",
    text: "#f3f4f6",
    textMuted: "#9ca3af",
    success: "#22c55e",
    danger: "#ef4444",
  },

  chat: {
    title: "BYOB Assistant",
    initial:
      "Hi! I can visualize and explain your connected data. Try: \"show me a chart of the latest metrics\".",
    placeholder: "Ask about your data…",
  },
} as const;

export type Brand = typeof brand;

// The Go backend base URL for the data plane (SSE stream + connector metadata).
export const BACKEND_URL =
  process.env.NEXT_PUBLIC_BYOB_BACKEND_URL ?? "http://localhost:8080";
