import { beforeEach, describe, expect, it, vi } from 'vitest';

import { etagCache } from '../etagCache';
import { ApprovalError } from '../mutationClient';
import { listRoutes, seedRouteEtag } from '../routeAdminApi';

vi.mock('../../../../lib/api', async () => {
  const actual = await vi.importActual<Record<string, unknown>>('../../../../lib/api');
  return {
    ...actual,
    dispatchAuthExpired: vi.fn(),
  };
});

const originalFetch = global.fetch;

beforeEach(() => {
  etagCache.clear();
  vi.restoreAllMocks();
  global.fetch = originalFetch;
});

describe('seedRouteEtag', () => {
  it('writes the canonical `"v<N>"` token the backend parseIfMatch accepts', () => {
    seedRouteEtag('route-1', 5);
    expect(etagCache.get('route-1')).toBe('"v5"');
  });

  it('is monotonic — a stale list refetch must not clobber a fresher ETag', () => {
    seedRouteEtag('route-1', 6);
    seedRouteEtag('route-1', 5);
    expect(etagCache.get('route-1')).toBe('"v6"');
  });

  it('overwrites when the new version is strictly greater', () => {
    seedRouteEtag('route-1', 4);
    seedRouteEtag('route-1', 7);
    expect(etagCache.get('route-1')).toBe('"v7"');
  });

  it('ignores non-positive or non-finite versions', () => {
    seedRouteEtag('route-1', 0);
    seedRouteEtag('route-1', -1);
    seedRouteEtag('route-1', Number.NaN);
    expect(etagCache.get('route-1')).toBeUndefined();
  });
});

describe('listRoutes', () => {
  it('parses problem+json and throws ApprovalError with the backend code', async () => {
    const problem = {
      type: 'about:blank',
      title: 'forbidden',
      status: 403,
      code: 'authz.capability_denied',
    };
    global.fetch = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(problem), {
        status: 403,
        headers: { 'Content-Type': 'application/problem+json' },
      }),
    ) as typeof fetch;

    await expect(listRoutes()).rejects.toMatchObject({
      code: 'authz.capability_denied',
      status: 403,
    });
    await expect(listRoutes()).rejects.toBeInstanceOf(ApprovalError);
  });

  it('returns the parsed body on 2xx', async () => {
    const body = { routes: [], total: 0 };
    global.fetch = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(body), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    ) as typeof fetch;

    await expect(listRoutes()).resolves.toEqual(body);
  });

  it('throws ApprovalError with `authn.expired` on bare 401 (no problem body)', async () => {
    global.fetch = vi.fn().mockResolvedValue(
      new Response(null, { status: 401 }),
    ) as typeof fetch;

    await expect(listRoutes()).rejects.toMatchObject({
      code: 'authn.expired',
      status: 401,
    });
  });
});
