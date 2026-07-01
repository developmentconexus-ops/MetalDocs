import { type FormEvent, useState } from 'react';

import { resolveErrorMessage } from '../../../lib/api';
import { cancel } from '../api/approvalApi';
import styles from './CancelInstanceDialog.module.css';

interface CancelInstanceDialogProps {
  documentId: string;
  onClose: () => void;
  onSuccess: () => void;
}

/**
 * Modal for cancelling an active approval instance. Collects a required reason,
 * calls the cancel endpoint, and on success notifies the caller (which refetches
 * the instance). Mirrors SupersedePublishDialog / SignoffDialog — the dialog owns
 * its own write + error state; the route only toggles its visibility.
 *
 * Replaces the former window.prompt so the flow stays in-app and accessible.
 */
export function CancelInstanceDialog({ documentId, onClose, onSuccess }: CancelInstanceDialogProps) {
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const trimmed = reason.trim();
  const submitDisabled = submitting || trimmed === '';

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (submitDisabled) {
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      await cancel(documentId, { reason: trimmed });
      onSuccess();
      onClose();
    } catch (err) {
      setError(resolveErrorMessage(err));
      setSubmitting(false);
    }
  };

  return (
    <div className={styles.overlay}>
      <div className={styles.dialog} role='dialog' aria-modal='true' aria-label='Cancelar instância de aprovação'>
        <h2 className={styles.title}>Cancelar instância de aprovação</h2>

        {error ? (
          <div className={styles.errorBox} role='alert'>
            {error}
          </div>
        ) : null}

        <form onSubmit={handleSubmit}>
          <div className={styles.field}>
            <label className={styles.label} htmlFor='cancel-instance-reason'>
              Motivo do cancelamento
            </label>
            <textarea
              id='cancel-instance-reason'
              className={styles.textarea}
              value={reason}
              onChange={(event) => setReason(event.target.value)}
              rows={3}
              disabled={submitting}
              required
            />
          </div>

          <div className={styles.actions}>
            <button
              type='button'
              className={`${styles.btn} ${styles.btnSecondary}`}
              onClick={onClose}
              disabled={submitting}
            >
              Voltar
            </button>
            <button type='submit' className={`${styles.btn} ${styles.btnPrimary}`} disabled={submitDisabled}>
              {submitting ? 'Cancelando...' : 'Confirmar cancelamento'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
