import { useCallback, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import { SearchBar } from "../../../components/ui/SearchBar";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useHasCapability } from "../hooks/useHasCapability";
import { useMembershipsQuery } from "../queries/useMembershipsQuery";
import { useUsersQuery } from "../queries/useUsersQuery";
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
import type { IamRole } from "../types";
import styles from "./MembershipsTab.module.css";

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
    areaCode: filterValue.areaCode,
    role: filterValue.role,
  });
  const usersQuery = useUsersQuery({ limit: USERS_LIMIT });
  const grantMutation = useGrantMembershipMutation();
  const revokeMutation = useRevokeMembershipMutation();

  const userLabelById = useMemo(() => {
    const map = new Map<string, string>();
    for (const u of usersQuery.data?.items ?? []) {
      map.set(u.userId, u.displayName || u.username);
    }
    return map;
  }, [usersQuery.data]);

  const rows: ReadonlyArray<MembershipRow> = useMemo(() => {
    const items = membershipsQuery.data?.items ?? [];
    return items.map((m) => ({
      userId: m.userId,
      userLabel: userLabelById.get(m.userId) ?? m.userId,
      areaCode: m.areaCode,
      role: m.role,
      effectiveFrom: m.effectiveFrom,
      grantedBy: m.grantedBy,
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
      for (const m of u.areaMemberships) set.add(m.areaCode);
    }
    return Array.from(set).sort((a, b) => a.localeCompare(b, "pt-BR"));
  }, [rows, usersQuery.data]);

  const userOptions: ReadonlyArray<GrantUserOption> = useMemo(() => {
    return (usersQuery.data?.items ?? [])
      .map((u) => ({ userId: u.userId, label: u.displayName || u.username }))
      .sort((a, b) => a.label.localeCompare(b.label, "pt-BR"));
  }, [usersQuery.data]);

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
    grantMutation.mutate(payload, {
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
