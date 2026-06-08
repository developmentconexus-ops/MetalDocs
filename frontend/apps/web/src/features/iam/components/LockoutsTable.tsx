import { toast } from "sonner";
import { Avatar } from "../../../components/ui/Avatar";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useLockoutsQuery } from "../queries/useLockoutsQuery";
import { useUnlockUserMutation } from "../mutations/useUnlockUserMutation";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import { SP_DATE_TIME_FORMATTER } from "../constants";
import styles from "./LockoutsTable.module.css";

export default function LockoutsTable() {
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
    unlock.mutate(userId, {
      onSuccess: () => {
        toast.success(`${name} desbloqueado(a).`);
      },
      onError: (err) => {
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
            <li key={l.user_id} className={styles.row}>
              <div className={styles.user}>
                <Avatar name={l.display_name} size="sm" />
                <span className={styles.userName}>{l.display_name}</span>
              </div>
              <span className={styles.attempts}>
                {l.failed_attempts} tent.
              </span>
              <span
                className={styles.time}
                title={
                  l.locked_until
                    ? SP_DATE_TIME_FORMATTER.format(new Date(l.locked_until))
                    : undefined
                }
              >
                {l.locked_until ? `até ${getRelativeTime(l.locked_until)}` : "—"}
              </span>
              <button
                type="button"
                className={styles.unlockBtn}
                onClick={() => handleUnlock(l.user_id, l.display_name)}
                disabled={unlock.isPending}
                aria-label={`Desbloquear ${l.display_name}`}
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
