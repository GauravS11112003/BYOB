"use client";

import { useEffect, useRef, useState } from "react";
import { BACKEND_URL } from "@/components/byob/theme";

// A single normalized record from the Go data plane (mirrors connectors.Record).
export interface DataRecord {
  source: string;
  timestamp: string;
  data: Record<string, unknown>;
}

export interface DataStreamState {
  records: DataRecord[];
  connected: boolean;
  error: string | null;
}

// useDataStream subscribes to the Go backend's SSE endpoint and keeps a rolling
// buffer of the most recent records. It auto-reconnects on disconnect.
export function useDataStream(maxRecords = 500): DataStreamState {
  const [records, setRecords] = useState<DataRecord[]>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    let closed = false;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

    const connect = () => {
      if (closed) return;
      const es = new EventSource(`${BACKEND_URL}/stream`);
      esRef.current = es;

      es.onopen = () => {
        setConnected(true);
        setError(null);
      };

      es.onmessage = (event) => {
        try {
          const rec = JSON.parse(event.data) as DataRecord;
          setRecords((prev) => {
            const next = [...prev, rec];
            return next.length > maxRecords ? next.slice(-maxRecords) : next;
          });
        } catch {
          // ignore malformed frames
        }
      };

      es.onerror = () => {
        setConnected(false);
        es.close();
        if (!closed) {
          setError("Disconnected from backend. Reconnecting…");
          reconnectTimer = setTimeout(connect, 2000);
        }
      };
    };

    connect();

    return () => {
      closed = true;
      if (reconnectTimer) clearTimeout(reconnectTimer);
      esRef.current?.close();
    };
  }, [maxRecords]);

  return { records, connected, error };
}
