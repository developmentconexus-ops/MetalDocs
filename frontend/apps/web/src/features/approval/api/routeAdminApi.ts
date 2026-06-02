// Thin transport layer for the approval/routes admin endpoints.
//
// All writes go through `mutate()` so they share idempotency-key, If-Match /
// ETag handling, and RFC 9457 `problem+json` error decoding with the rest of
// the approval module. Reads use plain `fetch` and surface the ETag so the
// cache stays warm for the next write.
//
// Types come from the OpenAPI codegen (`components['schemas']`) — never
// re-typed locally.

import { mutate, type MutateOptions } from './mutationClient';
import { etagCache } from './etagCache';
import type { components } from '../../../lib/api-types';

const BASE = '/api/v1/approval/routes';

export type RouteSummary = components['schemas']['RouteSummary'];
export type ListRoutesResponse = components['schemas']['ListRoutesResponse'];
export type CreateRouteRequest = components['schemas']['CreateRouteRequest'];
export type UpdateRouteRequest = components['schemas']['UpdateRouteRequest'];
export type DeactivateRouteRequest = components['schemas']['DeactivateRouteRequest'];
export type RouteResponse = components['schemas']['RouteResponse'];
export type StageRequest = components['schemas']['StageRequest'];

export async function listRoutes(): Promise<ListRoutesResponse> {
  const res = await fetch(BASE);
  if (!res.ok) {
    throw Object.assign(new Error(`http_${res.status}`), { status: res.status });
  }
  return (await res.json()) as ListRoutesResponse;
}

export function createRoute(
  body: CreateRouteRequest,
  opts?: MutateOptions,
): Promise<RouteResponse> {
  return mutate<CreateRouteRequest, RouteResponse>('POST', BASE, body, opts);
}

export function updateRoute(
  routeId: string,
  body: UpdateRouteRequest,
  opts?: MutateOptions,
): Promise<RouteResponse> {
  return mutate<UpdateRouteRequest, RouteResponse>(
    'PUT',
    `${BASE}/${encodeURIComponent(routeId)}`,
    body,
    { resourceId: routeId, ...opts },
  );
}

export function deactivateRoute(
  routeId: string,
  body: DeactivateRouteRequest,
  opts?: MutateOptions,
): Promise<RouteResponse> {
  return mutate<DeactivateRouteRequest, RouteResponse>(
    'DELETE',
    `${BASE}/${encodeURIComponent(routeId)}`,
    body,
    { resourceId: routeId, ...opts },
  );
}

/**
 * Seeds the ETag cache for a route. Useful when the list endpoint returns a
 * `version` that callers want to use as the optimistic-concurrency token for
 * the next update.
 */
export function seedRouteEtag(routeId: string, version: number): void {
  etagCache.set(routeId, `"${version}"`);
}
