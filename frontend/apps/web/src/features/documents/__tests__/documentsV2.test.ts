import { describe, it, expect, vi, afterEach } from 'vitest';
import { finalizeDocument, listDocuments } from '../api/documentsV2';
import { ApiError } from '../../../lib/api';

describe('documentsV2 with apiFetch', () => {
  afterEach(() => vi.restoreAllMocks());

  it('listDocuments returns array on 200', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify([{ id: '1', name: 'Doc', status: 'draft', template_version_id: 'tv1', updated_at: '2026-01-01' }]),
        { status: 200 },
      ),
    );
    const docs = await listDocuments();
    expect(docs).toHaveLength(1);
    expect(docs[0].id).toBe('1');
  });

  it('finalizeDocument throws ApiError on 404 with parsed code', async () => {
    vi.spyOn(global, 'fetch').mockImplementation(() =>
      Promise.resolve(
        new Response(
          JSON.stringify({ error: { code: 'not_found.route', message: 'no route' } }),
          { status: 404 },
        ),
      ),
    );
    await expect(finalizeDocument('doc-1')).rejects.toBeInstanceOf(ApiError);
    await expect(finalizeDocument('doc-1')).rejects.toMatchObject({ code: 'not_found.route', status: 404 });
  });
});
