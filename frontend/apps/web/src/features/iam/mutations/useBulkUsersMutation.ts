import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

type BulkUsersBody = {
  userIds: string[];
  action: "activate" | "deactivate" | "unlock" | "force-logout";
};

export function useBulkUsersMutation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: BulkUsersBody) => {
      const { data, error } = await api.POST("/iam/users/bulk", { body });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.iam.users() });
    },
  });
}
