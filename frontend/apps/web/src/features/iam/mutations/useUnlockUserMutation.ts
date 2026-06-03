import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

export function useUnlockUserMutation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (userId: string) => {
      const { data, error } = await api.POST("/iam/users/{userId}/unlock", {
        params: { path: { userId } },
      });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.iam.users() });
    },
  });
}
