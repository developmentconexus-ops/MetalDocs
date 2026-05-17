import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createTemplate, importTemplateDocx } from '../templates';

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
    vi.stubGlobal('crypto', {
      subtle: {
        digest: vi.fn().mockResolvedValue(
          new Uint8Array([
            0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
            0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
            0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
            0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef,
          ]).buffer,
        ),
      },
    });
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

  it('imports a wizard docx through presign upload and commit', async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock
      .mockResolvedValueOnce(
        makeResponse({
          data: {
            upload_url: 'https://minio/upload/docx',
            storage_key: 'templates/tpl_123/versions/1.docx',
            expires_at: '2026-01-01T00:10:00Z',
          },
        }),
      )
      .mockResolvedValueOnce(makeResponse({}, 200))
      .mockResolvedValueOnce(
        makeResponse({
          data: {
            version: {
              id: 'ver_123',
              template_id: 'tpl_123',
              version_number: 1,
              status: 'draft',
              docx_storage_key: 'templates/tpl_123/versions/1.docx',
              content_hash: 'expected-hash',
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

    const file = new File(['docx-bytes'], 'modelo.docx', {
      type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    });
    Object.defineProperty(file, 'arrayBuffer', {
      value: vi.fn().mockResolvedValue(new TextEncoder().encode('docx-bytes').buffer),
    });

    await importTemplateDocx('tpl_123', 1, file);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      '/api/v1/templates/tpl_123/versions/1/autosave/presign',
      { method: 'POST' },
    );
    expect(fetchMock.mock.calls[1][0]).toBe('https://minio/upload/docx');
    expect(fetchMock.mock.calls[1][1]).toMatchObject({
      method: 'PUT',
      headers: {
        'content-type': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      },
    });
    expect(fetchMock.mock.calls[2][0]).toBe('/api/v1/templates/tpl_123/versions/1/autosave/commit');
    expect(JSON.parse(String(fetchMock.mock.calls[2][1]?.body))).toEqual({
      expected_content_hash: '1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
    });
  });
});
