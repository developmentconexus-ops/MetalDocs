import { useId } from 'react';
import { Dialog } from '../../../components/ui/Dialog';
import { resolveQueryError } from '../../../lib/api';
import type { TokenDictionaryEntry } from '../api/tokensTypes';
import styles from './TokenDeleteDialog.module.css';

export interface TokenDeleteDialogProps {
  /** The token pending deletion, or null when the dialog is closed. */
  target: TokenDictionaryEntry | null;
  onConfirm: () => void;
  onClose: () => void;
  isPending: boolean;
  error: unknown;
}

export function TokenDeleteDialog({
  target,
  onConfirm,
  onClose,
  isPending,
  error,
}: TokenDeleteDialogProps) {
  const errorId = useId();

  return (
    <Dialog
      open={target !== null}
      onClose={onClose}
      title="Excluir token"
      description={
        target ? (
          <>
            Deseja excluir o token <span className={styles.code}>{`{${target.name}}`}</span>? Esta
            ação não pode ser desfeita.
          </>
        ) : undefined
      }
      disableBackdropClose={isPending}
      footer={
        <>
          <button type="button" className={styles.cancelBtn} onClick={onClose} disabled={isPending}>
            Cancelar
          </button>
          <button
            type="button"
            className={styles.dangerBtn}
            onClick={onConfirm}
            disabled={isPending}
            aria-describedby={error ? errorId : undefined}
          >
            {isPending ? 'Excluindo…' : 'Excluir'}
          </button>
        </>
      }
    >
      {error ? (
        <div id={errorId} role="alert" className={styles.error}>
          {resolveQueryError(error, 'Falha ao excluir token.')}
        </div>
      ) : null}

      <p className={styles.note}>
        Documentos que referenciam este token deixarão de tê-lo preenchido automaticamente.
      </p>
    </Dialog>
  );
}
