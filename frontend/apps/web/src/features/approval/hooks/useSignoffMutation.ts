// Shared document sign-off mutation.
//
// Lifted from SignoffDialog so the inline decision panel (shared cockpit) and any
// remaining caller drive the SAME react-query mutation, cache invalidation, and
// error classification — no forked copy. The hook owns the network call + cache
// side effects; error copy lives in `mapSignoffError` (pure). Callers get an
// async `signOff` that resolves on success and throws a `SignoffError` on failure
// (its `.message` is user-ready, `.stale` flags the 412 refresh case).

import { useMutation, useQueryClient } from '@tanstack/react-query';

import { signoff } from '../api/approvalApi';
import { mapSignoffError } from '../lib/signoffErrors';

export interface SignoffInput {
  decision: 'approve' | 'reject';
  reason?: string;
  password: string;
}

export interface UseSignoffMutationArgs {
  documentId: string;
  contentHash: string;
  revisionVersion: number;
}

export interface UseSignoffMutationResult {
  signOff: (input: SignoffInput) => Promise<void>;
  isPending: boolean;
}

export function useSignoffMutation({
  documentId,
  contentHash,
  revisionVersion,
}: UseSignoffMutationArgs): UseSignoffMutationResult {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (input: SignoffInput) =>
      signoff(
        documentId,
        {
          decision: input.decision,
          reason: input.reason,
          password: input.password,
          content_hash: contentHash,
        },
        { ifMatch: `"v${revisionVersion}"` },
      ),
    onSuccess: () => {
      // F-FE-NOINVALIDATE: a sign-off changes the inbox, the approval instance,
      // and the document's status — invalidate every affected cache subtree
      // instead of refetching a single query.
      void queryClient.invalidateQueries({ queryKey: ['approval'] });
      void queryClient.invalidateQueries({ queryKey: ['documents'] });
    },
  });

  const signOff = async (input: SignoffInput): Promise<void> => {
    try {
      await mutation.mutateAsync(input);
    } catch (error) {
      throw mapSignoffError(error);
    }
  };

  return { signOff, isPending: mutation.isPending };
}
