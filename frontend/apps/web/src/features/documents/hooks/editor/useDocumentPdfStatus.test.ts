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
    expect(result.current.stalled).toBe(false);
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

  it('stays pending after the poll ceiling when the server never reports a terminal status', async () => {
    const mockFetch = vi.fn().mockImplementation(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ pdf_status: 'pending' }),
      } as Response),
    );
    global.fetch = mockFetch;

    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    // runAllTimersAsync fires every scheduled timer until poll() stops
    // scheduling new ones, which happens once the ceiling (Date.now() -
    // startedAt > TIMEOUT_MS) is reached.
    await act(async () => {
      await vi.runAllTimersAsync();
    });

    expect(result.current.status).toBe('pending');
    expect(result.current.stalled).toBe(true);
    const callsAtCeiling = mockFetch.mock.calls.length;

    // Advance well past where more polls would have fired if the hook kept
    // going — call count must not grow, proving polling actually stopped.
    await act(async () => {
      await vi.advanceTimersByTimeAsync(120_000);
    });
    expect(mockFetch.mock.calls.length).toBe(callsAtCeiling);
    expect(result.current.status).toBe('pending');
  });

  it('surfaces failed only when the server reports it', async () => {
    global.fetch = makeFetch(
      { pdf_status: 'pending' },
      { pdf_status: 'pending' },
      { pdf_status: 'failed' },
    );

    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    await act(async () => {
      await vi.runAllTimersAsync();
    });

    expect(result.current.status).toBe('failed');
  });

  it('retry restarts polling and surfaces ready once the server reports it', async () => {
    let phase: 'pending' | 'ready' = 'pending';
    const mockFetch = vi.fn().mockImplementation(() => {
      const body = phase === 'ready'
        ? { pdf_status: 'ready', pdf_url: 'u' }
        : { pdf_status: 'pending' };
      return Promise.resolve({ ok: true, json: () => Promise.resolve(body) } as Response);
    });
    global.fetch = mockFetch;

    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    // Exhaust the poll ceiling while the server stays pending.
    await act(async () => { await vi.runAllTimersAsync(); });
    expect(result.current.status).toBe('pending');
    expect(result.current.stalled).toBe(true);

    // Switch to ready phase and retry.
    phase = 'ready';
    act(() => { result.current.retry(); });
    await act(async () => { await vi.runAllTimersAsync(); });
    expect(result.current.status).toBe('ready');
    expect(result.current.stalled).toBe(false);
  });

  it('retry clears stalled and resumes polling even while still pending', async () => {
    const mockFetch = vi.fn().mockImplementation(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ pdf_status: 'pending' }),
      } as Response),
    );
    global.fetch = mockFetch;

    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    await act(async () => { await vi.runAllTimersAsync(); });
    expect(result.current.stalled).toBe(true);
    const callsAtCeiling = mockFetch.mock.calls.length;

    act(() => { result.current.retry(); });
    // First poll of the new cycle fires immediately; the ceiling has not been
    // reached again yet, so stalled must be cleared and polling live again.
    await act(async () => { await vi.advanceTimersByTimeAsync(0); });
    expect(result.current.stalled).toBe(false);
    expect(mockFetch.mock.calls.length).toBeGreaterThan(callsAtCeiling);
  });

  it('does not stall when the server reaches ready before the ceiling', async () => {
    global.fetch = makeFetch(
      { pdf_status: 'pending' },
      { pdf_status: 'ready', pdf_url: 'u' },
    );

    const { result } = renderHook(() => useDocumentPdfStatus('doc-1', true));
    await act(async () => { await vi.runAllTimersAsync(); });

    expect(result.current.status).toBe('ready');
    expect(result.current.stalled).toBe(false);
  });
});
