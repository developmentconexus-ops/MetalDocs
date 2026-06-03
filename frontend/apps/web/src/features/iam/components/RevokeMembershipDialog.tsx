import { useId } from "react";
import { Dialog } from "../../../components/ui/Dialog";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { IAM_AREA_ROLES } from "../constants";
import type { MembershipRow } from "./MembershipsDirectory";
import styles from "./RevokeMembershipDialog.module.css";

interface RevokeMembershipDialogProps {
  /** The membership pending revocation, or null when the dialog is closed. */
  target: MembershipRow | null;
  onConfirm: () => void;
  onClose: () => void;
  isPending: boolean;
  error: unknown;
}

const ROLE_LABELS: Record<string, string> = Object.fromEntries(IAM_AREA_ROLES);

export default function RevokeMembershipDialog({
  target,
  onConfirm,
  onClose,
  isPending,
  error,
}: RevokeMembershipDialogProps) {
  const errorId = useId();
  const roleLabel = target ? (ROLE_LABELS[target.role] ?? target.role) : "";

  return (
    <Dialog
      open={target !== null}
      onClose={onClose}
      title="Revogar membership de área"
      description={
        target ? (
          <>
            Você está prestes a revogar o acesso de{" "}
            <span className={styles.strong}>{target.userLabel}</span> à área{" "}
            <span className={styles.code}>{target.areaCode}</span> com a função{" "}
            <span className={styles.strong}>{roleLabel}</span>.
          </>
        ) : undefined
      }
      footer={
        <>
          <button type="button" className={styles.cancelBtn} onClick={onClose}>
            Cancelar
          </button>
          <button
            type="button"
            className={styles.dangerBtn}
            onClick={onConfirm}
            disabled={isPending}
            aria-describedby={error ? errorId : undefined}
          >
            {isPending ? "Revogando…" : "Revogar acesso"}
          </button>
        </>
      }
    >
      {error ? (
        <div id={errorId} role="alert" className={styles.error}>
          {resolveQueryError(error, "Falha ao revogar membership.")}
        </div>
      ) : null}

      <p className={styles.note}>
        O usuário perde o acesso a esta área imediatamente. Esta ação é
        registrada em auditoria.
      </p>
    </Dialog>
  );
}
