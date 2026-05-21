import { describe, it, expect, vi, afterEach } from 'vitest';
import {
  finalizeDocument,
  getApprovalInstance,
  getDocument,
  listComments,
  listDocuments,
} from '../api/documents';
import { ApiError } from '../../../lib/api';
import type { components } from '../../../lib/api-types';

type ApprovalInstanceContract = components['schemas']['ApprovalInstanceByDocumentResponse'];

describe('documents with apiFetch', () => {
  afterEach(() => vi.restoreAllMocks());

  it('listDocuments returns array on 200', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          items: [
            {
              ID: '1',
              TenantID: 'tenant-1',
              TemplateVersionID: 'tv1',
              Name: 'Doc',
              Status: 'draft',
              FormDataJSON: {},
              CurrentRevisionID: 'rev-1',
              RevisionVersion: 1,
              ActiveSessionID: '',
              CreatedAt: '2026-01-01T00:00:00Z',
              UpdatedAt: '2026-01-01T00:00:00Z',
              CreatedBy: 'user-1',
              Code: 'DOC-1',
            },
          ],
          page: 1,
          pageSize: 20,
          total: 1,
        }),
        { status: 200 },
      ),
    );
    const docs = await listDocuments();
    expect(docs).toHaveLength(1);
    expect(docs[0].ID).toBe('1');
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
    await expect(finalizeDocument('doc-1', { revisionTitle: 'Ajuste operacional' })).rejects.toBeInstanceOf(ApiError);
    await expect(finalizeDocument('doc-1', { revisionTitle: 'Ajuste operacional' })).rejects.toMatchObject({ code: 'not_found.route', status: 404 });
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

  it('listComments returns typed comment payloads', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify([
          {
            id: '11111111-1111-4111-8111-111111111111',
            library_comment_id: 42,
            parent_library_id: null,
            author: 'Alice Doe',
            author_id: 'iam-alice',
            content: [{ type: 'paragraph', children: [{ text: 'hello' }] }],
            done: false,
            created_at: '2026-05-18T10:00:00Z',
            updated_at: '2026-05-18T10:00:00Z',
            resolved_at: null,
          },
        ]),
        { status: 200, headers: { 'content-type': 'application/json' } },
      ),
    );

    const comments = await listComments('doc-1');

    expect(comments).toHaveLength(1);
    expect(comments[0].library_comment_id).toBe(42);
    expect(comments[0].content[0]).toMatchObject({ type: 'paragraph' });
  });

  it('finalizeDocument sends Idempotency-Key and returns instanceId on 201', async () => {
    vi.spyOn(globalThis.crypto, 'randomUUID').mockReturnValue('11111111-1111-4111-8111-111111111111');
    const fetchSpy = vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({ instanceId: 'inst_1' }),
        { status: 201, headers: { 'content-type': 'application/json' } },
      ),
    );

    const result = await finalizeDocument('doc-1', { revisionTitle: 'Ajuste operacional' });

    expect(result).toEqual({ instanceId: 'inst_1' });
    const [, init] = fetchSpy.mock.calls[0] ?? [];
    const headers = init?.headers as Record<string, string> | undefined;
    expect(headers).toMatchObject({ 'Idempotency-Key': '11111111-1111-4111-8111-111111111111' });
    expect(JSON.parse(String(init?.body))).toEqual({ revisionTitle: 'Ajuste operacional' });
  });

  it('reads approval-instance with runtime-aligned statuses and signoff payload', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          id: 'inst-1',
          document_id: 'doc-1',
          route_id: 'route-1',
          tenant_id: 'tenant-1',
          status: 'in_progress',
          submitted_by: 'user-1',
          submitted_at: '2026-05-18T12:00:00Z',
          stages: [
            {
              id: 'stage-1',
              stage_index: 1,
              label: 'Qualidade',
              status: 'active',
              signoffs: [],
            },
            {
              id: 'stage-2',
              stage_index: 2,
              label: 'Diretoria',
              status: 'pending',
              signoffs: [
                {
                  id: 'signoff-1',
                  actor_user_id: 'Maria Souza',
                  decision: 'approve',
                  signature_method: 'password_reauth',
                  signed_at: '2026-05-18T12:30:00Z',
                },
              ],
            },
          ],
          etag: '"v1"',
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    );

    const result = await getApprovalInstance('doc-1');
    const typedResult: ApprovalInstanceContract = result;

    expect(typedResult.status).toBe('in_progress');
    expect(typedResult.stages[0]?.status).toBe('active');
    expect(typedResult.stages[1]?.signoffs[0]?.actor_user_id).toBe('Maria Souza');
  });
});
