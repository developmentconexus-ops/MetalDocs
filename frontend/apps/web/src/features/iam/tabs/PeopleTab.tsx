import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import { SearchBar } from "../../../components/ui/SearchBar";
import { Dialog } from "../../../components/ui/Dialog";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useUsersQuery } from "../queries/useUsersQuery";
import { useBulkUsersMutation } from "../mutations/useBulkUsersMutation";
import UsersDirectory, {
  type SortKey,
  type SortDir,
} from "../components/UsersDirectory";
import UsersFilterBar, {
  type UsersFilterValue,
  type StatusFacet,
  type MfaFacet,
  type LastLoginFacet,
} from "../components/UsersFilterBar";
import InviteUserDialog from "../components/InviteUserDialog";
import {
  matchesArea,
  matchesLastLogin,
  matchesMfa,
  matchesStatusFacet,
} from "../presenters/user-status-presenter";
import type { IamRole } from "../types";
import styles from "./PeopleTab.module.css";

type BulkActionKind = "deactivate" | "force-logout" | "unlock";

const VALID_STATUS: ReadonlyArray<StatusFacet> = [
  "active",
  "pending",
  "suspended",
  "locked",
];
const VALID_ROLE: ReadonlyArray<IamRole> = [
  "system_admin",
  "qms_admin",
  "area_admin",
  "approver",
  "author",
  "editor",
  "signer",
  "viewer",
];
const VALID_MFA: ReadonlyArray<MfaFacet> = ["on", "off"];
const VALID_LAST: ReadonlyArray<LastLoginFacet> = ["24h", "7d", "30d", "90d"];

function readEnum<T extends string>(
  raw: string | null,
  allowed: ReadonlyArray<T>,
): T | undefined {
  if (!raw) return undefined;
  return allowed.includes(raw as T) ? (raw as T) : undefined;
}

const PAGE_SIZE = 50;

