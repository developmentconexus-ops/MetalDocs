import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { QK } from '../../../lib/queryKeys';
import { resolveErrorMessage } from '../../../lib/api/problem';
import { extendSLA } from '../api/approvalApi';

export interface ExtendSLAVariables {
  instanceId: string;
  /** RFC3339 timestamp. Any offset is accepted server-side; not required to be UTC. */
  dueAt: string;
  reason: string;
}

/**
 * Task 10: postpones the active stage's due_at for ONE instance — the UI side
 * of Task 9's `POST /approval/instances/{id}/extend-sla`. The route's
 * due_in_days is the standard; this is the recorded, auditable exception, not
 * a convenience edit. Mirrors `useRouteAdminMutations.ts`: mutation hook +
 * `QK` + `resolveErrorMessage` toast, `Idempotency-Key` generated per attempt
 * by the shared `mutationClient` (via `extendSLA` -> `mutate`, opts left
 * empty so a fresh key is minted for this call).
 *
 * No optimistic update: the response carries the authoritative due_at and
 * there is nothing safe to guess client-side before it returns (spec
 * `wiki/architecture/frontend-structure.md` §8 — opt in only when the
 * rollback model is obvious and safe; it is not, for a governed deadline).
 *
 * Blank reason never reaches the network — mirrors the server's
 * `validation.sla_extension_reason_required` (422) client-side so an
 * obviously invalid request never fires. Every other rejection (not forward,
 * no active stage) is a live round-trip surfaced through the shared PT-BR
 * problem+json mapping.
 *
 * `ExtendSLAVariables` deliberately carries only `instanceId` (per the Task
 * 10 contract), not the `documentId` that `QK.approval.instance(documentId)`
 * is keyed by (see `useApprovalInstanceQuery`) — the cache key is
 * document-scoped, the mutation is instance-scoped, and the two ids are not
 * derivable from each other client-side. Without `documentId` a targeted
 * `QK.approval.instance(documentId)` invalidation is not possible, so success
 * invalidates the whole `QK.approval.all` subtree instead (still confined to
 * the approval module's own cache root, no broader collateral invalidation).
 */
export function useExtendSLA() {
  const queryClient = useQueryClient();

  return useMutation<ExtendSLAResult, Error, ExtendSLAVariables>({
    mutationFn: ({ instanceId, dueAt, reason }) => {
      const trimmedReason = reason.trim();
      if (trimmedReason === '') {
        return Promise.reject(new Error('Informe o motivo da prorrogação do prazo.'));
      }
      return extendSLA(instanceId, { due_at: dueAt, reason: trimmedReason });
    },
    onSuccess: () => {
      toast.success('Prazo prorrogado.');
      void queryClient.invalidateQueries({ queryKey: QK.approval.all });
    },
    onError: (err) => {
      toast.error(resolveErrorMessage(err));
    },
  });
}

type ExtendSLAResult = Awaited<ReturnType<typeof extendSLA>>;
