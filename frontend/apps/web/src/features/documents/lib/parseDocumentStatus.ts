// Single source of truth for narrowing a wire-shaped document status string
// (DocumentDetailResponse.status is `type: string` in the OpenAPI contract —
// not an enum, see api/openapi/v1/openapi.yaml DocumentDetailResponse) into
// the canonical `DocumentStatus` union rendered by StatusPill.
//
// FE-05: 'review' and 'finalized' were pre-cutover aliases (migration 0142
// removed draft->finalized; 'review' was renamed to 'under_review'). The Go
// domain enum (internal/modules/documents/domain/model.go DocumentStatus) and
// the generated DocumentSummaryStatus wire enum (internal/modules/documents/
// api/api.gen.go) both confirm neither value is reachable anymore — no FE or
// BE producer emits them today. They are treated as unrecognized/legacy and
// rejected (mapped to null) rather than silently aliased to a canonical value,
// so any resurfacing shows up as a gap instead of being masked.
//
// Anywhere that narrows a raw `doc.status` string into `DocumentStatus` should
// route through this function instead of hand-maintaining its own allow-list.

import type { DocumentStatus } from '../../../components/ui/StatusPill';

const CANONICAL_STATUSES: readonly DocumentStatus[] = [
  'draft',
  'under_review',
  'approved',
  'frozen',
  'rejected',
  'archived',
  'scheduled',
  'published',
  'superseded',
  'obsolete',
];

/**
 * Narrows a raw wire status string to the canonical `DocumentStatus` union.
 * Returns `null` for empty, unrecognized, or legacy pre-cutover values
 * (e.g. 'review', 'finalized') rather than aliasing them. (cilint:allow-legacy —
 * parse-boundary rejection doc, see tests/integration/scenarios/legacy_absent_test.go)
 */
export function parseDocumentStatus(status: string | null | undefined): DocumentStatus | null {
  if (!status) return null;
  return (CANONICAL_STATUSES as readonly string[]).includes(status)
    ? (status as DocumentStatus)
    : null;
}
