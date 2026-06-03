import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_60S = 60_000;

export function useKpiQuery() {
  return useQuery({
    queryKey: QK.iam.kpi(),
    queryFn: async () => {
      const { data, error } = await api.GET("/iam/kpi");
      if (error) throw error;
      return data;
    },
    staleTime: STALE_60S,
  });
}
