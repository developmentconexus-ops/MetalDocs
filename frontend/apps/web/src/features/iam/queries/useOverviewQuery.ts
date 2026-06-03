import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_30S = 30_000;

export function useOverviewQuery() {
  return useQuery({
    queryKey: QK.iam.adminOverview(),
    queryFn: async () => {
      const { data, error } = await api.GET("/iam/admin/overview");
      if (error) throw error;
      return data;
    },
    staleTime: STALE_30S,
  });
}
