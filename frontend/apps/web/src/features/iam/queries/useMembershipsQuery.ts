import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";
import type { operations } from "../../../lib/api-types";

const STALE_30S = 30_000;

// Derived from the generated contract so the snake_case wire keys
// (user_id, area_code) can never drift from the spec.
export type MembershipsQueryParams = NonNullable<
  operations["listAreaMemberships"]["parameters"]["query"]
>;

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
