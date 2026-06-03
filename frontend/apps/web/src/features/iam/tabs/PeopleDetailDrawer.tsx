import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { DrawerShell } from "../../../components/ui/DrawerShell";
import { useUsersQuery } from "../queries/useUsersQuery";
import { useAuditEventsQuery } from "../queries/useAuditEventsQuery";
import { usePatchUserMutation } from "../mutations/usePatchUserMutation";
import { useUnlockUserMutation } from "../mutations/useUnlockUserMutation";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { QK } from "../../../lib/queryKeys";
import UserDetailHeader from "../components/UserDetailHeader";
import UserMembershipsTable from "../components/UserMembershipsTable";
import UserSessionsTable from "../components/UserSessionsTable";
import ResetPasswordDialog from "../components/ResetPasswordDialog";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import type { ManagedUser } from "../presenters/user-status-presenter";
import {
  ROLE_OPTIONS,
  SP_DATE_TIME_FORMATTER as titleFormatter,
} from "../constants";
import type { IamRole } from "../types";
import styles from "./PeopleDetailDrawer.module.css";

type SubTab = "profile" | "roles" | "areas" | "sessions" | "activity" | "security";

const SUB_TABS: ReadonlyArray<[SubTab, string]> = [
  ["profile", "Perfil"],
  ["roles", "Funções"],
  ["areas", "Áreas"],
  ["sessions", "Sessões"],
  ["activity", "Atividade"],
  ["security", "Segurança"],
];

