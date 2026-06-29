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
// The OpenAI client is created per request. When OPENAI_API_KEY is missing we
// still construct the client with a placeholder so CopilotKit's runtime-info
// request (fired on page load) succeeds — only an actual chat completion will
// fail, surfaced cleanly in the chat UI rather than breaking the whole app.

const runtime = new CopilotRuntime();

export const POST = async (req: NextRequest) => {
  const apiKey = process.env.OPENAI_API_KEY ?? "missing-api-key";
  const openai = new OpenAI({ apiKey });
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
