import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createTemplate } from '../templates';

function makeResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
    headers: new Headers(),
  } as unknown as Response;
}

describe('templates.createTemplate', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('sends payload with idempotency header and parses template/version response', async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(
      makeResponse({
        data: {
          template: {
            id: 'tpl_123',
            tenant_id: 'tenant_1',
            doc_type_code: 'MDDM',
            key: 'inspecao-recebimento',
            name: 'Inspecao de Recebimento',
            description: null,
            areas: [],
            visibility: 'public',
            specific_areas: [],
            latest_version: 1,
            published_version_id: null,
            created_by: 'user_1',
            created_at: '2026-01-01T00:00:00Z',
            archived_at: null,
          },
          version: {
            id: 'ver_123',
            template_id: 'tpl_123',
            version_number: 1,
            status: 'draft',
            docx_storage_key: null,
            content_hash: null,
            metadata_schema: null,
            placeholder_schema: null,
            author_id: 'user_1',
            pending_reviewer_role: null,
            pending_approver_role: null,
            reviewer_id: null,
            approver_id: null,
            submitted_at: null,
            reviewed_at: null,
            approved_at: null,
            published_at: null,
            obsoleted_at: null,
            created_at: '2026-01-01T00:00:00Z',
          },
        },
      }),
    );

    const result = await createTemplate({
      key: 'inspecao-recebimento',
      name: 'Inspecao de Recebimento',
      doc_type_code: 'MDDM',
      idempotencyKey: 'idem-123',
    });

    expect(fetchMock).toHaveBeenCalledOnce();
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit | undefined];
    expect(url).toBe('/api/v1/templates');
    expect(init?.method).toBe('POST');

    const headers = init?.headers as Record<string, string>;
    expect(headers['Idempotency-Key']).toBe('idem-123');
    expect(headers['Content-Type']).toBe('application/json');

    expect(JSON.parse(String(init?.body))).toEqual({
      key: 'inspecao-recebimento',
      name: 'Inspecao de Recebimento',
      doc_type_code: 'MDDM',
    });

    expect(result.template.id).toBe('tpl_123');
    expect(result.version.version_number).toBe(1);
  });
});
