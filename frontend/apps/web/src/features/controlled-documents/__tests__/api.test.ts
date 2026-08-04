import { describe, it, expect, vi, afterEach } from 'vitest';
import { fetchControlledDocuments, fetchActiveDocumentInstance, fetchControlledDocument } from '../api/controlledDocuments';
import { ApiError } from '../../../lib/api';

describe('registry/api with apiFetch', () => {
  afterEach(() => vi.restoreAllMocks());

  it('fetchControlledDocuments returns items array on 200', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ items: [{ id: 'cd-1', name: 'Doc' }] }), { status: 200 }),
    );
    const docs = await fetchControlledDocuments();
    expect(docs).toHaveLength(1);
    expect(docs[0].id).toBe('cd-1');
  });

  it('fetchControlledDocuments tolerates the real envelope shape including page cursor', async () => {
    // FE-03: response type is now derived from operations['listControlledDocuments'],
    // which declares { items, page:{ next_cursor, has_more } }. This fixture
    // includes `page` (as the real API always does) to confirm the generated
    // type accepts the full envelope rather than the narrower hand-rolled
    // `{ items }` type that previously hid this field from tsc.
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({
        items: [{ id: 'cd-1', name: 'Doc' }],
        page: { next_cursor: 'abc123', has_more: true },
      }), { status: 200 }),
    );
    const docs = await fetchControlledDocuments();
    expect(docs).toHaveLength(1);
    expect(docs[0].id).toBe('cd-1');
  });

  it('fetchActiveDocumentInstance returns null on 404', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: { code: 'notfound.document', message: 'not found' } }), { status: 404 }),
    );
    const result = await fetchActiveDocumentInstance('cd-1');
    expect(result).toBeNull();
  });

  it('fetchActiveDocumentInstance throws ApiError on non-404 error', async () => {
    vi.spyOn(global, 'fetch').mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ error: { code: 'permission.capability_denied' } }), { status: 403 })),
    );
    await expect(fetchActiveDocumentInstance('cd-1')).rejects.toBeInstanceOf(ApiError);
  });

  it('reads controlled document detail including visibility', async () => {
    // FE-03: fixture uses the real wire format (snake_case, per
    // components["schemas"]["ControlledDocument"]/["ControlledDocumentVisibility"]
    // in lib/api-types). The previous camelCase fixture (tenantId, areaCodes,
    // userIds) passed only because fetchControlledDocument did an untyped
    // cast with no runtime validation — it never exercised the real shape.
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({
        id: 'cd-1',
        tenant_id: 'tenant-1',
        profile_code: 'POP',
        process_area_code: 'RH',
        code: 'POP-RH-001',
        title: 'Procedimento RH',
        owner_user_id: 'user-1',
        status: 'active',
        visibility: { scope: 'restricted', area_codes: ['RH'], user_ids: [] },
        created_at: '2026-05-18T12:00:00Z',
        updated_at: '2026-05-18T12:00:00Z',
      }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
    );

    const result = await fetchControlledDocument('cd-1');
    expect(result.visibility.scope).toBe('restricted');
    expect(result.visibility.area_codes).toEqual(['RH']);
  });
});
