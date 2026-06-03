// TODO PR-12 Fase 2: WS upgrade — replace snapshot polling with WebSocket stream.
import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_30S = 30_000;

export function usePresenceStream() {
  const query = useQuery({
    queryKey: QK.iam.presenceSnapshot(),
    queryFn: async () => {
      const { data, error } = await api.GET("/iam/presence/snapshot");
      if (error) throw error;
      return data;
    },
    staleTime: STALE_30S,
  });

  return {
    data: query.data,
    isLoading: query.isLoading,
    error: query.error,
  };
}
