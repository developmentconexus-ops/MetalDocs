// Thin transport layer for the approval/routes admin endpoints.
//
// All writes go through `mutate()` so they share idempotency-key, If-Match /
// ETag handling, and RFC 9457 `problem+json` error decoding with the rest of
// the approval module. Reads also go through the shared problem decoder via
// `parseProblem`, so 401 dispatches `authn.expired` and unknown error codes
// surface through `ApprovalError` for the same `resolveErrorMessage` path.
//
// Types come from the OpenAPI codegen (`components['schemas']`) — never
// re-typed locally.

import { dispatchAuthExpired } from '../../../lib/api';
import { parseProblem } from '../../../lib/api/problem';
import { ApprovalError, mutate, type MutateOptions } from './mutationClient';
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
  if (res.ok) {
    return (await res.json()) as ListRoutesResponse;
  }

  const prob = await parseProblem(res.clone());
  if (prob) {
    if (prob.status === 401) {
      dispatchAuthExpired(window.location.pathname + window.location.search);
    }
    throw new ApprovalError(prob.code, prob.status, prob.title ?? prob.detail ?? '');
  }
  if (res.status === 401) {
    dispatchAuthExpired(window.location.pathname + window.location.search);
    throw new ApprovalError('authn.expired', 401, 'Não autorizado');
  }
  throw new ApprovalError(`http_${res.status}`, res.status, 'Erro interno');
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
 * Seeds the ETag cache for a route. The format MUST match what the backend
 * `parseIfMatch` accepts: a quoted `v<N>` token (`"v5"`). Other approval
 * surfaces (submit, get-instance) already follow this contract — writing the
 * raw version (`"5"`) here would cause `parseIfMatch` to reject every
 * subsequent `PUT`/`DELETE` with 400 `validation.if_match_malformed`.
 *
 * Seeding is monotonic: a stale list refetch that returns an older version
 * must not clobber a fresher ETag a mutation just wrote.
 */
export function seedRouteEtag(routeId: string, version: number): void {
  if (!Number.isFinite(version) || version <= 0) return;
  const next = `"v${version}"`;
  const current = etagCache.get(routeId);
  if (current) {
    const currentVersion = parseEtagVersion(current);
    if (currentVersion != null && currentVersion >= version) return;
  }
  etagCache.set(routeId, next);
}

function parseEtagVersion(etag: string): number | null {
  const trimmed = etag.trim().replace(/^"|"$/g, '');
  if (!trimmed.startsWith('v')) return null;
  const n = Number(trimmed.slice(1));
  return Number.isFinite(n) && n > 0 ? n : null;
}
