import { useMutation } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";

type ResetPasswordVariables = {
  userId: string;
  newPassword: string;
};

export function useResetPasswordMutation() {
  return useMutation({
    mutationFn: async ({ userId, newPassword }: ResetPasswordVariables) => {
      const { data, error } = await api.POST("/iam/users/{user_id}/reset-password", {
        params: { path: { user_id: userId } },
        body: { new_password: newPassword },
      });
      if (error) throw error;
      return data;
    },
  });
}