export default function PeopleTab() {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  const filterValue: UsersFilterValue = useMemo(
    () => ({
      status: readEnum(searchParams.get("status"), VALID_STATUS),
      role: readEnum(searchParams.get("role"), VALID_ROLE),
      area: searchParams.get("area") ?? undefined,
      mfa: readEnum(searchParams.get("mfa"), VALID_MFA),
      lastLogin: readEnum(searchParams.get("lastLogin"), VALID_LAST),
    }),
    [searchParams],
  );
  const q = searchParams.get("q") ?? "";
  const sortKey = (searchParams.get("sort") as SortKey) ?? "displayName";
  const sortDir = (searchParams.get("dir") as SortDir) ?? "asc";

  const [limit, setLimit] = useState(PAGE_SIZE);
  const [selectedIds, setSelectedIds] = useState<ReadonlyArray<string>>([]);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [pendingAction, setPendingAction] = useState<BulkActionKind | null>(null);

  const serverIsActive: boolean | undefined =
    filterValue.status === "active"
      ? true
      : filterValue.status === "suspended"
        ? false
        : undefined;

  const usersQuery = useUsersQuery({
    limit,
    q: q.trim() || undefined,
    role: filterValue.role,
    areaCode: filterValue.area,
    isActive: serverIsActive,
  });

  const bulkMutation = useBulkUsersMutation();

  const allLoaded = usersQuery.data?.items ?? [];

  const filtered = useMemo(() => {
    return allLoaded.filter(
      (u) =>
        matchesStatusFacet(u, filterValue.status) &&
        matchesArea(u, filterValue.area) &&
        matchesMfa(u, filterValue.mfa) &&
        matchesLastLogin(u, filterValue.lastLogin),
    );
  }, [allLoaded, filterValue]);

  const areaCodes = useMemo(() => {
    const set = new Set<string>();
    for (const u of allLoaded) {
      for (const m of u.areaMemberships) set.add(m.areaCode);
    }
    return Array.from(set).sort((a, b) => a.localeCompare(b, "pt-BR"));
  }, [allLoaded]);

  useEffect(() => {
    // Drop selections that fall outside the current filtered set.
    if (selectedIds.length === 0) return;
    const visibleIds = new Set(filtered.map((u) => u.userId));
    const next = selectedIds.filter((id) => visibleIds.has(id));
    if (next.length !== selectedIds.length) setSelectedIds(next);
  }, [filtered, selectedIds]);

  const updateParam = useCallback(
    (key: string, value: string | undefined) => {
      const next = new URLSearchParams(searchParams);
      if (value === undefined || value === "") next.delete(key);
      else next.set(key, value);
      setSearchParams(next, { replace: true });
    },
    [searchParams, setSearchParams],
  );

  const handleFilterChange = useCallback(
    (next: UsersFilterValue) => {
      const params = new URLSearchParams(searchParams);
      const apply = (key: string, val: string | undefined) => {
        if (val) params.set(key, val);
        else params.delete(key);
      };
      apply("status", next.status);
      apply("role", next.role);
      apply("area", next.area);
      apply("mfa", next.mfa);
      apply("lastLogin", next.lastLogin);
      setSearchParams(params, { replace: true });
      setLimit(PAGE_SIZE);
    },
    [searchParams, setSearchParams],
  );

  const handleSearchChange = useCallback(
    (value: string) => {
      updateParam("q", value || undefined);
      setLimit(PAGE_SIZE);
    },
    [updateParam],
  );

  const handleSort = useCallback(
    (key: SortKey) => {
      const params = new URLSearchParams(searchParams);
      if (sortKey === key) {
        params.set("dir", sortDir === "asc" ? "desc" : "asc");
      } else {
        params.set("sort", key);
        params.set("dir", "asc");
      }
      setSearchParams(params, { replace: true });
    },
    [searchParams, setSearchParams, sortKey, sortDir],
  );

  const handleToggleSelect = useCallback((userId: string) => {
    setSelectedIds((prev) =>
      prev.includes(userId) ? prev.filter((x) => x !== userId) : [...prev, userId],
    );
  }, []);

  const handleToggleSelectAll = useCallback(() => {
    setSelectedIds((prev) => {
      const visibleIds = filtered.map((u) => u.userId);
      const allSelected = visibleIds.every((id) => prev.includes(id));
      if (allSelected) return prev.filter((id) => !visibleIds.includes(id));
      const set = new Set([...prev, ...visibleIds]);
      return Array.from(set);
    });
  }, [filtered]);

  const handleOpen = useCallback(
    (userId: string) => {
      navigate({
        pathname: `/admin/people/${userId}`,
        search: searchParams.toString(),
      });
    },
    [navigate, searchParams],
  );

  const handleConfirmBulk = () => {
    if (!pendingAction || selectedIds.length === 0) return;
    const action = pendingAction;
    const userIds = [...selectedIds];

    bulkMutation.mutate(
      { userIds, action },
      {
        onSuccess: (result) => {
          const succeeded = result?.succeeded?.length ?? 0;
          const failed = result?.failed?.length ?? 0;
          if (failed === 0) {
            toast.success(`Ação aplicada a ${succeeded} usuário(s).`);
          } else {
            toast.warning(
              `${succeeded} sucesso(s), ${failed} falha(s). Verifique a auditoria.`,
            );
          }
          setSelectedIds([]);
        },
        onError: (err) => {
          toast.error(resolveQueryError(err, "Falha na ação em lote."));
        },
      },
    );
    setPendingAction(null);
  };

  const ACTION_LABEL: Record<BulkActionKind, string> = {
    deactivate: "desativar",
    "force-logout": "encerrar sessões",
    unlock: "desbloquear",
  };

  return (
    <section
      className={styles.tab}
      data-testid="admin-people-tab"
      aria-labelledby="admin-people-heading"
    >
      <div className={styles.headerRow}>
        <div className={styles.titleBlock}>
          <h2 id="admin-people-heading">Pessoas</h2>
          <p className={styles.lede}>
            Diretório de usuários do workspace. Convide, gerencie funções e revise atividades.
          </p>
        </div>
        <button
          type="button"
          className={styles.primaryAction}
          onClick={() => setInviteOpen(true)}
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <path d="M12 5v14M5 12h14" />
          </svg>
          Convidar usuário
        </button>
      </div>

      <div className={styles.toolbar}>
        <div className={styles.searchBox}>
          <SearchBar
            value={q}
            onChange={handleSearchChange}
            placeholder="Buscar por nome, e-mail ou username"
            ariaLabel="Buscar usuários"
          />
        </div>
        <UsersFilterBar
          value={filterValue}
          onChange={handleFilterChange}
          areaCodes={areaCodes}
        />
      </div>

      {selectedIds.length > 0 ? (
        <div className={styles.bulkBar} role="region" aria-label="Ações em lote">
          <span>
            <strong>{selectedIds.length}</strong> selecionado(s)
          </span>
          <div className={styles.bulkActions}>
            <button
              type="button"
              className={`${styles.bulkBtn} ${styles.bulkBtnDanger}`}
              onClick={() => setPendingAction("deactivate")}
              disabled={bulkMutation.isPending}
            >
              Desativar
            </button>
            <button
              type="button"
              className={styles.bulkBtn}
              onClick={() => setPendingAction("force-logout")}
              disabled={bulkMutation.isPending}
            >
              Encerrar sessões
            </button>
            <button
              type="button"
              className={styles.bulkBtn}
              onClick={() => setPendingAction("unlock")}
              disabled={bulkMutation.isPending}
            >
              Desbloquear
            </button>
            <button
              type="button"
              className={styles.clearBulk}
              onClick={() => setSelectedIds([])}
            >
              Limpar seleção
            </button>
          </div>
        </div>
      ) : null}

      <UsersDirectory
        users={filtered}
        isLoading={usersQuery.isLoading}
        error={usersQuery.error}
        onRetry={() => usersQuery.refetch()}
        selectedIds={selectedIds}
        onToggleSelect={handleToggleSelect}
        onToggleSelectAll={handleToggleSelectAll}
        onOpen={handleOpen}
        sortKey={sortKey}
        sortDir={sortDir}
        onSort={handleSort}
        pageSize={PAGE_SIZE}
        totalLoaded={filtered.length}
        hasMore={usersQuery.data?.page?.has_more ?? false}
        onLoadMore={() => setLimit((l) => l + PAGE_SIZE)}
        isFetchingMore={usersQuery.isFetching && !usersQuery.isLoading}
      />

      <InviteUserDialog open={inviteOpen} onClose={() => setInviteOpen(false)} />

      <Dialog
        open={pendingAction !== null}
        onClose={() => setPendingAction(null)}
        title="Confirmar ação em lote"
        footer={
          <>
            <button
              type="button"
              className={styles.bulkBtn}
              onClick={() => setPendingAction(null)}
            >
              Cancelar
            </button>
            <button
              type="button"
              className={`${styles.primaryAction}`}
              onClick={handleConfirmBulk}
              disabled={bulkMutation.isPending}
            >
              {bulkMutation.isPending ? "Aplicando…" : "Confirmar"}
            </button>
          </>
        }
      >
        <p className={styles.confirmText}>
          Você está prestes a{" "}
          <span className={styles.confirmStrong}>
            {pendingAction ? ACTION_LABEL[pendingAction] : ""}
          </span>{" "}
          {selectedIds.length} usuário(s). Esta ação é registrada em auditoria.
        </p>
      </Dialog>
    </section>
  );
}
