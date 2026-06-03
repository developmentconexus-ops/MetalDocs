import { useMemo } from "react";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import type { ManagedUser } from "../presenters/user-status-presenter";
import UserRow from "./UserRow";
import styles from "./UsersDirectory.module.css";

export type SortKey = "displayName" | "email" | "tenantRole" | "lastLoginAt" | "mfa";
export type SortDir = "asc" | "desc";

interface UsersDirectoryProps {
  users: ReadonlyArray<ManagedUser>;
  isLoading: boolean;
  error: unknown;
  onRetry: () => void;
  selectedIds: ReadonlyArray<string>;
  onToggleSelect: (userId: string) => void;
  onToggleSelectAll: () => void;
  onOpen: (userId: string) => void;
  sortKey: SortKey;
  sortDir: SortDir;
  onSort: (key: SortKey) => void;
  pageSize: number;
  totalLoaded: number;
  hasMore: boolean;
  onLoadMore: () => void;
  isFetchingMore: boolean;
}

const SKELETON_ROWS = 8;

const COLUMNS: ReadonlyArray<{
  key: SortKey | "_select" | "_areas" | "_status" | "_actions";
  label: string;
  sortable?: boolean;
}> = [
  { key: "_select", label: "" },
  { key: "displayName", label: "Usuário", sortable: true },
  { key: "email", label: "E-mail", sortable: true },
  { key: "tenantRole", label: "Função", sortable: true },
  { key: "_areas", label: "Áreas" },
  { key: "_status", label: "Status" },
  { key: "lastLoginAt", label: "Último acesso", sortable: true },
  { key: "mfa", label: "MFA", sortable: true },
  { key: "_actions", label: "" },
];

function compareUsers(a: ManagedUser, b: ManagedUser, key: SortKey): number {
  switch (key) {
    case "displayName":
      return (a.displayName || a.username).localeCompare(b.displayName || b.username, "pt-BR");
    case "email":
      return (a.email ?? "").localeCompare(b.email ?? "", "pt-BR");
    case "tenantRole":
      return a.tenantRole.localeCompare(b.tenantRole);
    case "lastLoginAt": {
      const av = a.lastLoginAt ? new Date(a.lastLoginAt).getTime() : 0;
      const bv = b.lastLoginAt ? new Date(b.lastLoginAt).getTime() : 0;
      return av - bv;
    }
    case "mfa":
      return Number(!!a.mfaEnabled) - Number(!!b.mfaEnabled);
  }
}

export default function UsersDirectory({
  users,
  isLoading,
  error,
  onRetry,
  selectedIds,
  onToggleSelect,
  onToggleSelectAll,
  onOpen,
  sortKey,
  sortDir,
  onSort,
  totalLoaded,
  hasMore,
  onLoadMore,
  isFetchingMore,
}: UsersDirectoryProps) {
  const sorted = useMemo(() => {
    const copy = [...users];
    copy.sort((a, b) => {
      const diff = compareUsers(a, b, sortKey);
      return sortDir === "asc" ? diff : -diff;
    });
    return copy;
  }, [users, sortKey, sortDir]);

  const selectedSet = useMemo(() => new Set(selectedIds), [selectedIds]);

  const allOnPageSelected = sorted.length > 0 && sorted.every((u) => selectedSet.has(u.userId));

  return (
    <div
      className={styles.wrapper}
      data-testid="users-directory"
    >
      <div role="table" aria-label="Diretório de usuários" className={styles.grid}>
        <div role="rowgroup" className={styles.head}>
          <div role="row" style={{ display: "contents" }}>
            {COLUMNS.map((col) => {
              if (col.key === "_select") {
                return (
                  <div role="columnheader" key="_select" className={styles.th}>
                    <input
                      type="checkbox"
                      className={styles.headCheckbox}
                      checked={allOnPageSelected}
                      onChange={onToggleSelectAll}
                      aria-label="Selecionar todos visíveis"
                    />
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
                      onClick={() => onSort(col.key as SortKey)}
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
              <span>{resolveQueryError(error, "Não foi possível carregar os usuários.")}</span>
              <button type="button" className={styles.retry} onClick={onRetry}>
                Tentar novamente
              </button>
            </div>
          ) : null}

          {isLoading && sorted.length === 0 ? (
            <div className={styles.skeleton} aria-hidden="true">
              {Array.from({ length: SKELETON_ROWS }).map((_, i) => (
                <div key={i} className={styles.skeletonRow} style={{ gridTemplateColumns: "subgrid", gridColumn: "1 / -1" }}>
                  {COLUMNS.map((c, j) => (
                    <div key={c.key + j} className={styles.skeletonBar} style={{ maxWidth: `${50 + ((j * 13) % 40)}%` }} />
                  ))}
                </div>
              ))}
            </div>
          ) : null}

          {!isLoading && !error && sorted.length === 0 ? (
            <div role="row" className={styles.empty}>
              <span className={styles.emptyTitle}>Nenhum usuário encontrado</span>
              <span>Ajuste os filtros ou convide novos membros para o workspace.</span>
            </div>
          ) : null}

          {sorted.map((user) => (
            <UserRow
              key={user.userId}
              user={user}
              selected={selectedSet.has(user.userId)}
              onToggleSelect={onToggleSelect}
              onOpen={onOpen}
            />
          ))}
        </div>
      </div>

      <div className={styles.footer}>
        <span>
          {totalLoaded} usuário{totalLoaded === 1 ? "" : "s"} carregado{totalLoaded === 1 ? "" : "s"}
          {selectedIds.length > 0 ? ` · ${selectedIds.length} selecionado${selectedIds.length === 1 ? "" : "s"}` : null}
        </span>
        <button
          type="button"
          className={styles.pageBtn}
          onClick={onLoadMore}
          disabled={!hasMore || isFetchingMore}
        >
          {isFetchingMore ? "Carregando…" : hasMore ? "Carregar mais" : "Fim da lista"}
        </button>
      </div>
    </div>
  );
}