// TODO(PR-13 backend): replace list-derive with GET /iam/users/{userId}.
// Sessions endpoint also lacks ?userId= — UserSessionsTable still fetches global list (limit:200) and filters client-side.
export default function PeopleDetailDrawer() {
  const { userId = "" } = useParams<{ userId: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [activeTab, setActiveTab] = useState<SubTab>("profile");

  const userQuery = useUsersQuery({ limit: 500 });

  const user: ManagedUser | undefined = useMemo(
    () => userQuery.data?.items?.find((u) => u.userId === userId),
    [userQuery.data, userId],
  );

  const handleClose = useCallback(() => {
    navigate({
      pathname: "/admin/people",
      search: searchParams.toString(),
    });
  }, [navigate, searchParams]);

  const isLoading = userQuery.isLoading;

  return (
    <DrawerShell
      open
      onClose={handleClose}
      ariaLabel="Detalhes do usuário"
      testId="admin-people-detail-drawer"
    >
      {isLoading && !user ? (
        <div className={styles.loading}>Carregando usuário…</div>
      ) : !user ? (
        <div className={styles.notFound}>
          Usuário não encontrado.{" "}
          <button
            type="button"
            className={styles.secondaryBtn}
            onClick={() => userQuery.refetch()}
          >
            Recarregar
          </button>
        </div>
      ) : (
        <>
          <UserDetailHeader user={user} />

          <div role="tablist" aria-label="Seções do usuário" className={styles.tabs}>
            {SUB_TABS.map(([key, label]) => (
              <button
                key={key}
                role="tab"
                type="button"
                className={styles.tabBtn}
                aria-selected={activeTab === key}
                aria-controls={`detail-panel-${key}`}
                id={`detail-tab-${key}`}
                tabIndex={activeTab === key ? 0 : -1}
                onClick={() => setActiveTab(key)}
              >
                {label}
              </button>
            ))}
          </div>

          <div
            className={styles.content}
            role="tabpanel"
            id={`detail-panel-${activeTab}`}
            aria-labelledby={`detail-tab-${activeTab}`}
          >
            {activeTab === "profile" ? <ProfilePanel user={user} /> : null}
            {activeTab === "roles" ? <RolesPanel user={user} /> : null}
            {activeTab === "areas" ? (
              <UserMembershipsTable userId={user.userId} />
            ) : null}
            {activeTab === "sessions" ? (
              <UserSessionsTable userId={user.userId} />
            ) : null}
            {activeTab === "activity" ? <ActivityPanel userId={user.userId} /> : null}
            {activeTab === "security" ? (
              <SecurityPanel
                user={user}
                onChanged={() => {
                  qc.invalidateQueries({ queryKey: QK.iam.users() });
                }}
              />
            ) : null}
          </div>
        </>
      )}
    </DrawerShell>
  );
}

function ProfilePanel({ user }: { user: ManagedUser }) {
  return (
    <section aria-label="Perfil">
      <h3 className={styles.sectionTitle}>Informações</h3>
      <dl className={styles.profileGrid}>
        <dt>Username</dt>
        <dd>@{user.username}</dd>
        <dt>Nome</dt>
        <dd>{user.displayName}</dd>
        <dt>E-mail</dt>
        <dd>{user.email ?? "—"}</dd>
        <dt>Criado em</dt>
        <dd>{titleFormatter.format(new Date(user.createdAt))}</dd>
        <dt>Atualizado em</dt>
        <dd>{titleFormatter.format(new Date(user.updatedAt))}</dd>
        <dt>Último acesso</dt>
        <dd>
          {user.lastLoginAt
            ? `${titleFormatter.format(new Date(user.lastLoginAt))} (${getRelativeTime(user.lastLoginAt)})`
            : "Nunca"}
        </dd>
      </dl>
    </section>
  );
}

function RolesPanel({ user }: { user: ManagedUser }) {
  const qc = useQueryClient();
  const [tenantRole, setTenantRole] = useState<IamRole>(user.tenantRole);
  const patch = usePatchUserMutation();
  const dirty = tenantRole !== user.tenantRole;

  useEffect(() => {
    setTenantRole(user.tenantRole);
  }, [user.tenantRole, user.userId]);

  const handleSave = () => {
    const previous = user.tenantRole;
    // Optimistic on cached entries.
    const snapshot = qc.getQueriesData<{ items: ManagedUser[] }>({
      queryKey: ["iam", "admin", "users"],
    });
    for (const [key, value] of snapshot) {
      if (!value?.items) continue;
      qc.setQueryData(key, {
        ...value,
        items: value.items.map((u) =>
          u.userId === user.userId ? { ...u, tenantRole } : u,
        ),
      });
    }
    patch.mutate(
      { userId: user.userId, body: { tenantRole } },
      {
        onSuccess: () => {
          toast.success("Função atualizada.");
        },
        onError: (err) => {
          // Rollback.
          for (const [key, value] of snapshot) qc.setQueryData(key, value);
          setTenantRole(previous);
          toast.error(resolveQueryError(err, "Falha ao atualizar função."));
        },
      },
    );
  };

  return (
    <section aria-label="Funções">
      <h3 className={styles.sectionTitle}>Função no tenant</h3>
      <p
        style={{
          fontSize: "var(--font-size-xs)",
          color: "var(--text-muted)",
          marginBottom: "var(--sp-2)",
        }}
      >
        Cada usuário possui apenas uma função no tenant. Funções específicas por área são gerenciadas em <strong>Áreas</strong>.
      </p>
      <div className={styles.field}>
        <label className={styles.label} htmlFor="role-select">
          Função
        </label>
        <select
          id="role-select"
          className={styles.select}
          value={tenantRole}
          onChange={(e) => setTenantRole(e.target.value as IamRole)}
        >
          {ROLE_OPTIONS.map(([v, label]) => (
            <option key={v} value={v}>
              {label} ({v})
            </option>
          ))}
        </select>
      </div>
      <div className={styles.actionsRow}>
        <button
          type="button"
          className={styles.secondaryBtn}
          onClick={() => setTenantRole(user.tenantRole)}
          disabled={!dirty || patch.isPending}
        >
          Descartar
        </button>
        <button
          type="button"
          className={styles.saveBtn}
          onClick={handleSave}
          disabled={!dirty || patch.isPending}
        >
          {patch.isPending ? "Salvando…" : "Salvar"}
        </button>
      </div>
    </section>
  );
}

function ActivityPanel({ userId }: { userId: string }) {
  const { data, isLoading, error, refetch } = useAuditEventsQuery({
    actorId: userId,
    limit: 50,
  });
  const items = data?.pages[0]?.items ?? [];

  if (isLoading) return <div className={styles.loading}>Carregando atividade…</div>;
  if (error) {
    return (
      <div role="alert" style={{ color: "var(--danger)" }}>
        {resolveQueryError(error, "Falha ao carregar atividade.")}
        <div className={styles.actionsRow}>
          <button type="button" className={styles.secondaryBtn} onClick={() => refetch()}>
            Tentar novamente
          </button>
        </div>
      </div>
    );
  }
  if (items.length === 0) {
    return (
      <div className={styles.notFound}>Nenhum evento registrado para este usuário.</div>
    );
  }
  return (
    <ul className={styles.activityList}>
      {items.map((ev) => (
        <li key={ev.id} className={styles.activityItem}>
          <span className={styles.activityAction}>{ev.action}</span>
          <span className={styles.activityMeta}>
            {ev.resourceType}
            {ev.resourceId ? ` · ${ev.resourceId}` : ""} ·{" "}
            <time
              dateTime={ev.occurredAt}
              title={titleFormatter.format(new Date(ev.occurredAt))}
            >
              {getRelativeTime(ev.occurredAt)}
            </time>
          </span>
        </li>
      ))}
    </ul>
  );
}

function SecurityPanel({
  user,
  onChanged,
}: {
  user: ManagedUser;
  onChanged: () => void;
}) {
  const unlock = useUnlockUserMutation();
  const [resetOpen, setResetOpen] = useState(false);
  const isLocked = !!user.lockedUntil && new Date(user.lockedUntil).getTime() > Date.now();

  const handleUnlock = () => {
    unlock.mutate(user.userId, {
      onSuccess: () => {
        toast.success("Usuário desbloqueado.");
        onChanged();
      },
      onError: (err) => {
        toast.error(resolveQueryError(err, "Falha ao desbloquear."));
      },
    });
  };

  return (
    <section aria-label="Segurança">
      <h3 className={styles.sectionTitle}>Segurança</h3>
      <ul className={styles.securityList} style={{ listStyle: "none", padding: 0, margin: 0 }}>
        <li className={styles.securityItem}>
          <div className={styles.securityMeta}>
            <span className={styles.securityLabel}>MFA</span>
            <span className={styles.securityValue}>
              {user.mfaEnabled
                ? "Ativo — usuário possui segundo fator configurado."
                : "Não configurado."}
            </span>
          </div>
        </li>
        <li className={styles.securityItem}>
          <div className={styles.securityMeta}>
            <span className={styles.securityLabel}>Falhas de login</span>
            <span className={styles.securityValue}>
              {user.failedLoginAttempts} tentativa(s) acumulada(s).
            </span>
          </div>
        </li>
        <li className={styles.securityItem}>
          <div className={styles.securityMeta}>
            <span className={styles.securityLabel}>Bloqueio</span>
            <span className={styles.securityValue}>
              {isLocked
                ? `Bloqueado até ${titleFormatter.format(new Date(user.lockedUntil!))}`
                : "Sem bloqueio ativo."}
            </span>
          </div>
          <button
            type="button"
            className={styles.secondaryBtn}
            onClick={handleUnlock}
            disabled={unlock.isPending || !isLocked}
          >
            {unlock.isPending ? "Desbloqueando…" : "Desbloquear"}
          </button>
        </li>
        <li className={styles.securityItem}>
          <div className={styles.securityMeta}>
            <span className={styles.securityLabel}>Senha</span>
            <span className={styles.securityValue}>
              {user.mustChangePassword
                ? "Usuário deve trocar a senha no próximo acesso."
                : "Sem obrigação de troca."}
            </span>
          </div>
          <button
            type="button"
            className={styles.dangerBtn}
            onClick={() => setResetOpen(true)}
          >
            Redefinir senha
          </button>
        </li>
      </ul>

      <ResetPasswordDialog
        open={resetOpen}
        userId={user.userId}
        displayName={user.displayName}
        onClose={() => setResetOpen(false)}
      />
    </section>
  );
}
