// F2 (M2c): delegation create/revoke mutations.
//
// Both invalidate the whole approval subtree on success so the inbox (which the
// delegated eligibility changes) re-fetches. delegator_id is session-derived
// server-side — the create body carries only delegate/window/reason. Error
// mapping stays in the shared problem+json path (mutationClient).

import { useMutation, useQueryClient } from '@tanstack/react-query';

import { createDelegation, revokeDelegation } from '../api/approvalApi';
import type {
  ApprovalDelegation,
  CreateApprovalDelegationRequest,
} from '../api/approvalTypes';
import { QK } from '../../../lib/queryKeys';

export function useCreateDelegationMutation() {
  const queryClient = useQueryClient();

  return useMutation<ApprovalDelegation, Error, CreateApprovalDelegationRequest>({
    mutationFn: (body) => createDelegation(body),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: QK.approval.all });
    },
  });
}

export function useRevokeDelegationMutation() {
  const queryClient = useQueryClient();

  return useMutation<void, Error, string>({
    mutationFn: (id) => revokeDelegation(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: QK.approval.all });
    },
  });
}
