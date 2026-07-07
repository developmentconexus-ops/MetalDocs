import type { ListInboxParams } from '../api/approvalTypes';
import styles from './InboxFilters.module.css';

// F5 (M2c C3): worklist filter state, decoupled from the query-param shape.
// `due` is a UI-level bucket ('all' | 'next7' | 'overdue') that `toInboxParams`
// maps to the server's `due_before` cutoff. Doc-type is deliberately absent —
// no `doc_type` param exists in ListInboxParams/ApprovalInboxItem (contract
// gap, deferred at HS-1 as D1). Do not add one here.
export interface InboxFilterState {
  stageKind?: 'review' | 'approval';
  due?: 'all' | 'next7' | 'overdue';
  oversee: boolean;
}

export const DEFAULT_INBOX_FILTER_STATE: InboxFilterState = {
  stageKind: undefined,
  due: 'all',
  oversee: false,
};

const DAY_MS = 24 * 60 * 60 * 1000;

// Pure mapping from UI filter state to the API's ListInboxParams. `due_before`
// is a server-side pre-filter hint only — "Atrasadas" additionally needs a
// client-side `overdue` guard on the results (see InboxPage), since
// due_before=now is inclusive of items due later today.
export function toInboxParams(state: InboxFilterState, now: number = Date.now()): ListInboxParams {
  const params: ListInboxParams = {};
  if (state.stageKind) {
    params.stage_kind = state.stageKind;
  }
  if (state.due === 'next7') {
    params.due_before = new Date(now + 7 * DAY_MS).toISOString();
  } else if (state.due === 'overdue') {
    params.due_before = new Date(now).toISOString();
  }
  if (state.oversee) {
    params.scope = 'oversee';
  }
  return params;
}

export function isInboxFilterActive(state: InboxFilterState): boolean {
  return Boolean(state.stageKind) || (state.due != null && state.due !== 'all') || state.oversee;
}

interface InboxFiltersProps {
  value: InboxFilterState;
  onChange: (next: InboxFilterState) => void;
  oversee403?: boolean;
}

export function InboxFilters({ value, onChange, oversee403 }: InboxFiltersProps) {
  return (
    <div className={styles.filters}>
      <div className={styles.field}>
        <label htmlFor="inbox-filter-stage-kind">Estágio</label>
        <select
          id="inbox-filter-stage-kind"
          value={value.stageKind ?? ''}
          onChange={(e) => {
            const raw = e.target.value;
            onChange({
              ...value,
              stageKind: raw === '' ? undefined : (raw as 'review' | 'approval'),
            });
          }}
        >
          <option value="">Todos</option>
          <option value="review">Revisão</option>
          <option value="approval">Aprovação</option>
        </select>
      </div>

      <div className={styles.field}>
        <label htmlFor="inbox-filter-due">Vencimento</label>
        <select
          id="inbox-filter-due"
          value={value.due ?? 'all'}
          onChange={(e) => {
            onChange({ ...value, due: e.target.value as InboxFilterState['due'] });
          }}
        >
          <option value="all">Todas</option>
          <option value="next7">Vence em 7 dias</option>
          <option value="overdue">Atrasadas</option>
        </select>
      </div>

      <label className={styles.overseeToggle} htmlFor="inbox-filter-oversee">
        <input
          id="inbox-filter-oversee"
          type="checkbox"
          checked={value.oversee}
          onChange={(e) => {
            onChange({ ...value, oversee: e.target.checked });
          }}
        />
        Supervisão (oversee)
      </label>

      {oversee403 ? (
        <div className={styles.overseeDenied} role="alert">
          Você não tem permissão de supervisão.
        </div>
      ) : null}
    </div>
  );
}
