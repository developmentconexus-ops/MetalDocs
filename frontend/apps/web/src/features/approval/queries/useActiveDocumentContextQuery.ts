import { useQuery } from '@tanstack/react-query';
import { getActiveDocumentContext } from '../api/approvalApi';
import { QK } from '../../../lib/queryKeys';

/**
 * Resolves the approval context (content_hash, approval_instance_id,
 * revision_version, approval_state) for a controlled document. Keyed by the
 * controlled-document id because that is what the producer endpoint accepts;
 * the cockpit reaches it from the document's controlled_document_id.
 */
export function useActiveDocumentContextQuery(controlledDocumentId: string) {
  return useQuery({
    queryKey: QK.controlledDocuments.activeDocument(controlledDocumentId),
    queryFn: () => getActiveDocumentContext(controlledDocumentId),
    enabled: Boolean(controlledDocumentId),
    staleTime: 10_000,
  });
}
