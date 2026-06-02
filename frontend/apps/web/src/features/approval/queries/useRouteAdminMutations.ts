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

/**
 * Optimistic create: inserts a placeholder row keyed by a temporary id, then
 * replaces it with the server's `route_id` on success. Rolls the full list
 * back to its snapshot on error (409, 422, network).
 */
export function useCreateRoute() {
  const queryClient = useQueryClient();
  const queryKey = QK.approval.routes.list();

  return useMutation<RouteResponse, Error, CreateRouteRequest, ListSnapshot>({
    mutationFn: (body) => createRoute(body),
    onMutate: async (body) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<ListRoutesResponse>(queryKey);

      const optimistic: RouteSummary = {
        id: `optimistic-${body.name}-${nowIso()}`,
        name: body.name,
        tenant_id: '',
        profile_code: body.profile_code,
        active: true,
        version: 1,
        stages: body.stages.map((stage) => ({
          label: stage.name,
          required_role: stage.required_role,
          required_capability: stage.required_capability,
          area_code: stage.area_code,
          quorum_kind: stage.quorum,
          quorum_m: stage.quorum_m ?? null,
          drift_policy: stage.drift_policy,
        })),
        created_at: nowIso(),
        updated_at: nowIso(),
      };

      const base: ListRoutesResponse = previous ?? { routes: [], total: 0 };
      queryClient.setQueryData<ListRoutesResponse>(queryKey, {
        ...base,
        routes: [...base.routes, optimistic],
        total: base.total + 1,
      });

      return { previous };
    },
    onError: (_err, _body, context) => {
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
          stages: body.stages.map((stage) => ({
            label: stage.name,
            required_role: stage.required_role,
            required_capability: stage.required_capability,
            area_code: stage.area_code,
            quorum_kind: stage.quorum,
            quorum_m: stage.quorum_m ?? null,
            drift_policy: stage.drift_policy,
          })),
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
