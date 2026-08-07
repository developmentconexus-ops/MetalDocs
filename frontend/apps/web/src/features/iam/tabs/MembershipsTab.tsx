import { useCallback, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import { SearchBar } from "../../../components/ui/SearchBar";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useHasCapability } from "../../../lib/iam/useHasCapability";
import { useAuthStore } from "../../../store/auth.store";
import { useMembershipsQuery } from "../queries/useMembershipsQuery";
import { useUsersQuery } from "../../../lib/iam/useUsersQuery";
import { useGrantMembershipMutation } from "../mutations/useGrantMembershipMutation";
import { useRevokeMembershipMutation } from "../mutations/useRevokeMembershipMutation";
import MembershipsDirectory, {
  type MembershipRow,
  type MembershipSortKey,
  type MembershipSortDir,
} from "../components/MembershipsDirectory";
import MembershipsFilterBar, {
  type MembershipsFilterValue,
} from "../components/MembershipsFilterBar";
import GrantMembershipDialog, {
  type GrantMembershipPayload,
  type GrantUserOption,
} from "../components/GrantMembershipDialog";
import RevokeMembershipDialog from "../components/RevokeMembershipDialog";
import MembershipKpiStrip from "../components/MembershipKpiStrip";
import { AREA_ROLES, type AreaRole } from "../../../lib/iam/role-vocabulary";
import styles from "./MembershipsTab.module.css";

// This tab filters AREA memberships, so the accepted ?role= values are the
// area-assignable seven — matching both the filter dropdown (IAM_AREA_ROLES)
// and the user_process_areas role CHECK. Deriving it also keeps a hand-typed
// URL from selecting a role no membership row can ever have.
const VALID_ROLE: ReadonlyArray<AreaRole> = AREA_ROLES;
const VALID_SORT: ReadonlyArray<MembershipSortKey> = [
  "userLabel",
  "areaCode",
  "role",
  "effectiveFrom",
];
const VALID_DIR: ReadonlyArray<MembershipSortDir> = ["asc", "desc"];

function readEnum<T extends string>(
  raw: string | null,
  allowed: ReadonlyArray<T>,
): T | undefined {
  if (!raw) return undefined;
  return allowed.includes(raw as T) ? (raw as T) : undefined;
}

const USERS_LIMIT = 200;

