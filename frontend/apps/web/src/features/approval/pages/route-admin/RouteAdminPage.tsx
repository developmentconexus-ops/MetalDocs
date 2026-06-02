import { useRef, useState } from 'react';

import { ApprovalError } from '../../api/mutationClient';
import type { RouteSummary } from '../../api/routeAdminApi';
import {
  useCreateRoute,
  useDeactivateRoute,
  useUpdateRoute,
} from '../../queries/useRouteAdminMutations';
import { useRoutesQuery } from '../../queries/useRoutesQuery';
import { DeactivateRouteDialog } from './DeactivateRouteDialog';
import { RouteEditorDialog, type RouteEditorSubmission } from './RouteEditorDialog';
import { RouteListTable } from './RouteListTable';
import { messageForRouteError } from './routeAdminLabels';
import styles from './RouteAdmin.module.css';

export function RouteAdminPage() {
  const routesQuery = useRoutesQuery();
  const createMutation = useCreateRoute();
  const updateMutation = useUpdateRoute();
  const deactivateMutation = useDeactivateRoute();

  const [editorOpen, setEditorOpen] = useState(false);
  const [editingRoute, setEditingRoute] = useState<RouteSummary | null>(null);
  const [deactivateOpen, setDeactivateOpen] = useState(false);
  const [deactivateTarget, setDeactivateTarget] = useState<RouteSummary | null>(null);
  const [pageError, setPageError] = useState<string | null>(null);

  const newRouteButtonRef = useRef<HTMLButtonElement | null>(null);

  const openCreate = () => {
    setEditingRoute(null);
    setEditorOpen(true);
  };

  const openEdit = (route: RouteSummary) => {
    setEditingRoute(route);
    setEditorOpen(true);
  };

  const closeEditor = () => {
    setEditorOpen(false);
    setEditingRoute(null);
    // Return focus to the trigger so keyboard users land where they started.
    requestAnimationFrame(() => newRouteButtonRef.current?.focus());
  };

  const openDeactivate = (route: RouteSummary) => {
    setDeactivateTarget(route);
    setDeactivateOpen(true);
  };

  const closeDeactivate = () => {
    setDeactivateOpen(false);
    setDeactivateTarget(null);
  };

  const handleEditorSubmit = async (payload: RouteEditorSubmission) => {
    setPageError(null);
    if (payload.create) {
      await createMutation.mutateAsync(payload.create);
    } else if (payload.update && payload.routeId) {
      await updateMutation.mutateAsync({ routeId: payload.routeId, body: payload.update });
    }
    closeEditor();
  };

  const handleDeactivateConfirm = async (routeId: string, reason: string) => {
    setPageError(null);
    await deactivateMutation.mutateAsync({ routeId, body: { reason } });
    closeDeactivate();
  };

  const queryError = routesQuery.error;
  let loadErrorMessage: string | null = null;
  if (queryError) {
    if (queryError instanceof ApprovalError) {
      loadErrorMessage = messageForRouteError(queryError.code, queryError.status, queryError.message);
    } else {
      loadErrorMessage = 'Erro ao carregar rotas. Tente novamente.';
    }
  }

  const isMutating =
    createMutation.isPending || updateMutation.isPending || deactivateMutation.isPending;

  return (
    <section className={styles.page} aria-labelledby="route-admin-heading">
      <header className={styles.header}>
        <div>
          <h1 id="route-admin-heading">Administração de Rotas</h1>
          <p className={styles.subtitle}>
            Configure rotas de aprovação por perfil. Edições criam uma nova versão; desativações
            exigem motivo de governança.
          </p>
        </div>
        <button
          type="button"
          ref={newRouteButtonRef}
          className={styles.primaryButton}
          onClick={openCreate}
        >
          Nova rota
        </button>
      </header>

      {pageError ? (
        <div className={styles.errorBox} role="alert">
          {pageError}
        </div>
      ) : null}

      {loadErrorMessage ? (
        <div className={styles.errorBox} role="alert">
          {loadErrorMessage}
          <button
            type="button"
            className={styles.linkButton}
            onClick={() => void routesQuery.refetch()}
          >
            Tentar novamente
          </button>
        </div>
      ) : null}

      {routesQuery.isLoading ? (
        <p className={styles.state}>Carregando rotas…</p>
      ) : (
        <RouteListTable
          routes={routesQuery.data ?? []}
          isFetching={routesQuery.isFetching}
          onEdit={openEdit}
          onDeactivate={openDeactivate}
        />
      )}

      <RouteEditorDialog
        open={editorOpen}
        route={editingRoute}
        isSubmitting={isMutating}
        onSubmit={handleEditorSubmit}
        onClose={closeEditor}
      />

      <DeactivateRouteDialog
        open={deactivateOpen}
        route={deactivateTarget}
        isSubmitting={deactivateMutation.isPending}
        onConfirm={handleDeactivateConfirm}
        onClose={closeDeactivate}
      />
    </section>
  );
}
