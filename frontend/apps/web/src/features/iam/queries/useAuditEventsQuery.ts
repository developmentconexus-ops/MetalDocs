import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_30S = 30_000;

export type AuditEventsQueryParams = {
  cursor?: string;
  limit?: number;
  action?: string;
  actorId?: string;
  resourceType?: string;
  resourceId?: string;
  occurredAfter?: string;
  occurredBefore?: string;
  q?: string;
};

export type AuditEventItem = {
  id: string;
  occurredAt: string;
  actorId: string;
  action: string;
  resourceType: string;
  resourceId: string;
  payload: Record<string, unknown>;
  traceId?: string;
};

export type AuditEventsPage = {
  items: ReadonlyArray<AuditEventItem>;
  nextCursor: string | null;
  hasMore: boolean;
};

// Backend currently returns `{items, nextCursor, hasMore}` (flat) while the
// generated OpenAPI shape declares `page: { next_cursor, has_more }`. Adapt
// here so callers see a single, stable structure regardless of which shape
// the server emits on a given deploy.
function adaptPage(raw: unknown): AuditEventsPage {
  if (!raw || typeof raw !== "object") {
    return { items: [], nextCursor: null, hasMore: false };
  }
  const r = raw as {
    items?: unknown;
    nextCursor?: unknown;
    hasMore?: unknown;
    page?: { next_cursor?: unknown; has_more?: unknown };
  };
  const items = Array.isArray(r.items) ? (r.items as AuditEventItem[]) : [];
  const flatCursor = typeof r.nextCursor === "string" ? r.nextCursor : null;
  const nestedCursor =
    typeof r.page?.next_cursor === "string" ? r.page.next_cursor : null;
  const flatHas = typeof r.hasMore === "boolean" ? r.hasMore : undefined;
  const nestedHas =
    typeof r.page?.has_more === "boolean" ? r.page.has_more : undefined;
  const nextCursor = flatCursor ?? nestedCursor;
  const hasMore = flatHas ?? nestedHas ?? Boolean(nextCursor);
  return { items, nextCursor, hasMore };
}

export function useAuditEventsQuery(params: AuditEventsQueryParams = {}) {
  return useQuery({
    queryKey: QK.iam.audit(params as Record<string, unknown>),
    queryFn: async (): Promise<AuditEventsPage> => {
      const { data, error } = await api.GET("/audit/events", {
        params: { query: params },
      });
      if (error) throw error;
      return adaptPage(data);
    },
    staleTime: STALE_30S,
  });
}
