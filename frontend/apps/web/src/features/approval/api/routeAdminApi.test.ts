import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { etagCache } from './etagCache';
import {
  createRoute,
  listRoutes,
  updateRoute,
  type ListRoutesResponse,
  type RouteSummary,
  type StageRequest,
} from './routeAdminApi';

// sonner is imported transitively by mutationClient; keep it inert in jsdom.
vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

function jsonResponse(body: unknown, init?: { status?: number; etag?: string }): Response {
  const headers = new Headers();
  if (init?.etag) headers.set('ETag', init.etag);
  return {
    ok: (init?.status ?? 200) < 400,
    status: init?.status ?? 200,
    headers,
    json: async () => body,
    clone() {
      return this as unknown as Response;
    },
  } as unknown as Response;
}

function stage(): RouteSummary['stages'][number] {
  return {
    order: 1,
    name: 'Review',
    required_role: 'approver',
    required_capability: 'doc.signoff',
    area_code: 'ops',
    quorum: 'any_1_of',
    quorum_m: null,
    drift_policy: 'reduce_quorum',
  };
}

function reqStage(): StageRequest {
  return {
    order: 1,
    name: 'Review',
    required_role: 'approver',
    required_capability: 'doc.signoff',
    area_code: 'ops',
    quorum: 'any_1_of',
    drift_policy: 'reduce_quorum',
  };
}

function route(overrides: Partial<RouteSummary> = {}): RouteSummary {
  return {
    id: 'route-a',
    name: 'Rota A',
    tenant_id: 'tenant-1',
    profile_code: 'ops',
    active: true,
    version: 3,
    stages: [stage()],
    created_at: '2026-04-20T10:00:00.000Z',
    updated_at: '2026-04-20T10:00:00.000Z',
    ...overrides,
  };
}

const fetchMock = vi.fn();

describe('routeAdminApi transport', () => {
  beforeEach(() => {
    etagCache.clear();
    fetchMock.mockReset();
    vi.stubGlobal('fetch', fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('listRoutes seeds the ETag cache per row using the v<N> format', async () => {
    const body: ListRoutesResponse = {
      routes: [route({ id: 'a', version: 3 }), route({ id: 'b', version: 1 })],
      total: 2,
    };
    fetchMock.mockResolvedValueOnce(jsonResponse(body));

    await listRoutes();

    expect(etagCache.get('a')).toBe('"v3"');
    expect(etagCache.get('b')).toBe('"v1"');
  });

  it('listRoutes skips legacy rows with version <= 0', async () => {
    const body: ListRoutesResponse = {
      routes: [route({ id: 'legacy', version: 0 })],
      total: 1,
    };
    fetchMock.mockResolvedValueOnce(jsonResponse(body));

    await listRoutes();

    expect(etagCache.get('legacy')).toBeUndefined();
  });

  it('createRoute sends the profile_code verbatim (lowercase) in the POST body', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ route_id: 'new-1' }, { etag: '"v1"' }));

    await createRoute({ profile_code: 'qa-preview', name: 'QA', stages: [reqStage()] });

    const [, init] = fetchMock.mock.calls[0];
    expect(init.method).toBe('POST');
    expect(JSON.parse(init.body).profile_code).toBe('qa-preview');
  });

  it('updateRoute sends If-Match seeded from a prior listRoutes (cold-load edit)', async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({ routes: [route({ id: 'route-a', version: 3 })], total: 1 }),
    );
    await listRoutes();

    fetchMock.mockResolvedValueOnce(jsonResponse({ route_id: 'route-a', new_version: 4 }));
    await updateRoute('route-a', { name: 'Rota A v2', stages: [reqStage()] });

    const [, init] = fetchMock.mock.calls[1];
    expect(init.method).toBe('PUT');
    expect(init.headers['If-Match']).toBe('"v3"');
  });
});
