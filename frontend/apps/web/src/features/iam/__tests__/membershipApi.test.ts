import { describe, it, expect, vi, afterEach } from 'vitest';
import { fetchMemberships, grantMembership } from '../membershipApi';
import { ApiError } from '../../../lib/api';

describe('iam/membershipApi with apiFetch', () => {
  afterEach(() => vi.restoreAllMocks());

  it('fetchMemberships returns array on 200', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(JSON.stringify([{ userId: 'u-1', areaCode: 'A', role: 'viewer' }]), { status: 200 }),
    );
    const memberships = await fetchMemberships('u-1');
    expect(memberships).toHaveLength(1);
    expect(memberships[0].userId).toBe('u-1');
  });

  it('grantMembership throws ApiError on 403', async () => {
    vi.spyOn(global, 'fetch').mockImplementation(() =>
      Promise.resolve(
        new Response(JSON.stringify({ error: { code: 'authz.capability_denied' } }), { status: 403 }),
      ),
    );
    await expect(grantMembership({ userId: 'u-1', areaCode: 'A', role: 'viewer' })).rejects.toBeInstanceOf(ApiError);
  });
});
