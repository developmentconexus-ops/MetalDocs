// Characterization test for the current (post-ADR-0080) document sign-off contract.
//
// F5.1's decision surface was relocated from the retired SignoffDetailPage cockpit
// into the mode-adaptive document workspace; the decision network call now lives in
// this hook (the deleted SignoffDialog + its SignoffDialog.test guarded the same
// contract). These tests re-establish that guard against the current code: the
// signature re-auth POSTs to the real /signoff endpoint with the source content_hash
// and the revision-derived If-Match ETag, and a 412 conflict classifies as the
// non-terminal `stale` refresh case rather than a hard error.

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { ApprovalError } from '../api/mutationClient';
import * as approvalApi from '../api/approvalApi';
import { SignoffError } from '../lib/signoffErrors';
import { useSignoffMutation } from './useSignoffMutation';

vi.mock('../api/approvalApi', async (importOriginal) => {
  const actual = await importOriginal<typeof approvalApi>();
  return { ...actual, signoff: vi.fn() };
});

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({
    defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
  });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

const ARGS = { documentId: 'doc-1', contentHash: 'sha256:abc', revisionVersion: 3 };

describe('useSignoffMutation', () => {
  afterEach(() => vi.clearAllMocks());

  it('POSTs the sign-off with the source content_hash body and revision-derived If-Match', async () => {
    vi.mocked(approvalApi.signoff).mockResolvedValue(undefined as never);

    const { result } = renderHook(() => useSignoffMutation(ARGS), { wrapper });
    await result.current.signOff({ decision: 'approve', reason: 'ok', password: 'pw' });

    expect(approvalApi.signoff).toHaveBeenCalledWith(
      'doc-1',
      { decision: 'approve', reason: 'ok', password: 'pw', content_hash: 'sha256:abc' },
      { ifMatch: '"v3"' },
    );
  });

  it('classifies a 412 conflict as a non-terminal stale refresh, not a hard error', async () => {
    vi.mocked(approvalApi.signoff).mockRejectedValue(new ApprovalError('conflict.stale', 412, 'stale'));

    const { result } = renderHook(() => useSignoffMutation(ARGS), { wrapper });
    await expect(result.current.signOff({ decision: 'reject', password: 'pw' })).rejects.toMatchObject({
      name: 'SignoffError',
      kind: 'stale',
      stale: true,
    });
  });

  it('classifies an invalid-signature failure as a SignoffError (bad_password)', async () => {
    vi.mocked(approvalApi.signoff).mockRejectedValue(
      new ApprovalError('authn.signature_invalid', 422, 'invalid'),
    );

    const { result } = renderHook(() => useSignoffMutation(ARGS), { wrapper });
    const err = await result.current
      .signOff({ decision: 'approve', password: 'wrong' })
      .catch((e) => e);
    expect(err).toBeInstanceOf(SignoffError);
    expect(err).toMatchObject({ kind: 'bad_password' });
  });
});
