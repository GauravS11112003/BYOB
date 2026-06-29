"use client";

export interface SuggestedAction {
  label: string;
  description?: string;
}

export interface ActionPanelProps {
  title: string;
  actions: SuggestedAction[];
  onSelect?: (label: string) => void;
}

// ActionPanel renders AI-suggested next actions as clickable items.
export function ActionPanel({ title, actions, onSelect }: ActionPanelProps) {
  return (
    <div className="byob-card">
      <h3 className="byob-card-title">{title}</h3>
      <div className="byob-actions">
        {actions.map((a) => (
          <button
            key={a.label}
            className="byob-action"
            onClick={() => onSelect?.(a.label)}
          >
            <span className="byob-action-label">{a.label}</span>
            {a.description && (
              <span className="byob-action-desc">{a.description}</span>
            )}
          </button>
        ))}
      </div>
    </div>
  );
}
