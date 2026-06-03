import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_5MIN = 5 * 60 * 1000;

// NOTE: /security/signals has no query params in codegen (window param not exposed).
// The window arg is kept for the query key only (cache partitioning).
export function useSecuritySignalsQuery(window = "7d") {
  return useQuery({
    queryKey: QK.iam.securitySignals(window),
    queryFn: async () => {
      const { data, error } = await api.GET("/security/signals");
      if (error) throw error;
      return data;
    },
    staleTime: STALE_5MIN,
  });
}
