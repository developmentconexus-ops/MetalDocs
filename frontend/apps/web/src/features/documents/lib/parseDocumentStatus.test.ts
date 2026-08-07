import { describe, it, expect } from 'vitest';
import { parseDocumentStatus } from './parseDocumentStatus';
import { LEGACY_RETIRED_STATUS } from '../../../lib/legacyStatus';

describe('parseDocumentStatus', () => {
  it('passes through every canonical status unchanged', () => {
    const canonical = [
      'draft',
      'under_review',
      'approved',
      'frozen',
      'archived',
      'scheduled',
      'published',
      'superseded',
      'obsolete',
    ] as const;
    for (const status of canonical) {
      expect(parseDocumentStatus(status)).toBe(status);
    }
  });

  it('rejects the pre-cutover legacy aliases instead of mapping them', () => {
    // FE-05: 'review' (renamed to under_review) and the retired literal
    // named LEGACY_RETIRED_STATUS in frontend/apps/web/src/lib/legacyStatus.ts
    // (removed by migration 0142) are no longer emitted by any producer —
    // see the Go DocumentStatus enum
    // (internal/modules/documents/domain/model.go) and the generated
    // DocumentSummaryStatus wire enum. They must not resurface as
    // silently-aliased canonical values.
    expect(parseDocumentStatus('review')).toBeNull();
    expect(parseDocumentStatus(LEGACY_RETIRED_STATUS)).toBeNull(); // parse-boundary rejection
  });

  it('rejects the removed document-status "rejected" (M4 F4.1) — not the live approval-decision "rejected"', () => {
    // documents.status 'rejected' is dead (no producer, backend Task A/B
    // already removed the Go domain const + DB CHECK/trigger arcs). It must
    // not resurface as a canonical document status. This is unrelated to the
    // LIVE approval-decision 'rejected' under features/approval/*.
    expect(parseDocumentStatus('rejected')).toBeNull();
  });

  it('rejects unrecognized and empty values', () => {
    expect(parseDocumentStatus('bogus_status')).toBeNull();
    expect(parseDocumentStatus('')).toBeNull();
    expect(parseDocumentStatus(null)).toBeNull();
    expect(parseDocumentStatus(undefined)).toBeNull();
  });
});
