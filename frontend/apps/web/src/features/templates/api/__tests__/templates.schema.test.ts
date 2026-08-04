import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { putTemplateSchemas, SchemaSaveResponseShapeError, StaleLockVersionError } from '../templates';
import type { TemplateSchemas } from '../templates';

function makeResponse(body: unknown, status = 200, headers?: Record<string, string>): Response {
  const res = {
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
    headers: new Headers(headers),
  } as unknown as Response;
  (res as unknown as { clone: () => Response }).clone = () => res;
  return res;
}

const schemas: TemplateSchemas = {
  placeholders: [{ id: 'p1', label: 'Title', type: 'text' }],
  composition: null,
};

describe('templates.putTemplateSchemas', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  // FE-15: verified against internal/modules/templates/delivery/http/routes_schema.go
  // updateSchemas, which writes the inline { data: { version } } envelope declared
  // by operations['updateTemplateSchema'] (the shared TemplateVersionEnvelope
  // component this used to $ref was retired with the submit/review routes it was
  // written for — ADR 0082 slice S4 — the shape itself is unchanged).
  it('returns the bumped lock_version on a correctly enveloped 200', async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(
      makeResponse({
        data: {
          version: {
            id: 'ver_123',
            template_id: 'tpl_123',
            version_number: 1,
            status: 'draft',
            lock_version: 5,
          },
        },
      }),
    );

    const result = await putTemplateSchemas('tpl_123', 1, schemas, 4);

    expect(result).toEqual({ lockVersion: 5 });
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit | undefined];
    expect(url).toBe('/api/v1/templates/tpl_123/versions/1/schema');
    expect(init?.method).toBe('PUT');
    expect(JSON.parse(String(init?.body))).toMatchObject({ expected_lock_version: 4 });
  });

  // Regression: this fixture is the flat (non-enveloped) shape a stale
  // hand-rolled type would have accepted uncritically — before FE-15,
  // `body.data.version.lock_version` on a body with no `data` key is
  // `undefined`, and the old code silently returned
  // `{ lockVersion: expectedLockVersion + 1 }` (a fabricated CAS token that
  // masks real concurrent-write drift). The fixed code must fail loud instead.
  it('throws SchemaSaveResponseShapeError instead of guessing when the response has no envelope', async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(
      makeResponse({
        id: 'ver_123',
        template_id: 'tpl_123',
        version_number: 1,
        status: 'draft',
        lock_version: 5,
      }),
    );

    await expect(putTemplateSchemas('tpl_123', 1, schemas, 4)).rejects.toBeInstanceOf(
      SchemaSaveResponseShapeError,
    );
  });

  // Regression: same fail-loud requirement when the envelope is present but
  // lock_version itself is missing/non-numeric (e.g. version omitted or null).
  it('throws SchemaSaveResponseShapeError when data.version.lock_version is missing', async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(makeResponse({ data: { version: { id: 'ver_123' } } }));

    await expect(putTemplateSchemas('tpl_123', 1, schemas, 4)).rejects.toBeInstanceOf(
      SchemaSaveResponseShapeError,
    );
  });

  it('maps a 412 CONCURRENT_MODIFICATION problem+json to StaleLockVersionError', async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(
      makeResponse(
        {
          type: 'about:blank',
          title: 'Conflict',
          status: 412,
          code: 'conflict.concurrent_modification',
          detail: 'lock_version mismatch',
        },
        412,
        { 'content-type': 'application/problem+json' },
      ),
    );

    await expect(putTemplateSchemas('tpl_123', 1, schemas, 4)).rejects.toBeInstanceOf(
      StaleLockVersionError,
    );
  });
});