export default function MembershipsTab() {
  const [searchParams, setSearchParams] = useSearchParams();
  const canManage = useHasCapability("membership.manage");
  const currentUserId = useAuthStore((s) => s.user?.userId);

  const filterValue: MembershipsFilterValue = useMemo(
    () => ({
      areaCode: searchParams.get("areaCode") ?? undefined,
      role: readEnum(searchParams.get("role"), VALID_ROLE),
    }),
    [searchParams],
  );
  const q = searchParams.get("q") ?? "";
  const sortKey = readEnum(searchParams.get("sort"), VALID_SORT) ?? "userLabel";
  const sortDir = readEnum(searchParams.get("dir"), VALID_DIR) ?? "asc";

  const [grantOpen, setGrantOpen] = useState(false);
  const [revokeTarget, setRevokeTarget] = useState<MembershipRow | null>(null);

  const membershipsQuery = useMembershipsQuery({
    area_code: filterValue.areaCode,
    role: filterValue.role,
  });
  const usersQuery = useUsersQuery({ limit: USERS_LIMIT });
  const grantMutation = useGrantMembershipMutation();
  const revokeMutation = useRevokeMembershipMutation();

  const userLabelById = useMemo(() => {
    const map = new Map<string, string>();
    for (const u of usersQuery.data?.items ?? []) {
      map.set(u.user_id, u.display_name || u.username);
    }
    return map;
  }, [usersQuery.data]);

  const rows: ReadonlyArray<MembershipRow> = useMemo(() => {
    const items = membershipsQuery.data?.items ?? [];
    return items.map((m) => ({
      userId: m.user_id,
      userLabel: userLabelById.get(m.user_id) ?? m.user_id,
      areaCode: m.area_code,
      role: m.role,
      effectiveFrom: m.effective_from,
      grantedBy: m.granted_by,
    }));
  }, [membershipsQuery.data, userLabelById]);

  const filteredRows = useMemo(() => {
    const needle = q.trim().toLowerCase();
    if (!needle) return rows;
    return rows.filter(
      (r) =>
        r.userLabel.toLowerCase().includes(needle) ||
        r.areaCode.toLowerCase().includes(needle),
    );
  }, [rows, q]);

  const areaCodes = useMemo(() => {
    const set = new Set<string>();
    for (const r of rows) set.add(r.areaCode);
    for (const u of usersQuery.data?.items ?? []) {
      for (const m of u.area_memberships ?? []) set.add(m.area_code);
    }
    return Array.from(set).sort((a, b) => a.localeCompare(b, "pt-BR"));
  }, [rows, usersQuery.data]);

  const userOptions: ReadonlyArray<GrantUserOption> = useMemo(() => {
    // Self-grant is blocked server-side (403); drop the current user from the
    // grant dropdown so the action never offers a guaranteed-failing target.
    return (usersQuery.data?.items ?? [])
      .filter((u) => u.user_id !== currentUserId)
      .map((u) => ({ userId: u.user_id, label: u.display_name || u.username }))
      .sort((a, b) => a.label.localeCompare(b.label, "pt-BR"));
  }, [usersQuery.data, currentUserId]);

  const kpis = useMemo(() => {
    const areas = new Set(rows.map((r) => r.areaCode));
    const users = new Set(rows.map((r) => r.userId));
    return {
      totalActive: rows.length,
      distinctAreas: areas.size,
      distinctUsers: users.size,
    };
  }, [rows]);

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
    (next: MembershipsFilterValue) => {
      const params = new URLSearchParams(searchParams);
      const apply = (key: string, val: string | undefined) => {
        if (val) params.set(key, val);
        else params.delete(key);
      };
      apply("areaCode", next.areaCode);
      apply("role", next.role);
      setSearchParams(params, { replace: true });
    },
    [searchParams, setSearchParams],
  );

  const handleSearchChange = useCallback(
    (value: string) => updateParam("q", value || undefined),
    [updateParam],
  );

  const handleSort = useCallback(
    (key: MembershipSortKey) => {
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

  // Mutation objects are not referentially stable, so memoizing these handlers
  // buys nothing — keep them as plain functions (mirrors PeopleTab). Each close
  // path resets the mutation so a stale error banner never greets the next open.
  const closeGrant = () => {
    setGrantOpen(false);
    grantMutation.reset();
  };

  const closeRevoke = () => {
    setRevokeTarget(null);
    revokeMutation.reset();
  };

  const handleGrantSubmit = (payload: GrantMembershipPayload) => {
    grantMutation.mutate({ user_id: payload.userId, area_code: payload.areaCode, role: payload.role }, {
      onSuccess: () => {
        toast.success(`Membership concedida em ${payload.areaCode}.`);
        closeGrant();
      },
      onError: (err) => {
        toast.error(resolveQueryError(err, "Falha ao conceder membership."));
      },
    });
  };

  const handleRevokeConfirm = () => {
    if (!revokeTarget) return;
    const target = revokeTarget;
    revokeMutation.mutate(
      { userId: target.userId, areaCode: target.areaCode },
      {
        onSuccess: () => {
          toast.success(
            `Acesso de ${target.userLabel} à área ${target.areaCode} revogado.`,
          );
          closeRevoke();
        },
        onError: (err) => {
          toast.error(resolveQueryError(err, "Falha ao revogar membership."));
        },
      },
    );
  };

  return (
    <section
      className={styles.tab}
      data-testid="admin-memberships-tab"
      aria-labelledby="admin-memberships-heading"
    >
      <div className={styles.headerRow}>
        <div className={styles.titleBlock}>
          <h2 id="admin-memberships-heading">Memberships de área</h2>
          <p className={styles.lede}>
            Acessos de área concedidos a usuários do workspace. Conceda e revogue
            funções por área de processo.
          </p>
        </div>
        {canManage ? (
          <button
            type="button"
            className={styles.primaryAction}
            onClick={() => {
              grantMutation.reset();
              setGrantOpen(true);
            }}
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              aria-hidden="true"
            >
              <path d="M12 5v14M5 12h14" />
            </svg>
            Conceder
          </button>
        ) : null}
      </div>

      <MembershipKpiStrip
        totalActive={kpis.totalActive}
        distinctAreas={kpis.distinctAreas}
        distinctUsers={kpis.distinctUsers}
        isLoading={membershipsQuery.isLoading}
      />

      <div className={styles.toolbar}>
        <div className={styles.searchBox}>
          <SearchBar
            value={q}
            onChange={handleSearchChange}
            placeholder="Buscar por usuário ou área"
            ariaLabel="Buscar memberships"
          />
        </div>
        <MembershipsFilterBar
          value={filterValue}
          onChange={handleFilterChange}
          areaCodes={areaCodes}
        />
      </div>

      <MembershipsDirectory
        memberships={filteredRows}
        isLoading={membershipsQuery.isLoading}
        error={membershipsQuery.error}
        onRetry={() => membershipsQuery.refetch()}
        sortKey={sortKey}
        sortDir={sortDir}
        onSort={handleSort}
        canManage={canManage}
        onRevoke={(row) => {
          revokeMutation.reset();
          setRevokeTarget(row);
        }}
      />

      {canManage ? (
        <>
          <GrantMembershipDialog
            open={grantOpen}
            onClose={closeGrant}
            users={userOptions}
            areaSuggestions={areaCodes}
            onSubmit={handleGrantSubmit}
            isPending={grantMutation.isPending}
            error={grantMutation.error}
          />
          <RevokeMembershipDialog
            target={revokeTarget}
            onConfirm={handleRevokeConfirm}
            onClose={closeRevoke}
            isPending={revokeMutation.isPending}
            error={revokeMutation.error}
          />
        </>
      ) : null}
    </section>
  );
}
