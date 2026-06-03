import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import { QK } from "../../../lib/queryKeys";

type LockoutsCacheShape = {
  items: Array<{ userId: string; [k: string]: unknown }>;
  [k: string]: unknown;
};

type Context = {
  prevLockouts?: LockoutsCacheShape;
};

export function useUnlockUserMutation() {
  const qc = useQueryClient();
  return useMutation<unknown, Error, string, Context>({
    mutationFn: async (userId: string) => {
      const { data, error } = await api.POST("/iam/users/{userId}/unlock", {
        params: { path: { userId } },
      });
      if (error) throw error;
      return data;
    },
    onMutate: async (userId) => {
      const key = QK.iam.lockouts();
      await qc.cancelQueries({ queryKey: key });
      const prevLockouts = qc.getQueryData<LockoutsCacheShape>(key);
      if (prevLockouts) {
        qc.setQueryData<LockoutsCacheShape>(key, {
          ...prevLockouts,
          items: prevLockouts.items.filter((l) => l.userId !== userId),
        });
      }
      return { prevLockouts };
    },
    onError: (_err, _userId, ctx) => {
      if (ctx?.prevLockouts) {
        qc.setQueryData(QK.iam.lockouts(), ctx.prevLockouts);
      }
    },
    onSettled: () => {
      void qc.invalidateQueries({ queryKey: QK.iam.lockouts() });
      void qc.invalidateQueries({ queryKey: QK.iam.users() });
    },
  });
}
