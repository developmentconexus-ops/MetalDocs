import { afterEach, describe, expect, it, vi } from 'vitest';
import { submit } from '../approvalApi';

// FE-03: submit's request/response types moved from a hand-rolled
// SubmitRequest/SubmitResponse pair to the generated
// components['schemas']['SubmitDocumentRequest']/['SubmitDocumentResponse']
// (operations['submitDocumentForApproval'] targets this exact route). This
// test confirms the real wire shape — { instance_id, was_replay, etag } —
// still round-trips through submit() unchanged.
describe('approvalApi.submit', () => {
  afterEach(() => vi.restoreAllMocks());

  it('posts route_id/content_hash and returns the generated SubmitDocumentResponse shape', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({ instance_id: 'inst-1', was_replay: false, etag: '"v2"' }),
        { status: 201, headers: { 'Content-Type': 'application/json' } },
      ),
    );

    const result = await submit('doc-1', { route_id: 'route-1', content_hash: 'abc123' });

    expect(result).toEqual({ instance_id: 'inst-1', was_replay: false, etag: '"v2"' });
    const fetchMock = vi.mocked(fetch);
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit | undefined];
    expect(url).toBe('/api/v1/documents/doc-1/submit');
    expect(JSON.parse(String(init?.body))).toEqual({ route_id: 'route-1', content_hash: 'abc123' });
  });
});
