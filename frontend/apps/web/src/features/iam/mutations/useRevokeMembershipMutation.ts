import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

export type RevokeMembershipVars = {
  userId: string;
  areaCode: string;
};

export function useRevokeMembershipMutation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ userId, areaCode }: RevokeMembershipVars) => {
      const { data, error } = await api.DELETE(
        "/iam/area-memberships/{user_id}/{area_code}",
        { params: { path: { user_id: userId, area_code: areaCode } } },
      );
      if (error) throw error;
      return data;
    },
    onSuccess: (_data, vars) => {
      void qc.invalidateQueries({ queryKey: QK.iam.memberships.all });
      // The users listing embeds areaMemberships and the per-user membership
      // view is keyed separately — both go stale on a revoke.
      void qc.invalidateQueries({ queryKey: QK.iam.userMemberships(vars.userId) });
    },
  });
}
