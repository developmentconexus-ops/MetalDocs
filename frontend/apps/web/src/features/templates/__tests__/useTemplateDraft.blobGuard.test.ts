/**
 * Tests for F-FE3: useTemplateDraft blob-404 guard by version status.
 *
 * Under ADR 0088 every template version is materialized with content (and a
 * content_hash) at creation time — there is no "blank canvas, no object yet"
 * state. A missing blob is therefore a real failure for every version status:
 *
 * draft        + blob 404 → docxError set (retry banner)
 * under_review + blob 404 → docxError set (retry banner)
 * approved     + blob 404 → docxError set
 * published    + blob 404 → docxError set
 * any status   + blob 5xx → docxError set
 *
 * Mirrors DocumentEditorPage.tsx:122-129 (every blob failure incl. 404 becomes
 * editorLoadError, no per-status carve-out).
 */
import { renderHook, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

const getTemplateMock = vi.fn();
const getVersionMock = vi.fn();
const getDocxURLMock = vi.fn();

vi.mock('../api/templates', () => ({
  getTemplate: (...a: unknown[]) => getTemplateMock(...a),
  getVersion: (...a: unknown[]) => getVersionMock(...a),
  getDocxURL: (...a: unknown[]) => getDocxURLMock(...a),
}));

import { useTemplateDraft } from '../hooks/useTemplateDraft';

const BLOB_URL = 'https://s3.example.com/blob';
const TEMPLATE = { template: { template_id: 'tpl-1', name: 'T' } };

function makeVersion(status: string) {
  return {
    id: 'ver-1',
    template_id: 'tpl-1',
    version_number: 1,
    revision_number: 0,
    status,
    docx_storage_key: 'key/blob.docx',
    content_hash: null,
  };
}

afterEach(() => {
  vi.restoreAllMocks();
  getTemplateMock.mockReset();
  getVersionMock.mockReset();
  getDocxURLMock.mockReset();
});

function stubFetch(status: number, body = '') {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(body, { status })));
}

describe('useTemplateDraft — blob 404 guard (F-FE3)', () => {
  it.each(['draft', 'under_review', 'approved', 'published'] as const)(
    '%s + 404 → docxError set (retry banner)',
    async (status) => {
      getTemplateMock.mockResolvedValue(TEMPLATE);
      getVersionMock.mockResolvedValue(makeVersion(status));
      getDocxURLMock.mockResolvedValue(BLOB_URL);
      stubFetch(404);

      const { result } = renderHook(() => useTemplateDraft('tpl-1', 1));
      await waitFor(() => expect(result.current.loading).toBe(false));

      expect(result.current.docxError).toMatch(/404/);
      expect(result.current.docxBytes).toBeNull();
      expect(result.current.error).toBeNull();
    },
  );

  it('draft + 5xx → docxError set (real failure)', async () => {
    getTemplateMock.mockResolvedValue(TEMPLATE);
    getVersionMock.mockResolvedValue(makeVersion('draft'));
    getDocxURLMock.mockResolvedValue(BLOB_URL);
    stubFetch(503);

    const { result } = renderHook(() => useTemplateDraft('tpl-1', 1));
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.docxError).toMatch(/503/);
    expect(result.current.docxBytes).toBeNull();
  });

  it('draft + ok blob → docxBytes set, no error', async () => {
    getTemplateMock.mockResolvedValue(TEMPLATE);
    getVersionMock.mockResolvedValue(makeVersion('draft'));
    getDocxURLMock.mockResolvedValue(BLOB_URL);
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(new Uint8Array([1, 2, 3]).buffer, { status: 200 }),
      ),
    );

    const { result } = renderHook(() => useTemplateDraft('tpl-1', 1));
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.docxError).toBeNull();
    expect(result.current.docxBytes).toBeInstanceOf(ArrayBuffer);
  });
});
