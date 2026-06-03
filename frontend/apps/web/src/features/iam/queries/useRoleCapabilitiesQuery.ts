import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_10MIN = 10 * 60 * 1000;

export function useRoleCapabilitiesQuery() {
  return useQuery({
    queryKey: QK.iam.roleCapabilities(),
    queryFn: async () => {
      const { data, error } = await api.GET("/iam/role-capabilities");
      if (error) throw error;
      return data;
    },
    staleTime: STALE_10MIN,
  });
}
