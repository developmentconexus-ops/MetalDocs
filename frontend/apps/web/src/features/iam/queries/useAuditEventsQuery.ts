import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_30S = 30_000;

type AuditEventsQueryParams = {
  cursor?: string;
  limit?: number;
  action?: string;
  actorId?: string;
  resourceType?: string;
  from?: string;
  to?: string;
};

export function useAuditEventsQuery(params: AuditEventsQueryParams = {}) {
  return useQuery({
    queryKey: QK.iam.audit(params as Record<string, unknown>),
    queryFn: async () => {
      const { data, error } = await api.GET("/audit/events", {
        params: { query: params },
      });
      if (error) throw error;
      return data;
    },
    staleTime: STALE_30S,
  });
}
