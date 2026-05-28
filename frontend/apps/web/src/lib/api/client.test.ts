import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { apiFetch } from './client';
import { ApiError } from './errors';
import * as authBus from './authBus';

function problemResponse(status: number, code: string): Response {
  return new Response(JSON.stringify({ code, title: code, status }), {
    status,
    headers: { 'Content-Type': 'application/problem+json' },
  });
}

describe('assertApiResponse — 401 discrimination', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('treats AUTH_UNAUTHORIZED as session-expired: dispatches authExpired and throws authn.expired', async () => {
    const dispatch = vi.spyOn(authBus, 'dispatchAuthExpired').mockImplementation(() => {});
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(problemResponse(401, 'AUTH_UNAUTHORIZED'));

    await expect(apiFetch('/me')).rejects.toMatchObject({ code: 'authn.expired', status: 401 });
    expect(dispatch).toHaveBeenCalledOnce();
  });

  it('preserves a domain 401 code (AUTH_INVALID_CREDENTIALS) and does NOT dispatch authExpired', async () => {
    const dispatch = vi.spyOn(authBus, 'dispatchAuthExpired').mockImplementation(() => {});
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(problemResponse(401, 'AUTH_INVALID_CREDENTIALS'));

    const err: unknown = await apiFetch('/auth/change-password', { method: 'POST' }).catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).code).toBe('AUTH_INVALID_CREDENTIALS');
    expect((err as ApiError).status).toBe(401);
    expect(dispatch).not.toHaveBeenCalled();
  });

  it('treats a bodyless 401 as session-expired (conservative default)', async () => {
    const dispatch = vi.spyOn(authBus, 'dispatchAuthExpired').mockImplementation(() => {});
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(null, { status: 401 }));

    await expect(apiFetch('/me')).rejects.toMatchObject({ code: 'authn.expired', status: 401 });
    expect(dispatch).toHaveBeenCalledOnce();
  });
});
