import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

type PatchUserVariables = {
  userId: string;
  body: Record<string, unknown>;
};

export function usePatchUserMutation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ userId, body }: PatchUserVariables) => {
      const { data, error } = await api.PATCH("/iam/users/{userId}", {
        params: { path: { userId } },
        body,
      });
      if (error) throw error;
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.iam.users() });
    },
  });
}
