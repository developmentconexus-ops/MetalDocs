import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";
import type { components } from "../../../lib/api-types";

type GrantMembershipBody = components["schemas"]["GrantAreaMembershipRequest"];

export function useGrantMembershipMutation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: GrantMembershipBody) => {
      const { data, error } = await api.POST("/iam/area-memberships", { body });
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, vars) => {
      void qc.invalidateQueries({ queryKey: QK.iam.memberships.all });
      // The users listing embeds areaMemberships and the per-user membership
      // view is keyed separately — both go stale on a grant.
      void qc.invalidateQueries({ queryKey: QK.iam.userMemberships(vars.user_id) });
    },
  });
}
