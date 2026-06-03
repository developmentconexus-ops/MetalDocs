import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";
import type { IamRole } from "../types";

const STALE_30S = 30_000;

type UsersQueryParams = {
  cursor?: string;
  limit?: number;
  q?: string;
  isActive?: boolean;
  role?: IamRole;
  areaCode?: string;
};

export function useUsersQuery(params: UsersQueryParams = {}) {
  return useQuery({
    queryKey: QK.iam.users(params as Record<string, unknown>),
    queryFn: async () => {
      const { data, error } = await api.GET("/iam/users", {
        params: { query: params },
      });
      if (error) throw error;
      return data;
    },
    staleTime: STALE_30S,
  });
}
