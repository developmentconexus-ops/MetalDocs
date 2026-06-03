import { useMutation } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";

type ResetPasswordVariables = {
  userId: string;
  newPassword: string;
};

export function useResetPasswordMutation() {
  return useMutation({
    mutationFn: async ({ userId, newPassword }: ResetPasswordVariables) => {
      const { data, error } = await api.POST("/iam/users/{userId}/reset-password", {
        params: { path: { userId } },
        body: { newPassword },
      });
      if (error) throw error;
      return data;
    },
  });
}
