"use client";

import React from "react";

interface Props {
  children: React.ReactNode;
}

interface State {
  error: Error | null;
}

// WidgetBoundary isolates rendering failures: a single broken generative widget
// shows an inline error instead of crashing the whole dashboard.
export class WidgetBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  render() {
    if (this.state.error) {
      return (
        <div className="byob-card byob-widget-error">
          <strong>Widget failed to render</strong>
          <p>{this.state.error.message}</p>
        </div>
      );
    }
    return this.props.children;
  }
}
