import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Avatar } from "../../../components/ui/Avatar";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { QK } from "../../../lib/queryKeys";
import { useLockoutsQuery } from "../queries/useLockoutsQuery";
import { useUnlockUserMutation } from "../mutations/useUnlockUserMutation";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import { SP_DATE_TIME_FORMATTER } from "../constants";
import styles from "./LockoutsTable.module.css";

export default function LockoutsTable() {
  const qc = useQueryClient();
  const { data, isLoading, error } = useLockoutsQuery();
  const unlock = useUnlockUserMutation();

  if (isLoading) {
    return (
      <section className={styles.panel} aria-busy="true">
        <div className={styles.empty}>Carregando bloqueios…</div>
      </section>
    );
  }

  if (error) {
    return (
      <section className={styles.panel}>
        <div role="alert" className={styles.error}>
          {resolveQueryError(error, "Não foi possível carregar os bloqueios.")}
        </div>
      </section>
    );
  }

  const items = data?.items ?? [];

  const handleUnlock = (userId: string, name: string) => {
    const key = QK.iam.lockouts();
    const prev = qc.getQueryData<typeof data>(key);
    if (prev) {
      qc.setQueryData(key, {
        ...prev,
        items: prev.items.filter((l) => l.userId !== userId),
      });
    }
    unlock.mutate(userId, {
      onSuccess: () => {
        toast.success(`${name} desbloqueado(a).`);
        void qc.invalidateQueries({ queryKey: QK.iam.lockouts() });
      },
      onError: (err) => {
        if (prev) qc.setQueryData(key, prev);
        toast.error(resolveQueryError(err, "Falha ao desbloquear usuário."));
      },
    });
  };

  return (
    <section
      className={styles.panel}
      aria-labelledby="lockouts-title"
      data-testid="lockouts-table"
    >
      <header className={styles.header}>
        <div>
          <h3 id="lockouts-title" className={styles.title}>
            Contas bloqueadas
          </h3>
          <p className={styles.subtitle}>
            Lockouts ativos. Desbloquear libera o login imediatamente.
          </p>
        </div>
        <span className={styles.count} aria-live="polite">
          {items.length}
        </span>
      </header>

      {items.length === 0 ? (
        <div className={styles.empty}>
          Nenhum bloqueio ativo no momento.
        </div>
      ) : (
        <ul className={styles.list} role="list">
          {items.map((l) => (
            <li key={l.userId} className={styles.row}>
              <div className={styles.user}>
                <Avatar name={l.displayName} size="sm" />
                <span className={styles.userName}>{l.displayName}</span>
              </div>
              <span className={styles.attempts}>
                {l.failedAttempts} tent.
              </span>
              <span
                className={styles.time}
                title={
                  l.lockedUntil
                    ? SP_DATE_TIME_FORMATTER.format(new Date(l.lockedUntil))
                    : undefined
                }
              >
                {l.lockedUntil ? `até ${getRelativeTime(l.lockedUntil)}` : "—"}
              </span>
              <button
                type="button"
                className={styles.unlockBtn}
                onClick={() => handleUnlock(l.userId, l.displayName)}
                disabled={unlock.isPending}
                aria-label={`Desbloquear ${l.displayName}`}
              >
                Desbloquear
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
