import { useQuery } from '@tanstack/react-query';
import { fetchCreationContext, type CreationContextResponse } from '../api/controlledDocuments';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from '../../documents/queries/_constants';

/**
 * Pre-flight context for the create-document wizard: the tenant's active
 * profiles annotated with approval-route readiness, plus ONLY the process
 * areas in which the caller holds `controlled_documents.create`.
 *
 * This is the authority for what the wizard may offer. The area narrowing is
 * computed server-side from the caller's capability grants — never re-derive
 * it from the raw taxonomy catalog (`useAreasQuery`), which is the tenant-wide
 * list and carries no per-caller authorization.
 *
 * The route itself is tier-1 gated on `controlled_documents.create`, so a
 * caller who may not create anywhere gets a 403 here rather than an empty
 * form that only fails at submit time.
 */
export function useCreationContextQuery() {
  return useQuery<CreationContextResponse>({
    queryKey: QK.controlledDocuments.creationContext(),
    queryFn: () => fetchCreationContext(),
    staleTime: STALE_FIVE_MINUTES,
  });
}
