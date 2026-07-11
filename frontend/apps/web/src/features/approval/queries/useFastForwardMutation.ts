// G3 (unit 2.4): fast-forward mutation for the "Aprovar já" gesture — records
// the ready verdict leg + the next-stage approve-signoff leg in one tx.
// Mirrors useReviewVerdictMutation.ts: same invalidation subtree (approval
// instance + inbox share the 'approval' prefix, plus documents — a
// fast-forward advances the instance and can change the document's status).
// No manual staleness banner; error mapping stays in the shared problem+json
// path (mutationClient).

import { useMutation, useQueryClient } from '@tanstack/react-query';

import { fastForward } from '../api/approvalApi';
import type { FastForwardRequest, FastForwardResponse } from '../api/approvalTypes';
import { QK } from '../../../lib/queryKeys';

export interface FastForwardInput {
  instanceId: string;
  stageId: string;
  etag: string;
  body: FastForwardRequest;
}

export function useFastForwardMutation() {
  const queryClient = useQueryClient();

  return useMutation<FastForwardResponse, Error, FastForwardInput>({
    mutationFn: (input) => fastForward(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: QK.approval.all });
      void queryClient.invalidateQueries({ queryKey: QK.documents.all });
    },
  });
}
