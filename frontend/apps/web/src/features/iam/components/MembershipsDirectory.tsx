import { useMemo } from "react";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { IAM_AREA_ROLES, SP_DATE_TIME_FORMATTER } from "../constants";
import type { IamRole } from "../types";
import styles from "./MembershipsDirectory.module.css";

export type MembershipSortKey =
  | "userLabel"
  | "areaCode"
  | "role"
  | "effectiveFrom";
export type MembershipSortDir = "asc" | "desc";

export interface MembershipRow {
  userId: string;
  /** Display label resolved from the users directory; falls back to userId. */
  userLabel: string;
  areaCode: string;
  role: IamRole;
  effectiveFrom: string;
  grantedBy?: string | null;
}

interface MembershipsDirectoryProps {
  memberships: ReadonlyArray<MembershipRow>;
  isLoading: boolean;
  error: unknown;
  onRetry: () => void;
  sortKey: MembershipSortKey;
  sortDir: MembershipSortDir;
  onSort: (key: MembershipSortKey) => void;
  canManage: boolean;
  onRevoke: (row: MembershipRow) => void;
}

const SKELETON_ROWS = 8;

const ROLE_LABELS: Record<IamRole, string> = Object.fromEntries(
  IAM_AREA_ROLES.map(([value, label]) => [value, label]),
) as Record<IamRole, string>;

const COLUMNS: ReadonlyArray<{
  key: MembershipSortKey | "_actions";
  label: string;
  sortable?: boolean;
}> = [
  { key: "userLabel", label: "Usuário", sortable: true },
  { key: "areaCode", label: "Área", sortable: true },
  { key: "role", label: "Função", sortable: true },
  { key: "effectiveFrom", label: "Concedido em", sortable: true },
  { key: "_actions", label: "" },
];

function compareRows(a: MembershipRow, b: MembershipRow, key: MembershipSortKey): number {
  switch (key) {
    case "userLabel":
      return a.userLabel.localeCompare(b.userLabel, "pt-BR");
    case "areaCode":
      return a.areaCode.localeCompare(b.areaCode, "pt-BR");
    case "role":
      return a.role.localeCompare(b.role);
    case "effectiveFrom": {
      const av = a.effectiveFrom ? new Date(a.effectiveFrom).getTime() : 0;
      const bv = b.effectiveFrom ? new Date(b.effectiveFrom).getTime() : 0;
      return av - bv;
    }
  }
}

function formatGrantedAt(iso: string): string {
  if (!iso) return "—";
  const ts = new Date(iso);
  if (Number.isNaN(ts.getTime())) return "—";
  return SP_DATE_TIME_FORMATTER.format(ts);
}

export default function MembershipsDirectory({
  memberships,
  isLoading,
  error,
  onRetry,
  sortKey,
  sortDir,
  onSort,
  canManage,
  onRevoke,
}: MembershipsDirectoryProps) {
  const sorted = useMemo(() => {
    const copy = [...memberships];
    copy.sort((a, b) => {
      const diff = compareRows(a, b, sortKey);
      return sortDir === "asc" ? diff : -diff;
    });
    return copy;
  }, [memberships, sortKey, sortDir]);

  return (
    <div className={styles.wrapper} data-testid="memberships-directory">
      <div
        role="table"
        aria-label="Diretório de memberships de área"
        className={`${styles.grid} ${canManage ? styles.gridManage : styles.gridView}`}
      >
        <div role="rowgroup" className={styles.head}>
          <div role="row" style={{ display: "contents" }}>
            {COLUMNS.map((col) => {
              if (col.key === "_actions") {
                return (
                  <div role="columnheader" key="_actions" className={styles.th}>
                    <span className="visually-hidden">Ações</span>
                  </div>
                );
              }
              const isActiveSort = col.sortable && sortKey === col.key;
              return (
                <div
                  role="columnheader"
                  key={col.key}
                  className={styles.th}
                  aria-sort={
                    isActiveSort
                      ? sortDir === "asc"
                        ? "ascending"
                        : "descending"
                      : col.sortable
                        ? "none"
                        : undefined
                  }
                >
                  {col.sortable ? (
                    <button
                      type="button"
                      className={styles.sortable}
                      onClick={() => onSort(col.key as MembershipSortKey)}
                    >
                      {col.label}
                      <svg
                        viewBox="0 0 10 10"
                        className={`${styles.sortIcon} ${isActiveSort ? styles.sortIconActive : ""}`}
                        aria-hidden="true"
                      >
                        {isActiveSort && sortDir === "desc" ? (
                          <path d="M1 3 L5 7 L9 3" stroke="currentColor" fill="none" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                        ) : (
                          <path d="M1 7 L5 3 L9 7" stroke="currentColor" fill="none" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                        )}
                      </svg>
                    </button>
                  ) : (
                    <span>{col.label}</span>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        <div role="rowgroup" style={{ display: "contents" }}>
          {error ? (
            <div role="row" className={styles.error}>
              <span>{resolveQueryError(error, "Não foi possível carregar os memberships.")}</span>
              <button type="button" className={styles.retry} onClick={onRetry}>
                Tentar novamente
              </button>
            </div>
          ) : null}

          {isLoading && sorted.length === 0 ? (
            <div className={styles.skeleton} aria-hidden="true">
              {Array.from({ length: SKELETON_ROWS }).map((_, i) => (
                <div
                  key={i}
                  className={styles.skeletonRow}
                  style={{ gridTemplateColumns: "subgrid", gridColumn: "1 / -1" }}
                >
                  {COLUMNS.map((c, j) => (
                    <div
                      key={c.key + String(j)}
                      className={styles.skeletonBar}
                      style={{ maxWidth: `${50 + ((j * 13) % 40)}%` }}
                    />
                  ))}
                </div>
              ))}
            </div>
          ) : null}

          {!isLoading && !error && sorted.length === 0 ? (
            <div role="row" className={styles.empty}>
              <span className={styles.emptyTitle}>Nenhum membership encontrado</span>
              <span>Ajuste os filtros ou conceda acesso de área a um usuário.</span>
            </div>
          ) : null}

          {sorted.map((row) => (
            <div role="row" key={`${row.userId}:${row.areaCode}`} className={styles.row}>
              <div role="cell" className={styles.cell}>
                <span className={styles.userLabel}>{row.userLabel}</span>
              </div>
              <div role="cell" className={styles.cell}>
                <span className={styles.areaCode}>{row.areaCode}</span>
              </div>
              <div role="cell" className={styles.cell}>
                <span className={styles.roleBadge}>{ROLE_LABELS[row.role] ?? row.role}</span>
              </div>
              <div role="cell" className={`${styles.cell} ${styles.muted}`}>
                {formatGrantedAt(row.effectiveFrom)}
              </div>
              <div role="cell" className={`${styles.cell} ${styles.actionsCell}`}>
                {canManage ? (
                  <button
                    type="button"
                    className={styles.revokeBtn}
                    onClick={() => onRevoke(row)}
                    aria-label={`Revogar acesso de ${row.userLabel} à área ${row.areaCode}`}
                  >
                    Revogar
                  </button>
                ) : null}
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className={styles.footer}>
        <span>
          {sorted.length} membership{sorted.length === 1 ? "" : "s"} ativo
          {sorted.length === 1 ? "" : "s"}
        </span>
      </div>
    </div>
  );
}
