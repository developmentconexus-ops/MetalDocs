import { formatISODate } from '../../../../lib/formatDate';
import type { RouteSummary } from '../../api/routeAdminApi';
import { tooltipForDisabledEdit } from './routeAdminLabels';
import styles from './RouteAdmin.module.css';

interface RouteListTableProps {
  routes: ReadonlyArray<RouteSummary>;
  isFetching: boolean;
  onEdit: (route: RouteSummary) => void;
  onDeactivate: (route: RouteSummary) => void;
}

function toLocalDate(iso: string): string {
  const date = new Date(iso);
  return Number.isNaN(date.getTime()) ? '-' : formatISODate(date);
}

export function RouteListTable({ routes, isFetching, onEdit, onDeactivate }: RouteListTableProps) {
  if (routes.length === 0) {
    return (
      <p className={styles.empty}>
        Nenhuma rota cadastrada. Use <strong>Nova rota</strong> para criar a primeira.
      </p>
    );
  }

  return (
    <div className={styles.tableWrap} aria-busy={isFetching}>
      <table className={styles.table}>
        <thead>
          <tr>
            <th scope="col">Nome</th>
            <th scope="col">Perfil</th>
            <th scope="col">Etapas</th>
            <th scope="col">Status</th>
            <th scope="col">Atualizado em</th>
            <th scope="col">Ações</th>
          </tr>
        </thead>
        <tbody>
          {routes.map((route) => {
            const editDisabled = !route.active;
            const deactivateDisabled = !route.active;
            const editTooltip = tooltipForDisabledEdit(route.active);

            return (
              <tr key={route.id}>
                <td>{route.name}</td>
                <td>
                  <code className={styles.code}>{route.profile_code}</code>
                </td>
                <td>{route.stages.length} etapa(s)</td>
                <td>
                  <span
                    className={route.active ? styles.badgeActive : styles.badgeInactive}
                    aria-label={route.active ? 'Status: ativa' : 'Status: inativa'}
                  >
                    <span aria-hidden="true" className={styles.badgeDot} />
                    {route.active ? 'Ativa' : 'Inativa'}
                  </span>
                </td>
                <td>{toLocalDate(route.updated_at)}</td>
                <td className={styles.rowActions}>
                  <button
                    type="button"
                    className={styles.secondaryButton}
                    aria-label={`Editar ${route.name}`}
                    disabled={editDisabled}
                    title={editTooltip}
                    onClick={() => onEdit(route)}
                  >
                    Editar
                  </button>
                  <button
                    type="button"
                    className={styles.warnButton}
                    aria-label={`Desativar ${route.name}`}
                    disabled={deactivateDisabled}
                    title={deactivateDisabled ? 'Rota já desativada.' : undefined}
                    onClick={() => onDeactivate(route)}
                  >
                    Desativar
                  </button>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
