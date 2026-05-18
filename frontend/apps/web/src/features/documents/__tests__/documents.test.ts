import { describe, it, expect, vi, afterEach } from 'vitest';
import { finalizeDocument, getDocument, listDocuments } from '../api/documents';
import { ApiError } from '../../../lib/api';

describe('documents with apiFetch', () => {
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

  it('getDocument returns typed detail payload with embedded FormDataJSON', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          ID: 'doc-1',
          Name: 'Doc',
          Status: 'draft',
          FormDataJSON: { foo: 'bar' },
          CurrentRevisionID: 'rev-1',
          RevisionVersion: 1,
          CreatedBy: 'user-1',
          Code: 'DOC-1',
        }),
        { status: 200, headers: { 'content-type': 'application/json' } },
      ),
    );

    const doc = await getDocument('doc-1');

    expect(doc.FormDataJSON).toEqual({ foo: 'bar' });
    expect(doc.CurrentRevisionID).toBe('rev-1');
  });

  it('finalizeDocument sends Idempotency-Key and returns instanceId on 201', async () => {
    vi.spyOn(globalThis.crypto, 'randomUUID').mockReturnValue('11111111-1111-4111-8111-111111111111');
    const fetchSpy = vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({ instanceId: 'inst_1' }),
        { status: 201, headers: { 'content-type': 'application/json' } },
      ),
    );

    const result = await finalizeDocument('doc-1');

    expect(result).toEqual({ instanceId: 'inst_1' });
    const [, init] = fetchSpy.mock.calls[0] ?? [];
    const headers = init?.headers as Record<string, string> | undefined;
    expect(headers).toMatchObject({ 'Idempotency-Key': '11111111-1111-4111-8111-111111111111' });
  });
});
