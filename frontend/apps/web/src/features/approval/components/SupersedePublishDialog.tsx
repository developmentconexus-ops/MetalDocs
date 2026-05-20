import { type FormEvent, useMemo, useState } from 'react';

import { publish, schedulePublish, supersede } from '../api/approvalApi';
import { ApprovalError } from '../api/mutationClient';
import { resolveErrorMessage } from '../../../lib/api';
import styles from './SupersedePublishDialog.module.css';

type PublishMode = 'now' | 'schedule';
type DialogState =
  | 'idle'
  | 'submitting'
  | 'success'
  | 'error_invalid_supersede_target'
  | 'error_network'
  | 'error_server';

const ERROR_MESSAGES: Record<Exclude<DialogState, 'idle' | 'submitting' | 'success'>, string> = {
  error_invalid_supersede_target:
    'A publicação agendada não pode continuar porque a versão publicada vigente mudou. Atualize a página e revise o contexto antes de tentar novamente.',
  error_network: 'Erro de conexão. Verifique sua internet e tente novamente.',
  error_server: 'Não foi possível concluir a publicação. Tente novamente.',
};

interface SupersedePublishDialogProps {
  documentId: string;
  contentHash: string;
  revisionVersion?: number;
  publishedDocumentId?: string;
  onClose: () => void;
  onSuccess: () => void;
}

function pad(value: number): string {
  return String(value).padStart(2, '0');
}

function toDateTimeLocalValue(date: Date): string {
  const year = date.getFullYear();
  const month = pad(date.getMonth() + 1);
  const day = pad(date.getDate());
  const hours = pad(date.getHours());
  const minutes = pad(date.getMinutes());
  return `${year}-${month}-${day}T${hours}:${minutes}`;
}

export function SupersedePublishDialog({
  documentId,
  contentHash,
  revisionVersion,
  publishedDocumentId,
  onClose,
  onSuccess,
}: SupersedePublishDialogProps) {
  const [mode, setMode] = useState<PublishMode>('now');
  const [scheduledAt, setScheduledAt] = useState('');
  const [state, setState] = useState<DialogState>('idle');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [showStaleBanner, setShowStaleBanner] = useState(false);

  const minScheduleDate = useMemo(() => {
    const date = new Date(Date.now() + 5 * 60 * 1000);
    return toDateTimeLocalValue(date);
  }, []);

  const isSubmitting = state === 'submitting';
  const isSuccess = state === 'success';
  const scheduleValidationMessage =
    mode === 'schedule' && scheduledAt
      ? (() => {
          const scheduledDate = new Date(scheduledAt);
          const minDateMs = Date.now() + 5 * 60 * 1000;
          if (Number.isNaN(scheduledDate.getTime()) || scheduledDate.getTime() < minDateMs) {
            return 'A data deve ser pelo menos 5 minutos no futuro.';
          }
          return null;
        })()
      : null;
  const submitDisabled = isSubmitting || (mode === 'schedule' && (!scheduledAt || scheduleValidationMessage !== null));
  const ifMatch = typeof revisionVersion === 'number' ? `"v${revisionVersion}"` : undefined;
  const hasError = state.startsWith('error_');

  const mapErrorToState = (error: unknown): DialogState => {
    if (error instanceof ApprovalError) {
      if (error.status === 412 || error.code === 'conflict.stale' || error.code === 'conflict.stale_revision') {
        setShowStaleBanner(true);
        return 'idle';
      }
      if (error.code === 'publish.invalid_supersede_target') {
        return 'error_invalid_supersede_target';
      }
      const resolvedMessage = resolveErrorMessage(error.code);
      setErrorMessage(resolvedMessage === 'Erro inesperado.' ? null : resolvedMessage);
      return 'error_server';
    }

    if (error instanceof TypeError) {
      return 'error_network';
    }

    if (error instanceof Error && /network|failed to fetch|fetch/i.test(error.message)) {
      return 'error_network';
    }

    return 'error_server';
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (isSubmitting) {
      return;
    }

    setErrorMessage(null);
    setShowStaleBanner(false);
    setState('submitting');

    try {
      if (mode === 'schedule') {
        if (!scheduledAt || scheduleValidationMessage !== null) {
          setState('idle');
          return;
        }
        const scheduledDate = new Date(scheduledAt);
        await schedulePublish(documentId, {
          effective_from: scheduledDate.toISOString(),
          superseded_document_id: publishedDocumentId,
        }, { ifMatch });
      } else if (publishedDocumentId) {
        await supersede(documentId, {
          superseded_document_id: publishedDocumentId,
        }, { ifMatch });
      } else {
        await publish(documentId, {
          content_hash: contentHash,
        }, { ifMatch });
      }

      setState('success');
      onSuccess();
      onClose();
    } catch (error) {
      setState(mapErrorToState(error));
    }
  };

  return (
    <div className={styles.overlay}>
      <div className={styles.dialog} role="dialog" aria-modal="true" aria-label="Publicação">
        <h2 className={styles.title}>Publicação</h2>

        {showStaleBanner ? (
          <div className={styles.errorBox} role="status">
            O documento foi alterado. Atualize a página antes de tentar novamente.
          </div>
        ) : null}

        {hasError ? (
          <div className={styles.errorBox} role="alert">
            {errorMessage ?? ERROR_MESSAGES[state as keyof typeof ERROR_MESSAGES]}
          </div>
        ) : null}

        {isSuccess ? (
          <div className={styles.success} role="status">
            Publicação concluída com sucesso.
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            <fieldset className={styles.fieldset}>
              <legend className={styles.legend}>Modo</legend>
              <label className={styles.radio}>
                <input
                  type="radio"
                  name="publish-mode"
                  value="now"
                  checked={mode === 'now'}
                  onChange={() => setMode('now')}
                  disabled={isSubmitting}
                />
                Publicar agora
              </label>
              <label className={styles.radio}>
                <input
                  type="radio"
                  name="publish-mode"
                  value="schedule"
                  checked={mode === 'schedule'}
                  onChange={() => setMode('schedule')}
                  disabled={isSubmitting}
                />
                Agendar publicação
              </label>
            </fieldset>

            {mode === 'schedule' ? (
              <div className={styles.field}>
                <label className={styles.label} htmlFor="publish-scheduled-at">
                  Data e hora da publicação
                </label>
                <input
                  id="publish-scheduled-at"
                  type="datetime-local"
                  className={styles.input}
                  min={minScheduleDate}
                  value={scheduledAt}
                  onChange={(event) => setScheduledAt(event.target.value)}
                  disabled={isSubmitting}
                />
                {scheduleValidationMessage ? (
                  <p className={styles.inlineError} role="alert">
                    {scheduleValidationMessage}
                  </p>
                ) : null}
                {publishedDocumentId ? (
                  <p className={styles.label}>
                    A publicação agendada substituirá automaticamente a versão publicada atual.
                  </p>
                ) : null}
              </div>
            ) : null}

            {mode === 'now' && publishedDocumentId ? (
              <p className={styles.label}>
                Esta publicação substituirá automaticamente a versão publicada atual.
              </p>
            ) : null}

            <div className={styles.actions}>
              <button
                type="button"
                className={`${styles.btn} ${styles.btnSecondary}`}
                onClick={onClose}
                disabled={isSubmitting}
              >
                Cancelar
              </button>
              <button type="submit" className={`${styles.btn} ${styles.btnPrimary}`} disabled={submitDisabled}>
                {isSubmitting ? 'Enviando...' : 'Confirmar publicação'}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
