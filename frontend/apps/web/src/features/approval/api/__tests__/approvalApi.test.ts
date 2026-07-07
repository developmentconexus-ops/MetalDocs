import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  createDelegation,
  listInbox,
  reviewVerdict,
  revokeDelegation,
  submit,
} from '../approvalApi';

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

// F2 (M2c): reviewVerdict posts the ReviewVerdictRequest with the caller-supplied
// ETag as If-Match plus an auto-generated Idempotency-Key, and returns the
// generated ReviewVerdictResponse shape unchanged.
describe('approvalApi.reviewVerdict', () => {
  afterEach(() => vi.restoreAllMocks());

  it('posts verdict/comment with If-Match and Idempotency-Key headers', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({ verdict_id: 'verdict-1', was_replay: false, outcome: 'advanced' }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    );

    const result = await reviewVerdict({
      instanceId: 'inst-1',
      stageId: 'stage-1',
      etag: '"v3"',
      body: { verdict: 'ready', comment: 'looks good' },
    });

    expect(result).toEqual({ verdict_id: 'verdict-1', was_replay: false, outcome: 'advanced' });
    const fetchMock = vi.mocked(fetch);
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit | undefined];
    expect(url).toBe('/api/v1/approval/instances/inst-1/stages/stage-1/review-verdict');
    expect(init?.method).toBe('POST');
    expect(JSON.parse(String(init?.body))).toEqual({ verdict: 'ready', comment: 'looks good' });

    const headers = new Headers(init?.headers);
    expect(headers.get('If-Match')).toBe('"v3"');
    expect(headers.get('Idempotency-Key')).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
    );
  });
});

// F2 (M2c): listInbox forwards the three new W4/P2/P3/P8 filter params
// (stage_kind/due_before/scope) into the query string alongside the existing
// area_code/limit/offset, but only when defined.
describe('approvalApi.listInbox', () => {
  afterEach(() => vi.restoreAllMocks());

  it('forwards stage_kind/due_before/scope query params when provided', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ items: [], total: 0 }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );

    await listInbox({
      area_code: 'AREA-1',
      stage_kind: 'review',
      due_before: '2026-07-08T00:00:00Z',
      scope: 'oversee',
    });

    const fetchMock = vi.mocked(fetch);
    const [url] = fetchMock.mock.calls[0] as [string, RequestInit | undefined];
    const parsed = new URL(url, 'http://localhost');
    expect(parsed.pathname).toBe('/api/v1/approval/inbox');
    expect(parsed.searchParams.get('area_code')).toBe('AREA-1');
    expect(parsed.searchParams.get('stage_kind')).toBe('review');
    expect(parsed.searchParams.get('due_before')).toBe('2026-07-08T00:00:00Z');
    expect(parsed.searchParams.get('scope')).toBe('oversee');
  });
});

// F9 (M2c): createDelegation posts CreateApprovalDelegationRequest with an
// auto-generated Idempotency-Key and returns the generated ApprovalDelegation
// record unchanged.
describe('approvalApi.createDelegation', () => {
  afterEach(() => vi.restoreAllMocks());

  it('posts delegate_id/starts_at/ends_at/reason and returns the ApprovalDelegation shape', async () => {
    const delegation = {
      id: 'deleg-1',
      tenant_id: 'tenant-1',
      delegator_id: 'user-1',
      delegate_id: 'user-2',
      starts_at: '2026-07-08T00:00:00Z',
      ends_at: '2026-07-15T00:00:00Z',
      reason: 'vacation',
      created_by: 'user-1',
      created_at: '2026-07-07T00:00:00Z',
    };
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(JSON.stringify(delegation), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      }),
    );

    const result = await createDelegation({
      delegate_id: 'user-2',
      starts_at: '2026-07-08T00:00:00Z',
      ends_at: '2026-07-15T00:00:00Z',
      reason: 'vacation',
    });

    expect(result).toEqual(delegation);
    const fetchMock = vi.mocked(fetch);
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit | undefined];
    expect(url).toBe('/api/v1/approval/delegations');
    expect(init?.method).toBe('POST');
    expect(JSON.parse(String(init?.body))).toEqual({
      delegate_id: 'user-2',
      starts_at: '2026-07-08T00:00:00Z',
      ends_at: '2026-07-15T00:00:00Z',
      reason: 'vacation',
    });
    const headers = new Headers(init?.headers);
    expect(headers.get('Idempotency-Key')).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
    );
  });
});

// F9 (M2c): revokeDelegation issues a DELETE with no body and resolves cleanly
// on the 204 No Content the endpoint returns (no JSON body to parse).
describe('approvalApi.revokeDelegation', () => {
  afterEach(() => vi.restoreAllMocks());

  it('deletes by id and resolves on 204 with no body', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(new Response(null, { status: 204 }));

    await expect(revokeDelegation('deleg-1')).resolves.toBeUndefined();

    const fetchMock = vi.mocked(fetch);
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit | undefined];
    expect(url).toBe('/api/v1/approval/delegations/deleg-1');
    expect(init?.method).toBe('DELETE');
  });
});
