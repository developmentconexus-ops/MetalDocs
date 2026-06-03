import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";
import type { IamRole } from "../types";

type AreaMembershipInput = {
  areaCode: string;
  role: IamRole;
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
      const { data, error } = await api.POST("/iam/users/invite", { body });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.iam.users() });
    },
  });
}
