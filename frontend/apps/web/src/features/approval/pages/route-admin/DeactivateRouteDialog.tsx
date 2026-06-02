import { type FormEvent, useEffect, useState } from 'react';

import { Dialog } from '../../../../components/ui';
import { ApprovalError } from '../../api/mutationClient';
import type { RouteSummary } from '../../api/routeAdminApi';
import { messageForRouteError } from './routeAdminLabels';
import styles from './RouteAdmin.module.css';

interface DeactivateRouteDialogProps {
  open: boolean;
  route: RouteSummary | null;
  isSubmitting: boolean;
  onConfirm: (routeId: string, reason: string) => Promise<void>;
  onClose: () => void;
}

const MIN_REASON_LENGTH = 4;

export function DeactivateRouteDialog({
  open,
  route,
  isSubmitting,
  onConfirm,
  onClose,
}: DeactivateRouteDialogProps) {
  const [reason, setReason] = useState('');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setReason('');
      setError(null);
    }
  }, [open, route?.id]);

  const trimmed = reason.trim();
  const canSubmit = trimmed.length >= MIN_REASON_LENGTH && !isSubmitting;

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    if (!route) {
      return;
    }
    if (trimmed.length < MIN_REASON_LENGTH) {
      setError(`Informe um motivo com pelo menos ${MIN_REASON_LENGTH} caracteres.`);
      return;
    }

    try {
      await onConfirm(route.id, trimmed);
    } catch (err) {
      if (err instanceof ApprovalError) {
        setError(messageForRouteError(err.code, err.status, err.message));
      } else if (err instanceof Error) {
        setError(messageForRouteError(undefined, undefined, err.message));
      } else {
        setError(messageForRouteError(undefined, undefined));
      }
    }
  };

  if (!route) {
    return null;
  }

  const footer = (
    <>
      <button
        type="button"
        className={styles.secondaryButton}
        onClick={onClose}
        disabled={isSubmitting}
      >
        Cancelar
      </button>
      <button
        type="submit"
        form="deactivate-route-form"
        className={styles.warnButton}
        disabled={!canSubmit}
      >
        {isSubmitting ? 'Desativando…' : 'Confirmar desativação'}
      </button>
    </>
  );

  return (
    <Dialog
      open={open}
      title="Confirmar desativação"
      description={
        <>
          Deseja desativar a rota <strong>{route.name}</strong>? Esta ação é registrada como evento
          de governança e exige um motivo.
        </>
      }
      onClose={isSubmitting ? () => {} : onClose}
      footer={footer}
      disableBackdropClose={isSubmitting}
    >
      {error ? (
        <div className={styles.errorBox} role="alert">
          {error}
        </div>
      ) : null}

      <form id="deactivate-route-form" className={styles.form} onSubmit={handleSubmit} noValidate>
        <label className={styles.fieldLabel} htmlFor="deactivate-reason">
          Motivo da desativação
        </label>
        <textarea
          id="deactivate-reason"
          className={styles.textarea}
          value={reason}
          onChange={(event) => setReason(event.target.value)}
          disabled={isSubmitting}
          rows={3}
          minLength={MIN_REASON_LENGTH}
          required
          aria-describedby="deactivate-reason-help"
          autoFocus
        />
        <small id="deactivate-reason-help" className={styles.helpText}>
          O motivo é armazenado no histórico de eventos de governança da rota.
        </small>
      </form>
    </Dialog>
  );
}
