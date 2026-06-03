import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_30S = 30_000;

type SessionsQueryParams = {
  cursor?: string;
  limit?: number;
};

export function useSessionsQuery(params: SessionsQueryParams = {}) {
  return useQuery({
    queryKey: QK.iam.sessions(params as Record<string, unknown>),
    queryFn: async () => {
      const { data, error } = await api.GET("/auth/sessions", {
        params: { query: params },
      });
      if (error) throw error;
      return data;
    },
    staleTime: STALE_30S,
  });
}
