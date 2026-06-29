"use client";

// ByobChat wraps CopilotKit's sidebar chat with BYOB branding and copy. The rest
// of the app imports ByobChat, never CopilotSidebar directly.

import { CopilotSidebar } from "@copilotkit/react-ui";
import { brand } from "./theme";

export function ByobChat() {
  return (
    <CopilotSidebar
      labels={{
        title: brand.chat.title,
        initial: brand.chat.initial,
        placeholder: brand.chat.placeholder,
      }}
      defaultOpen={true}
      clickOutsideToClose={false}
    />
  );
}
