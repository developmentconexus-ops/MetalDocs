import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { changePassword, login, logout, me } from '../auth';

function makeResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
    headers: new Headers(),
  } as unknown as Response;
}

// FE-20: login/me/changePassword responses are now typed off the generated
// operations (AuthLoginResponse / CurrentUser / ChangePasswordResponse)
// instead of a hand-rolled WireCurrentUser inline shape. Verified against
// internal/modules/auth/delivery/http/handler.go — handleLogin/handleMe/
// handleChangePassword all write a flat body (no `{ data }` envelope).
const wireUser = {
  user_id: 'user-1',
  tenant_id: 'tenant-1',
  tenant_name: 'Acme',
  username: 'jdoe',
  email: 'jdoe@example.com',
  display_name: 'Jane Doe',
  must_change_password: false,
  roles: ['editor', 'signer'],
  capabilities: ['documents.read', 'documents.write'],
};

describe('auth api', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('login normalizes the flat AuthLoginResponse body, preserving roles beyond the generated 5-value enum', async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(
      makeResponse({ user: wireUser, expires_at: '2026-07-03T00:00:00Z' }),
    );

    const result = await login({ identifier: 'jdoe', password: 'secret' });

    expect(result.expires_at).toBe('2026-07-03T00:00:00Z');
    expect(result.user).toEqual({
      userId: 'user-1',
      tenantId: 'tenant-1',
      tenantName: 'Acme',
      username: 'jdoe',
      email: 'jdoe@example.com',
      displayName: 'Jane Doe',
      mustChangePassword: false,
      // 'signer' is outside the generated CurrentUser.roles 5-value union but
      // is a real runtime role (iamdomain.Role has 8) — normalizeRoles must
      // still pass it through via its own allowlist, not the generated type.
      roles: ['editor', 'signer'],
      capabilities: ['documents.read', 'documents.write'],
    });
  });

  it('me() normalizes a flat CurrentUser body with no envelope', async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(makeResponse(wireUser));

    const result = await me();

    expect(result.userId).toBe('user-1');
    expect(result.roles).toEqual(['editor', 'signer']);
    expect(result.capabilities).toEqual(['documents.read', 'documents.write']);
  });

  it('changePassword normalizes the flat ChangePasswordResponse body', async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(makeResponse({ changed: true, user: wireUser }));

    const result = await changePassword({ currentPassword: 'old', newPassword: 'new' });

    expect(result.changed).toBe(true);
    expect(result.user.userId).toBe('user-1');
    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit | undefined];
    expect(JSON.parse(String(init?.body))).toEqual({
      current_password: 'old',
      new_password: 'new',
    });
  });

  it('logout resolves without a body on 204', async () => {
    const fetchMock = vi.mocked(fetch);
    fetchMock.mockResolvedValueOnce(makeResponse(undefined, 204));

    await expect(logout()).resolves.toBeUndefined();
  });
});
