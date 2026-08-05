import { type FormEvent, useState } from 'react';

import { Dialog } from '../../../components/ui';
import { formatDateTime } from '../../../lib/formatDate';
import { resolveErrorMessage } from '../../../lib/api/problem';
import styles from './ExtendSlaDialog.module.css';

// datetime-local inputs require `YYYY-MM-DDTHH:mm` in the browser's local
// timezone (matching how `dueAtLocal` itself is later interpreted via
// `new Date(dueAtLocal)`). Used only to seed the input's `min` bound — the
// forward-only rule is enforced authoritatively server-side (422
// `sla_extension_not_forward`), this is the friendly first line.
function toDatetimeLocalMin(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

interface ExtendSlaDialogProps {
  open: boolean;
  /** Label of the stage whose deadline is being extended (e.g. "Revisão técnica"). */
  stageLabel: string;
  /** Current due_at (ISO), used both as the display anchor and the `min` bound on the new date. */
  currentDueAt: string;
  isSubmitting: boolean;
  onConfirm: (dueAt: string, reason: string) => Promise<void>;
  onClose: () => void;
}

/**
 * Task 10: the UI for the extend-SLA governed exception (Task 9's
 * `POST /approval/instances/{id}/extend-sla`). Modeled on
 * `DeactivateRouteDialog` — a presentational dialog that receives
 * `isSubmitting`/`onConfirm` from a caller that owns the `useExtendSLA`
 * mutation, rather than issuing the write itself (`CancelInstanceDialog`'s
 * older pattern). The reason field is mandatory here too, on purpose: this is
 * a recorded governance event, not a quiet field edit.
 */
export function ExtendSlaDialog({
  open,
  stageLabel,
  currentDueAt,
  isSubmitting,
  onConfirm,
  onClose,
}: ExtendSlaDialogProps) {
  const [dueAtLocal, setDueAtLocal] = useState('');
  const [reason, setReason] = useState('');
  const [error, setError] = useState<string | null>(null);

  const trimmedReason = reason.trim();
  const canSubmit = dueAtLocal !== '' && trimmedReason !== '' && !isSubmitting;

  const handleClose = () => {
    if (isSubmitting) return;
    setDueAtLocal('');
    setReason('');
    setError(null);
    onClose();
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    if (!canSubmit) return;

    const parsed = new Date(dueAtLocal);
    if (Number.isNaN(parsed.getTime())) {
      setError('Informe uma data válida.');
      return;
    }

    try {
      await onConfirm(parsed.toISOString(), trimmedReason);
      setDueAtLocal('');
      setReason('');
    } catch (err) {
      setError(resolveErrorMessage(err));
    }
  };

  const footer = (
    <>
      <button
        type="button"
        className={`${styles.btn} ${styles.btnSecondary}`}
        onClick={handleClose}
        disabled={isSubmitting}
      >
        Voltar
      </button>
      <button
        type="submit"
        form="extend-sla-form"
        className={`${styles.btn} ${styles.btnPrimary}`}
        disabled={!canSubmit}
      >
        {isSubmitting ? 'Prorrogando…' : 'Confirmar prorrogação'}
      </button>
    </>
  );

  return (
    <Dialog
      open={open}
      title={`Prorrogar prazo — ${stageLabel}`}
      description="O prazo da rota é o padrão; esta é uma exceção registrada para esta instância, com motivo obrigatório."
      onClose={handleClose}
      footer={footer}
      disableBackdropClose={isSubmitting}
    >
      {error ? (
        <div className={styles.errorBox} role="alert">
          {error}
        </div>
      ) : null}

      <p className={styles.currentDue}>Prazo atual: {formatDateTime(currentDueAt)}</p>

      <form id="extend-sla-form" onSubmit={handleSubmit} noValidate>
        <div className={styles.field}>
          <label className={styles.label} htmlFor="extend-sla-due-at">
            Novo prazo
          </label>
          <input
            id="extend-sla-due-at"
            type="datetime-local"
            className={styles.input}
            value={dueAtLocal}
            min={toDatetimeLocalMin(currentDueAt)}
            onChange={(event) => setDueAtLocal(event.target.value)}
            disabled={isSubmitting}
            required
          />
          <small className={styles.helpText}>Deve ser posterior ao prazo atual.</small>
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="extend-sla-reason">
            Motivo da prorrogação
          </label>
          <textarea
            id="extend-sla-reason"
            className={styles.textarea}
            value={reason}
            onChange={(event) => setReason(event.target.value)}
            rows={3}
            disabled={isSubmitting}
            required
          />
        </div>
      </form>
    </Dialog>
  );
}
