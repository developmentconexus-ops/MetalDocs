import { useMutation, useQueryClient } from '@tanstack/react-query';

import { QK } from '../../../lib/queryKeys';
import {
  createRoute,
  deactivateRoute,
  seedRouteEtag,
  updateRoute,
  type CreateRouteRequest,
  type DeactivateRouteRequest,
  type ListRoutesResponse,
  type RouteResponse,
  type RouteSummary,
  type UpdateRouteRequest,
} from '../api/routeAdminApi';

interface ListSnapshot {
  previous: ListRoutesResponse | undefined;
}

function nowIso(): string {
  return new Date().toISOString();
}

function stageRequestToSummary(
  stage: CreateRouteRequest['stages'][number],
): RouteSummary['stages'][number] {
  return {
    label: stage.name,
    required_role: stage.required_role,
    required_capability: stage.required_capability,
    area_code: stage.area_code,
    quorum_kind: stage.quorum,
    quorum_m: stage.quorum_m ?? null,
    drift_policy: stage.drift_policy,
  };
}

/**
 * Create mutation — non-optimistic.
 *
 * The previous implementation inserted a placeholder row keyed by a synthetic
 * `optimistic-…` id. That id leaked into `RouteListTable` for the window
 * between `onMutate` and the `onSettled` refetch, letting the user click
 * Edit/Deactivate on a row whose backing UUID did not exist yet — every
 * follow-up `PUT`/`DELETE` then fired against a non-UUID path and 404'd.
 * Server creates are sub-second; the cost of a true optimistic insert is
 * higher than the latency it saves, so we skip it and rely on
 * `invalidateQueries` plus the trigger's `isPending` state for feedback.
 *
 * On success we seed the new route's ETag at `v1` (per ADR 0018 routes are
 * created at version 1) so an immediate Edit does not race the list refetch
 * for the `If-Match` token.
 */
export function useCreateRoute() {
  const queryClient = useQueryClient();
  const queryKey = QK.approval.routes.list();

  return useMutation<RouteResponse, Error, CreateRouteRequest>({
    mutationFn: (body) => createRoute(body),
    onSuccess: (response) => {
      seedRouteEtag(response.route_id, 1);
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });
}

/**
 * Optimistic update: rewrites the affected row in the cached list, bumps the
 * local version, and refreshes the ETag cache so the next PUT carries the new
 * If-Match. Rolls back on any error (412 stale, 409 active-instance, 422).
 */
export function useUpdateRoute() {
  const queryClient = useQueryClient();
  const queryKey = QK.approval.routes.list();

  return useMutation<
    RouteResponse,
    Error,
    { routeId: string; body: UpdateRouteRequest },
    ListSnapshot
  >({
    mutationFn: ({ routeId, body }) => updateRoute(routeId, body),
    onMutate: async ({ routeId, body }) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<ListRoutesResponse>(queryKey);
      if (!previous) {
        return { previous };
      }

      const nextRoutes = previous.routes.map((route): RouteSummary => {
        if (route.id !== routeId) {
          return route;
        }
        return {
          ...route,
          name: body.name,
          stages: body.stages.map(stageRequestToSummary),
          version: route.version + 1,
          updated_at: nowIso(),
        };
      });

      queryClient.setQueryData<ListRoutesResponse>(queryKey, {
        ...previous,
        routes: nextRoutes,
      });

      return { previous };
    },
    onSuccess: (response, variables) => {
      if (response.new_version != null) {
        seedRouteEtag(variables.routeId, response.new_version);
      }
    },
    onError: (_err, _vars, context) => {
      if (context?.previous !== undefined) {
        queryClient.setQueryData(queryKey, context.previous);
      }
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });
}

/**
 * Optimistic deactivate: flips `active=false` on the cached row, bumps the
 * version, and re-seeds the ETag cache on success. Rolls back on error.
 */
export function useDeactivateRoute() {
  const queryClient = useQueryClient();
  const queryKey = QK.approval.routes.list();

  return useMutation<
    RouteResponse,
    Error,
    { routeId: string; body: DeactivateRouteRequest },
    ListSnapshot
  >({
    mutationFn: ({ routeId, body }) => deactivateRoute(routeId, body),
    onMutate: async ({ routeId }) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<ListRoutesResponse>(queryKey);
      if (!previous) {
        return { previous };
      }

      const nextRoutes = previous.routes.map((route): RouteSummary =>
        route.id === routeId
          ? { ...route, active: false, version: route.version + 1, updated_at: nowIso() }
          : route,
      );

      queryClient.setQueryData<ListRoutesResponse>(queryKey, {
        ...previous,
        routes: nextRoutes,
      });

      return { previous };
    },
    onSuccess: (response, variables) => {
      if (response.new_version != null) {
        seedRouteEtag(variables.routeId, response.new_version);
      }
    },
    onError: (_err, _vars, context) => {
      if (context?.previous !== undefined) {
        queryClient.setQueryData(queryKey, context.previous);
      }
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
  });
}
