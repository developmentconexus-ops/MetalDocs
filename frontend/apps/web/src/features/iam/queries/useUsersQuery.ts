import { useQuery } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";
import type { operations } from "../../../lib/api-types";

const STALE_30S = 30_000;

// Derived from the generated contract so the snake_case wire keys
// (is_active, area_code) can never drift from the spec.
type UsersQueryParams = NonNullable<operations["listUsers"]["parameters"]["query"]>;

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
