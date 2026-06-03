import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_10MIN = 10 * 60 * 1000;

export function useRolesQuery() {
  return useQuery({
    queryKey: QK.iam.roles(),
    queryFn: async () => {
      const { data, error } = await api.GET("/iam/roles");
      if (error) throw error;
      return data;
    },
    staleTime: STALE_10MIN,
  });
}
