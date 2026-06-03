import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";
import type { IamRole } from "../types";

const STALE_30S = 30_000;

export type MembershipsQueryParams = {
  userId?: string;
  areaCode?: string;
  role?: IamRole;
};

export function useMembershipsQuery(params: MembershipsQueryParams = {}) {
  return useQuery({
    queryKey: QK.iam.memberships.list(params as Record<string, unknown>),
    queryFn: async () => {
      const { data, error } = await api.GET("/iam/area-memberships", {
        params: { query: params },
      });
      if (error) throw error;
      return data;
    },
    staleTime: STALE_30S,
  });
}
