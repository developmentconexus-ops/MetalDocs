import { useMemo } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useSessionsQuery } from "../queries/useSessionsQuery";
import { useRevokeSessionMutation } from "../mutations/useRevokeSessionMutation";
import { useBulkUsersMutation } from "../mutations/useBulkUsersMutation";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import { SP_DATE_TIME_FORMATTER } from "../constants";
import { QK } from "../../../lib/queryKeys";
import styles from "./UserSessionsTable.module.css";

interface UserSessionsTableProps {
  userId: string;
}

export default function UserSessionsTable({ userId }: UserSessionsTableProps) {
  const qc = useQueryClient();
  const { data, isLoading, error, refetch } = useSessionsQuery({ limit: 200 });
  const revoke = useRevokeSessionMutation();
  const bulkRevoke = useBulkUsersMutation();

  const userSessions = useMemo(
    () => (data?.items ?? []).filter((s) => s.userId === userId),
    [data, userId],
  );

  if (isLoading) {
    return (
      <div className={styles.empty} aria-busy="true">
        Carregando sessões…
      </div>
    );
  }

  if (error) {
    return (
      <div role="alert" className={styles.error}>
        {resolveQueryError(error, "Não foi possível carregar as sessões.")}
      </div>
    );
  }

  const handleRevoke = (sessionId: string) => {
    const previousList = data?.items ?? [];
    // Optimistic: drop locally first.
    qc.setQueryData(QK.iam.sessions({ limit: 200 }), {
      ...data,
      items: previousList.filter((s) => s.sessionId !== sessionId),
    });
    revoke.mutate(sessionId, {
      onError: (err) => {
        qc.setQueryData(QK.iam.sessions({ limit: 200 }), data);
        toast.error(resolveQueryError(err, "Falha ao revogar sessão."));
      },
      onSuccess: () => {
        toast.success("Sessão revogada.");
      },
    });
  };

  const handleRevokeAll = () => {
    if (userSessions.length === 0) return;
    bulkRevoke.mutate(
      { userIds: [userId], action: "force-logout" },
      {
        onSuccess: () => {
          toast.success("Todas as sessões do usuário foram encerradas.");
          refetch();
        },
        onError: (err) => {
          toast.error(resolveQueryError(err, "Falha ao encerrar sessões."));
        },
      },
    );
  };

  return (
    <div className={styles.wrapper}>
      <div className={styles.toolbar}>
        <span className={styles.count}>
          {userSessions.length} sessão(ões) ativa(s)
        </span>
        <button
          type="button"
          className={styles.revokeAllBtn}
          onClick={handleRevokeAll}
          disabled={userSessions.length === 0 || bulkRevoke.isPending}
        >
          {bulkRevoke.isPending ? "Encerrando…" : "Encerrar todas"}
        </button>
      </div>

      {userSessions.length === 0 ? (
        <div className={styles.empty}>
          Nenhuma sessão ativa no momento.
        </div>
      ) : (
        <table className={styles.table}>
          <thead>
            <tr>
              <th>Dispositivo</th>
              <th>IP</th>
              <th>Iniciada</th>
              <th>Última atividade</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {userSessions.map((s) => (
              <tr key={s.sessionId}>
                <td className={styles.device}>
                  {s.deviceLabel ?? s.userAgent ?? "Dispositivo desconhecido"}
                </td>
                <td>
                  <span className={styles.ip}>{s.ipAddress ?? "—"}</span>
                </td>
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
                  <button
                    type="button"
                    className={styles.revokeBtn}
                    onClick={() => handleRevoke(s.sessionId)}
                    disabled={revoke.isPending}
                    aria-label="Revogar esta sessão"
                  >
                    Revogar
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
