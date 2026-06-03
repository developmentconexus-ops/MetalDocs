import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_5MIN = 5 * 60 * 1000;

export function useUsageQuery() {
  return useQuery({
    queryKey: QK.iam.usage(),
    queryFn: async () => {
      const { data, error } = await api.GET("/iam/usage");
      if (error) throw error;
      return data;
    },
    staleTime: STALE_5MIN,
  });
}
