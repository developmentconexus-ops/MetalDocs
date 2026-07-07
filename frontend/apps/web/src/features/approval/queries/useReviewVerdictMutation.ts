// F2 (M2c): review-verdict mutation for the review sidebar (F4).
//
// Wraps approvalApi.reviewVerdict. On success it invalidates the whole approval
// subtree (instance + inbox share the 'approval' prefix) plus the documents root
// — matching useSignoffMutation, since a verdict advances the instance and can
// change the document's status. No manual staleness banner (spec C2): the caches
// re-fetch. Error mapping stays in the shared problem+json path (mutationClient).

import { useMutation, useQueryClient } from '@tanstack/react-query';

import { reviewVerdict } from '../api/approvalApi';
import type { ReviewVerdictRequest, ReviewVerdictResponse } from '../api/approvalTypes';
import { QK } from '../../../lib/queryKeys';

export interface ReviewVerdictInput {
  instanceId: string;
  stageId: string;
  etag: string;
  body: ReviewVerdictRequest;
}

export function useReviewVerdictMutation() {
  const queryClient = useQueryClient();

  return useMutation<ReviewVerdictResponse, Error, ReviewVerdictInput>({
    mutationFn: (input) => reviewVerdict(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: QK.approval.all });
      void queryClient.invalidateQueries({ queryKey: QK.documents.all });
    },
  });
}
