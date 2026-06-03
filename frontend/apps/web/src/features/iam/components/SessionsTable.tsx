import { useEffect, useMemo, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Avatar } from "../../../components/ui/Avatar";
import { Dialog } from "../../../components/ui/Dialog";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { QK } from "../../../lib/queryKeys";
import { useSessionsQuery } from "../queries/useSessionsQuery";
import { useRevokeSessionMutation } from "../mutations/useRevokeSessionMutation";
import { getDeviceLabel } from "../presenters/device-label-presenter";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import { SP_DATE_TIME_FORMATTER } from "../constants";
import styles from "./SessionsTable.module.css";

// TODO PR-12 Fase 7+: server-side paging + filter pushdown when SessionItem grows beyond bounded tenants.
const LIMIT = 200;

export default function SessionsTable() {
  const qc = useQueryClient();
  const { data, isLoading, error } = useSessionsQuery({ limit: LIMIT });
  const revoke = useRevokeSessionMutation();

  const [userFilter, setUserFilter] = useState("");
  const [ipFilter, setIpFilter] = useState("");
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [bulkPending, setBulkPending] = useState(false);
  const headerCheckboxRef = useRef<HTMLInputElement | null>(null);

  const items = data?.items ?? [];

  const filtered = useMemo(() => {
    const u = userFilter.trim().toLowerCase();
    const ip = ipFilter.trim().toLowerCase();
    return items.filter((s) => {
      if (u && !s.displayName.toLowerCase().includes(u)) return false;
      if (ip && !(s.ipAddress ?? "").toLowerCase().includes(ip)) return false;
      return true;
    });
  }, [items, userFilter, ipFilter]);

  const allSelected =
    filtered.length > 0 && filtered.every((s) => selected.has(s.sessionId));
  const someSelected = selected.size > 0 && !allSelected;

  useEffect(() => {
    if (headerCheckboxRef.current) {
      headerCheckboxRef.current.indeterminate = someSelected;
    }
  }, [someSelected]);

  if (isLoading) {
    return (
      <section className={styles.panel} aria-busy="true">
        <div className={styles.empty}>Carregando sessões…</div>
      </section>
    );
  }

  if (error) {
    return (
      <section className={styles.panel}>
        <div role="alert" className={styles.error}>
          {resolveQueryError(error, "Não foi possível carregar as sessões.")}
        </div>
      </section>
    );
  }

  const toggleAll = () => {
    if (allSelected) {
      setSelected(new Set());
    } else {
      setSelected(new Set(filtered.map((s) => s.sessionId)));
    }
  };

  const toggleOne = (id: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const optimisticDrop = (ids: ReadonlySet<string>) => {
    const key = QK.iam.sessions({ limit: LIMIT });
    const prev = qc.getQueryData<typeof data>(key);
    if (!prev) return () => {};
    qc.setQueryData(key, {
      ...prev,
      items: prev.items.filter((s) => !ids.has(s.sessionId)),
    });
    return (succeededIds?: ReadonlySet<string>) => {
      const current = qc.getQueryData<typeof data>(key);
      if (!current) {
        qc.setQueryData(key, prev);
        return;
      }
      if (!succeededIds || succeededIds.size === 0) {
        qc.setQueryData(key, prev);
        return;
      }
      // Re-insert only the ids that failed; keep succeeded ones dropped.
      const restored = prev.items.filter((s) => !succeededIds.has(s.sessionId));
      qc.setQueryData(key, { ...prev, items: restored });
    };
  };

  const handleRevoke = (sessionId: string) => {
    const rollback = optimisticDrop(new Set([sessionId]));
    revoke.mutate(sessionId, {
      onSuccess: () => {
        toast.success("Sessão revogada.");
        void qc.invalidateQueries({ queryKey: QK.iam.sessions({ limit: LIMIT }) });
      },
      onError: (err) => {
        rollback();
        toast.error(resolveQueryError(err, "Falha ao revogar sessão."));
      },
    });
  };

  const handleBulkConfirm = async () => {
    const ids = Array.from(selected);
    if (ids.length === 0) return;
    setBulkPending(true);
    const rollback = optimisticDrop(selected);
    const results = await Promise.allSettled(
      ids.map((id) => revoke.mutateAsync(id)),
    );
    setBulkPending(false);
    setConfirmOpen(false);
    const succeededIds = new Set<string>();
    let failureCount = 0;
    results.forEach((r, i) => {
      if (r.status === "fulfilled") succeededIds.add(ids[i]);
      else failureCount += 1;
    });
    if (failureCount > 0) {
      rollback(succeededIds);
      toast.error(`Falha ao revogar ${failureCount} sessão(ões).`);
      setSelected(
        (prev) => new Set(Array.from(prev).filter((id) => !succeededIds.has(id))),
      );
    } else {
      toast.success(`${ids.length} sessão(ões) revogada(s).`);
      setSelected(new Set());
    }
    void qc.invalidateQueries({ queryKey: QK.iam.sessions({ limit: LIMIT }) });
  };

  return (
    <section
      className={styles.panel}
      aria-labelledby="sessions-table-title"
      data-testid="sessions-table"
    >
      <header className={styles.header}>
        <div>
          <h3 id="sessions-table-title" className={styles.title}>
            Sessões ativas no tenant
          </h3>
          <p className={styles.subtitle}>
            Todos os logins atualmente válidos. Revogar encerra o token.
          </p>
        </div>
        <span className={styles.count} aria-live="polite">
          {filtered.length}
        </span>
      </header>

      <div className={styles.toolbar} role="search">
        <input
          type="search"
          placeholder="Filtrar por usuário"
          value={userFilter}
          onChange={(e) => setUserFilter(e.target.value)}
          aria-label="Filtrar por usuário"
        />
        <input
          type="search"
          placeholder="Filtrar por IP"
          value={ipFilter}
          onChange={(e) => setIpFilter(e.target.value)}
          aria-label="Filtrar por IP"
        />
        <input
          type="search"
          placeholder="País"
          value=""
          onChange={() => {}}
          aria-label="Filtrar por país"
          aria-disabled="true"
          disabled
          title="Indisponível nesta versão"
        />
        <select
          value="all"
          onChange={() => {}}
          aria-label="Filtrar por MFA no login"
          aria-disabled="true"
          disabled
          title="Indisponível nesta versão"
        >
          <option value="all">MFA: todos</option>
        </select>
        <button
          type="button"
          className={styles.bulkBtn}
          onClick={() => setConfirmOpen(true)}
          disabled={selected.size === 0}
        >
          Revogar selecionadas ({selected.size})
        </button>
      </div>

      {filtered.length === 0 ? (
        <div className={styles.empty}>Nenhuma sessão encontrada.</div>
      ) : (
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>
                  <input
                    ref={headerCheckboxRef}
                    type="checkbox"
                    className={styles.checkbox}
                    checked={allSelected}
                    onChange={toggleAll}
                    aria-label="Selecionar todas"
                  />
                </th>
                <th>Usuário</th>
                <th>Dispositivo</th>
                <th>IP</th>
                <th>País</th>
                <th>Iniciada</th>
                <th>Última atividade</th>
                <th>MFA no login</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {filtered.map((s) => (
                <tr key={s.sessionId}>
                  <td>
                    <input
                      type="checkbox"
                      className={styles.checkbox}
                      checked={selected.has(s.sessionId)}
                      onChange={() => toggleOne(s.sessionId)}
                      aria-label={`Selecionar sessão de ${s.displayName}`}
                    />
                  </td>
                  <td>
                    <div className={styles.userCell}>
                      <Avatar name={s.displayName} size="sm" />
                      <span className={styles.userName}>{s.displayName}</span>
                    </div>
                  </td>
                  <td className={styles.device}>
                    {s.deviceLabel ?? getDeviceLabel(s.userAgent)}
                  </td>
                  <td>
                    <span className={styles.mono}>{s.ipAddress ?? "—"}</span>
                  </td>
                  <td className={styles.mono}>—</td>
                  <td>
                    <span
                      className={styles.time}
                      title={SP_DATE_TIME_FORMATTER.format(new Date(s.createdAt))}
                    >
                      {getRelativeTime(s.createdAt)}
                    </span>
                  </td>
                  <td>
                    <span
                      className={styles.time}
                      title={SP_DATE_TIME_FORMATTER.format(new Date(s.lastSeenAt))}
                    >
                      {getRelativeTime(s.lastSeenAt)}
                    </span>
                  </td>
                  <td>
                    <span className={`${styles.pill} ${styles.pillNo}`}>—</span>
                  </td>
                  <td>
                    <button
                      type="button"
                      className={styles.revokeBtn}
                      onClick={() => handleRevoke(s.sessionId)}
                      disabled={revoke.isPending}
                      aria-label={`Revogar sessão de ${s.displayName}`}
                    >
                      Revogar
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Dialog
        open={confirmOpen}
        onClose={() => {
          if (!bulkPending) setConfirmOpen(false);
        }}
        disableBackdropClose={bulkPending}
        title="Revogar sessões selecionadas"
        description={`Você vai encerrar ${selected.size} sessão(ões). Os usuários afetados precisarão fazer login novamente.`}
        footer={
          <div className={styles.confirmActions}>
            <button
              type="button"
              className={styles.confirmCancel}
              onClick={() => setConfirmOpen(false)}
              disabled={bulkPending}
            >
              Cancelar
            </button>
            <button
              type="button"
              className={styles.confirmConfirm}
              onClick={handleBulkConfirm}
              disabled={bulkPending}
            >
              {bulkPending ? "Revogando…" : "Confirmar revogação"}
            </button>
          </div>
        }
      >
        <p>Esta ação é irreversível.</p>
      </Dialog>
    </section>
  );
}
