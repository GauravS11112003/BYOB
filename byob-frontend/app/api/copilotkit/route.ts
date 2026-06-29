import {
  CopilotRuntime,
  OpenAIAdapter,
  copilotRuntimeNextJSAppRouterEndpoint,
} from "@copilotkit/runtime";
import OpenAI from "openai";
import { NextRequest } from "next/server";

// AI plane. CopilotKit's runtime needs to speak its own protocol, so we host it
// here in a Next.js route. The Go backend owns the data plane (SSE stream +
// connectors); this route owns chat + tool-calling for generative UI.
//
// The runtime and adapter are created lazily per request so a missing API key
// surfaces as a request-time error rather than breaking the build.

const runtime = new CopilotRuntime();

export const POST = async (req: NextRequest) => {
  const openai = new OpenAI({ apiKey: process.env.OPENAI_API_KEY });
  const serviceAdapter = new OpenAIAdapter({
    openai,
    model: process.env.OPENAI_MODEL ?? "gpt-4o",
  });

  const { handleRequest } = copilotRuntimeNextJSAppRouterEndpoint({
    runtime,
    serviceAdapter,
    endpoint: "/api/copilotkit",
  });
  return handleRequest(req);
};
