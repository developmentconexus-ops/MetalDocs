import { useInfiniteQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";
import type { components, operations } from "../../../lib/api-types";

const STALE_30S = 30_000;

// Derived from the generated contract so the snake_case wire keys (actor_id,
// resource_type, resource_id, occurred_after, occurred_before) can never drift
// from the spec.
export type AuditEventsQueryParams = NonNullable<
  operations["listAuditEvents"]["parameters"]["query"]
>;

export type AuditEventItem = components["schemas"]["AuditEventItem"];

export type AuditEventsPage = {
  items: ReadonlyArray<AuditEventItem>;
  nextCursor: string | null;
  hasMore: boolean;
};

// Canonical cursor envelope: { items, page:{ next_cursor, has_more } } — matches
// the OpenAPI ListAuditEventsResponse / CursorPage and the runtime handler. (The
// prior flat-shape fallback + drift was closed in Phase F; see
// wiki/decisions/0028-audit-events-cursor-shape.md.)
function adaptPage(raw: unknown): AuditEventsPage {
  if (!raw || typeof raw !== "object") {
    return { items: [], nextCursor: null, hasMore: false };
  }
  const r = raw as {
    items?: unknown;
    page?: { next_cursor?: unknown; has_more?: unknown };
  };
  const items = Array.isArray(r.items) ? (r.items as AuditEventItem[]) : [];
  const nextCursor =
    typeof r.page?.next_cursor === "string" ? r.page.next_cursor : null;
  const hasMore =
    typeof r.page?.has_more === "boolean" ? r.page.has_more : Boolean(nextCursor);
  return { items, nextCursor, hasMore };
}

export function useAuditEventsQuery(params: AuditEventsQueryParams = {}) {
  return useInfiniteQuery({
    queryKey: QK.iam.audit(params as Record<string, unknown>),
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam }): Promise<AuditEventsPage> => {
      const { data, error } = await api.GET("/audit/events", {
        params: { query: { ...params, cursor: pageParam } },
      });
      if (error) throw error;
      return adaptPage(data);
    },
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
    staleTime: STALE_30S,
  });
}
