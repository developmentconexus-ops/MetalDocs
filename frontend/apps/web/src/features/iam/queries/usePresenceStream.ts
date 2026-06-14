// Presence stream hook — PR-12b.
//
// Opens a tenant-scoped WebSocket to /api/v1/iam/presence/stream and merges
// snapshot / join / leave / online / idle events into a single items list.
// Heartbeats every 30s match the server-side HeartbeatInterval; the server
// idles the connection after 90s without traffic. Reconnect uses exponential
// backoff up to 30s. After 5 consecutive failed reconnects the hook falls
// back to the legacy 30s HTTP-snapshot polling so the panel keeps rendering.
import { useEffect, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { api, API_BASE_URL } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";
import type { components } from "../../../lib/api-types";

type OnlinePresenceItem = components["schemas"]["OnlinePresenceItem"];
export type PresenceStatus = "online" | "idle";
export type PresenceItem = OnlinePresenceItem & { status?: PresenceStatus };

type PresenceEvent =
  | { type: "snapshot"; presence: PresenceItem[] }
  | { type: "join" | "leave" | "online" | "idle"; user_id: string; username?: string; display_name?: string; last_seen_at?: string; status?: PresenceStatus }
  | { type: "heartbeat" };

type PresenceSnapshot = { items: PresenceItem[] };

const HEARTBEAT_MS = 30_000;
const MAX_BACKOFF_MS = 30_000;
const POLL_FALLBACK_MS = 30_000;
const FALLBACK_AFTER_FAILS = 5;

function streamUrl(): string {
  if (typeof window === "undefined") return "";
  const base = API_BASE_URL;
  // base is typically "/api/v1"; resolve against current origin.
  const httpUrl = base.startsWith("http") ? new URL(base) : new URL(base, window.location.origin);
  httpUrl.pathname = `${httpUrl.pathname.replace(/\/$/, "")}/iam/presence/stream`;
  httpUrl.protocol = httpUrl.protocol === "https:" ? "wss:" : "ws:";
  return httpUrl.toString();
}

export function applyEvent(prev: PresenceItem[], ev: PresenceEvent): PresenceItem[] {
  if (ev.type === "snapshot") return [...ev.presence];
  if (ev.type === "heartbeat") return prev;
  if (ev.type === "leave") return prev.filter((p) => p.user_id !== ev.user_id);
  // join / online / idle: upsert by user_id.
  const existing = prev.find((p) => p.user_id === ev.user_id);
  const next: PresenceItem = {
    user_id: ev.user_id,
    username: ev.username ?? existing?.username ?? "",
    display_name: ev.display_name ?? existing?.display_name ?? "",
    last_seen_at: ev.last_seen_at ?? existing?.last_seen_at ?? new Date().toISOString(),
    status: ev.status ?? existing?.status ?? "online",
  };
  const without = prev.filter((p) => p.user_id !== ev.user_id);
  return [next, ...without];
}

export function usePresenceStream() {
  const queryClient = useQueryClient();
  const [items, setItems] = useState<PresenceItem[] | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [fellBack, setFellBack] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const heartbeatTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const failedAttempts = useRef(0);
  const closedByUs = useRef(false);

  useEffect(() => {
    if (typeof window === "undefined" || typeof WebSocket === "undefined") {
      setFellBack(true);
      return;
    }
    closedByUs.current = false;

    function clearReconnect() {
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
        reconnectTimer.current = null;
      }
    }
    function clearHeartbeat() {
      if (heartbeatTimer.current) {
        clearInterval(heartbeatTimer.current);
        heartbeatTimer.current = null;
      }
    }

    function scheduleReconnect() {
      if (closedByUs.current) return;
      failedAttempts.current += 1;
      if (failedAttempts.current >= FALLBACK_AFTER_FAILS) {
        setFellBack(true);
        return;
      }
      const delay = Math.min(MAX_BACKOFF_MS, 1000 * 2 ** (failedAttempts.current - 1));
      clearReconnect();
      reconnectTimer.current = setTimeout(connect, delay);
    }

    function connect() {
      clearReconnect();
      let ws: WebSocket;
      try {
        ws = new WebSocket(streamUrl());
      } catch (err) {
        setError(err);
        scheduleReconnect();
        return;
      }
      wsRef.current = ws;

      ws.onopen = () => {
        failedAttempts.current = 0;
        setError(null);
        setFellBack(false);
        clearHeartbeat();
        heartbeatTimer.current = setInterval(() => {
          if (ws.readyState === WebSocket.OPEN) {
            try {
              ws.send(JSON.stringify({ type: "ping" }));
            } catch {
              // best-effort heartbeat
            }
          }
        }, HEARTBEAT_MS);
      };

      ws.onmessage = (msg) => {
        let parsed: PresenceEvent;
        try {
          parsed = JSON.parse(typeof msg.data === "string" ? msg.data : "") as PresenceEvent;
        } catch {
          return;
        }
        setItems((prev) => {
          const base = prev ?? [];
          const next = applyEvent(base, parsed);
          queryClient.setQueryData<PresenceSnapshot>(QK.iam.presenceSnapshot(), { items: next });
          return next;
        });
      };

      ws.onerror = (ev) => {
        setError(ev);
      };

      ws.onclose = () => {
        clearHeartbeat();
        wsRef.current = null;
        scheduleReconnect();
      };
    }

    connect();

    return () => {
      closedByUs.current = true;
      clearReconnect();
      clearHeartbeat();
      if (wsRef.current && wsRef.current.readyState <= WebSocket.OPEN) {
        wsRef.current.close();
      }
      wsRef.current = null;
    };
  }, [queryClient]);

  // Fallback polling — enabled only after FALLBACK_AFTER_FAILS reconnects fail.
  const pollQuery = useQuery({
    queryKey: QK.iam.presenceSnapshot(),
    queryFn: async () => {
      const { data, error: gErr } = await api.GET("/iam/presence/snapshot");
      if (gErr) throw gErr;
      return data;
    },
    enabled: fellBack,
    refetchInterval: fellBack ? POLL_FALLBACK_MS : false,
    staleTime: POLL_FALLBACK_MS,
  });

  if (fellBack) {
    return {
      data: pollQuery.data as PresenceSnapshot | undefined,
      isLoading: pollQuery.isLoading,
      error: pollQuery.error,
    };
  }

  return {
    data: items != null ? ({ items } as PresenceSnapshot) : undefined,
    isLoading: items == null && error == null,
    error,
  };
}
