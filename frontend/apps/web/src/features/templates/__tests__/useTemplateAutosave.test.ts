/**
 * Tests for F-FE1 + F-FE7: useTemplateAutosave upload-failure handling.
 *
 * F-FE1: flush() must return false (not throw) on upload failure and NOT proceed
 *        to commitAutosave.
 * F-FE7: importDocx() must throw when the S3 PUT returns !ok.
 */
import { renderHook, act } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useTemplateAutosave } from '../hooks/useTemplateAutosave';

const presignMock = vi.fn();
const commitMock = vi.fn();

vi.mock('../api/templates', () => ({
  presignAutosave: (...args: unknown[]) => presignMock(...args),
  commitAutosave: (...args: unknown[]) => commitMock(...args),
}));

const UPLOAD_URL = 'https://s3.example.com/put';

function makeOkFetch() {
  return vi.fn().mockResolvedValue(
    new Response(null, { status: 200 }),
  );
}
function makeFailFetch(status = 503) {
  return vi.fn().mockResolvedValue(
    new Response('service unavailable', { status }),
  );
}

beforeEach(() => {
  presignMock.mockResolvedValue({ upload_url: UPLOAD_URL, storage_key: 'k' });
  commitMock.mockResolvedValue({});
  // crypto.subtle stub for jsdom, matching useDocumentAutosave.test.ts and
  // templates.create.test.ts. The ArrayBuffer these tests construct belongs to
  // the jsdom realm, and Node 22's webcrypto digest does a strict instanceof
  // check against its own realm's ArrayBuffer — so the real implementation
  // throws ERR_INVALID_ARG_TYPE on Node 22 while passing on Node 26. That is a
  // cross-realm test artifact, not a product defect: in a browser the buffer is
  // same-realm. These tests cover upload-failure handling (F-FE1/F-FE7), not
  // hashing, so stubbing the digest costs no coverage.
  vi.spyOn(crypto.subtle, 'digest').mockResolvedValue(new Uint8Array(32).buffer);
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('useTemplateAutosave — dirty status (F-FE4)', () => {
  it('sets status to dirty immediately when queueDocx is called', async () => {
    vi.stubGlobal('fetch', makeOkFetch());
    vi.useFakeTimers();
    const { result } = renderHook(() => useTemplateAutosave('tpl-1', 1));

    expect(result.current.status).toBe('idle');

    await act(async () => {
      result.current.queueDocx(new ArrayBuffer(4));
    });

    expect(result.current.status).toBe('dirty');
    vi.useRealTimers();
  });
});

describe('useTemplateAutosave — flush (F-FE1)', () => {
  it('returns true and calls commitAutosave when upload succeeds', async () => {
    vi.stubGlobal('fetch', makeOkFetch());
    const { result } = renderHook(() => useTemplateAutosave('tpl-1', 1));

    await act(async () => {
      result.current.queueDocx(new ArrayBuffer(4));
    });

    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.flush();
    });

    expect(ok).toBe(true);
    expect(commitMock).toHaveBeenCalledTimes(1);
    expect(result.current.status).toBe('saved');
  });

  it('returns false and does NOT call commitAutosave when upload returns !ok', async () => {
    vi.stubGlobal('fetch', makeFailFetch(503));
    const { result } = renderHook(() => useTemplateAutosave('tpl-1', 1));

    await act(async () => {
      result.current.queueDocx(new ArrayBuffer(4));
    });

    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.flush();
    });

    expect(ok).toBe(false);
    expect(commitMock).not.toHaveBeenCalled();
    expect(result.current.status).toBe('error');
  });

  it('returns true immediately (no-op) when there is no pending docx', async () => {
    const fetchSpy = vi.fn().mockResolvedValue(new Response(null, { status: 200 }));
    vi.stubGlobal('fetch', fetchSpy);
    const { result } = renderHook(() => useTemplateAutosave('tpl-1', 1));

    let ok: boolean | undefined;
    await act(async () => {
      ok = await result.current.flush();
    });

    expect(ok).toBe(true);
    expect(fetchSpy).not.toHaveBeenCalled();
    expect(commitMock).not.toHaveBeenCalled();
  });
});

describe('useTemplateAutosave — importDocx (F-FE7)', () => {
  it('resolves without error when upload succeeds', async () => {
    vi.stubGlobal('fetch', makeOkFetch());
    const { result } = renderHook(() => useTemplateAutosave('tpl-1', 1));

    await act(async () => {
      await result.current.importDocx(new ArrayBuffer(4));
    });

    expect(commitMock).toHaveBeenCalledTimes(1);
    expect(result.current.status).toBe('saved');
  });

  it('throws and sets status=error when upload returns !ok', async () => {
    vi.stubGlobal('fetch', makeFailFetch(400));
    const { result } = renderHook(() => useTemplateAutosave('tpl-1', 1));

    let thrown: Error | undefined;
    await act(async () => {
      try {
        await result.current.importDocx(new ArrayBuffer(4));
      } catch (e) {
        thrown = e as Error;
      }
    });

    expect(thrown).toBeInstanceOf(Error);
    expect(thrown?.message).toMatch(/400/);
    expect(commitMock).not.toHaveBeenCalled();
    expect(result.current.status).toBe('error');
  });
});
