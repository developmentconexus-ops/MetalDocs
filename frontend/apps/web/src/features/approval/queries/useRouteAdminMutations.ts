import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { QK } from '../../../lib/queryKeys';
import { resolveErrorMessage } from '../../../lib/api/problem';
import { ApprovalError } from '../api/mutationClient';
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
 * Surfaces a non-412 mutation failure as a toast. 412 (stale ETag) is already
 * toasted inside `mutationClient`, so we skip it here to avoid a duplicate.
 */
function toastMutationError(err: unknown): void {
  if (err instanceof ApprovalError && err.status === 412) return;
  toast.error(resolveErrorMessage(err));
}

function stageRequestToSummary(
  stage: CreateRouteRequest['stages'][number],
): RouteSummary['stages'][number] {
  return {
    order: stage.order,
    name: stage.name,
    required_role: stage.required_role,
    required_capability: stage.required_capability,
    area_code: stage.area_code,
    quorum: stage.quorum,
    quorum_m: stage.quorum_m ?? null,
    drift_policy: stage.drift_policy,
  };
}

/**
 * Options for the shared optimistic-mutation factory used by update and
 * deactivate. Create has a distinct non-optimistic shape and is not folded in.
 */
interface OptimisticRouteMutationOptions<TVariables extends { routeId: string }> {
  mutationFn: (variables: TVariables) => Promise<RouteResponse>;
  applyOptimistic: (route: RouteSummary, variables: TVariables) => RouteSummary;
  successMessage: string;
}

/**
 * Factory that produces the shared onMutate/onError/onSuccess/onSettled
 * scaffold used by `useUpdateRoute` and `useDeactivateRoute`.
 *
 * Both mutations snapshot the list cache, apply an optimistic row patch, roll
 * back on any error (412 stale, 409 active-instance, 422), re-seed the ETag
 * on success, and invalidate the list on settled. Only the row-patch function
 * and success-toast text differ between the two callers.
 */
function makeOptimisticRouteMutation<TVariables extends { routeId: string }>(
  queryClient: ReturnType<typeof useQueryClient>,
  queryKey: readonly unknown[],
  options: OptimisticRouteMutationOptions<TVariables>,
) {
  const { mutationFn, applyOptimistic, successMessage } = options;

  return {
    mutationFn,
    onMutate: async (variables: TVariables): Promise<ListSnapshot> => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<ListRoutesResponse>(queryKey);
      if (!previous) {
        return { previous };
      }

      const nextRoutes = previous.routes.map((route): RouteSummary =>
        route.id === variables.routeId ? applyOptimistic(route, variables) : route,
      );

      queryClient.setQueryData<ListRoutesResponse>(queryKey, {
        ...previous,
        routes: nextRoutes,
      });

      return { previous };
    },
    onSuccess: (response: RouteResponse, variables: TVariables): void => {
      if (response.new_version != null) {
        seedRouteEtag(variables.routeId, response.new_version);
      }
      toast.success(successMessage);
    },
    onError: (err: unknown, _vars: TVariables, context: ListSnapshot | undefined): void => {
      if (context?.previous !== undefined) {
        queryClient.setQueryData(queryKey, context.previous);
      }
      toastMutationError(err);
    },
    onSettled: (): void => {
      void queryClient.invalidateQueries({ queryKey });
    },
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
      toast.success('Rota criada.');
    },
    onError: (err) => {
      toastMutationError(err);
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
  >(
    makeOptimisticRouteMutation(queryClient, queryKey, {
      mutationFn: ({ routeId, body }) => updateRoute(routeId, body),
      applyOptimistic: (route, { body }) => ({
        ...route,
        name: body.name,
        stages: body.stages.map(stageRequestToSummary),
        version: (route.version ?? 0) + 1,
        updated_at: nowIso(),
      }),
      successMessage: 'Rota atualizada.',
    }),
  );
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
  >(
    makeOptimisticRouteMutation(queryClient, queryKey, {
      mutationFn: ({ routeId, body }) => deactivateRoute(routeId, body),
      applyOptimistic: (route) => ({
        ...route,
        active: false,
        version: (route.version ?? 0) + 1,
        updated_at: nowIso(),
      }),
      successMessage: 'Rota desativada.',
    }),
  );
}
