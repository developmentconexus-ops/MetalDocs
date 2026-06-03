import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

const STALE_30S = 30_000;

export function useUserMembershipsQuery(userId: string) {
  return useQuery({
    queryKey: QK.iam.userMemberships(userId),
    queryFn: async () => {
      const { data, error } = await api.GET("/iam/users/{userId}/memberships", {
        params: { path: { userId } },
      });
      if (error) throw error;
      return data;
    },
    staleTime: STALE_30S,
    enabled: userId.length > 0,
  });
}
