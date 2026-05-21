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

  it('fetchActiveDocumentInstance returns null on 404', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ error: { code: 'not_found.document', message: 'not found' } }), { status: 404 }),
    );
    const result = await fetchActiveDocumentInstance('cd-1');
    expect(result).toBeNull();
  });

  it('fetchActiveDocumentInstance throws ApiError on non-404 error', async () => {
    vi.spyOn(global, 'fetch').mockImplementation(() =>
      Promise.resolve(new Response(JSON.stringify({ error: { code: 'authz.capability_denied' } }), { status: 403 })),
    );
    await expect(fetchActiveDocumentInstance('cd-1')).rejects.toBeInstanceOf(ApiError);
  });

  it('reads controlled document detail including visibility', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({
        id: 'cd-1',
        tenantId: 'tenant-1',
        profileCode: 'POP',
        processAreaCode: 'RH',
        code: 'POP-RH-001',
        title: 'Procedimento RH',
        ownerUserId: 'user-1',
        status: 'active',
        visibility: { scope: 'restricted', areaCodes: ['RH'], userIds: [] },
        createdAt: '2026-05-18T12:00:00Z',
        updatedAt: '2026-05-18T12:00:00Z',
      }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
    );

    const result = await fetchControlledDocument('cd-1');
    expect(result.visibility.scope).toBe('restricted');
  });
});
