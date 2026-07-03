import { describe, it, expect } from 'vitest';
import { parseDocumentStatus } from './parseDocumentStatus';

describe('parseDocumentStatus', () => {
  it('passes through every canonical status unchanged', () => {
    const canonical = [
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
    ] as const;
    for (const status of canonical) {
      expect(parseDocumentStatus(status)).toBe(status);
    }
  });

  it('rejects the pre-cutover legacy aliases instead of mapping them', () => {
    // FE-05: 'review' (renamed to under_review) and 'finalized' (removed by
    // migration 0142) are no longer emitted by any producer — see the Go
    // DocumentStatus enum (internal/modules/documents/domain/model.go) and the
    // generated DocumentSummaryStatus wire enum. They must not resurface as
    // silently-aliased canonical values.
    expect(parseDocumentStatus('review')).toBeNull();
    expect(parseDocumentStatus('finalized')).toBeNull();
  });

  it('rejects unrecognized and empty values', () => {
    expect(parseDocumentStatus('bogus_status')).toBeNull();
    expect(parseDocumentStatus('')).toBeNull();
    expect(parseDocumentStatus(null)).toBeNull();
    expect(parseDocumentStatus(undefined)).toBeNull();
  });
});
