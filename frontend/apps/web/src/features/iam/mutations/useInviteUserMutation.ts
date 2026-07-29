import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";
import type { IamRole } from "../types";
import type { AreaRole } from "../../../lib/iam/role-vocabulary";

// F-QA4-2: an invite-time area membership becomes a user_process_areas row, so
// its role is drawn from AreaRole (the seven area-assignable roles), NOT the
// broader 8-value tenant enum. Mirrors AreaMembershipInput.role in the contract.
type AreaMembershipInput = {
  areaCode: string;
  role: AreaRole;
};

type InviteUserBody = {
  username: string;
  email: string;
  displayName: string;
  tenantRole: IamRole;
  areaMemberships?: AreaMembershipInput[];
};

export function useInviteUserMutation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: InviteUserBody) => {
      const { data, error } = await api.POST("/iam/users/invite", {
        body: {
          username: body.username,
          email: body.email,
          display_name: body.displayName,
          tenant_role: body.tenantRole,
          area_memberships: body.areaMemberships?.map((m) => ({ area_code: m.areaCode, role: m.role })),
        },
      });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.iam.users() });
    },
  });
}
