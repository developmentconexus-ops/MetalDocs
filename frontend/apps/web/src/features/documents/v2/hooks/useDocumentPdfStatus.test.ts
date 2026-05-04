import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDocumentPdfStatus } from './useDocumentPdfStatus';

function makeFetch(...responses: Array<{ pdf_status: string; pdf_url?: string }>) {
  let call = 0;
  return vi.fn().mockImplementation(() => {
    const body = responses[Math.min(call++, responses.length - 1)];
    return Promise.resolve({
      ok: true,
      json: () => Promise.resolve(body),
    } as Response);
  });
}

beforeEach(() => vi.useFakeTimers());
afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks(); });

describe('useDocumentPdfStatus', () => {
  it('polls until ready and exposes URL', async () => {
    global.fetch = makeFetch(
      { pdf_status: 'pending' },
      { pdf_status: 'pending' },
      { pdf_status: 'ready', pdf_url: 'https://s3/x.pdf' },
    );

    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    // runAllTimersAsync fires all scheduled timers (including those created by
    // timer callbacks) and awaits their async bodies between each fire.
    await act(async () => {
      await vi.runAllTimersAsync();
    });

    expect(result.current.status).toBe('ready');
    expect(result.current.url).toBe('https://s3/x.pdf');
  });

  it('does not poll when disabled', async () => {
    const mockFetch = vi.fn();
    global.fetch = mockFetch;

    renderHook(() => useDocumentPdfStatus('doc-1', false));
    await act(async () => {
      await vi.runAllTimersAsync();
    });

    expect(mockFetch).not.toHaveBeenCalled();
  });

  it('marks failed after 60s timeout', async () => {
    global.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ pdf_status: 'pending' }),
      } as Response),
    );

    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    // runAllTimersAsync will keep firing until poll() stops scheduling timers,
    // which happens when Date.now() - startedAt > TIMEOUT_MS.
    await act(async () => {
      await vi.runAllTimersAsync();
    });

    expect(result.current.status).toBe('failed');
  });

  it('retry restarts polling', async () => {
    let phase: 'timeout' | 'ready' = 'timeout';
    global.fetch = vi.fn().mockImplementation(() => {
      const body = phase === 'ready'
        ? { pdf_status: 'ready', pdf_url: 'u' }
        : { pdf_status: 'pending' };
      return Promise.resolve({ ok: true, json: () => Promise.resolve(body) } as Response);
    });

    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    // Exhaust timeout phase → failed
    await act(async () => { await vi.runAllTimersAsync(); });
    expect(result.current.status).toBe('failed');

    // Switch to ready phase and retry
    phase = 'ready';
    act(() => { result.current.retry(); });
    await act(async () => { await vi.runAllTimersAsync(); });
    expect(result.current.status).toBe('ready');
  });
});
