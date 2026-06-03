import { useEffect, useId, useRef, useState } from "react";
import type { FormEvent } from "react";
import { toast } from "sonner";
import { Dialog } from "../../../components/ui/Dialog";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useResetPasswordMutation } from "../mutations/useResetPasswordMutation";
import styles from "./ResetPasswordDialog.module.css";

interface ResetPasswordDialogProps {
  open: boolean;
  userId: string;
  displayName: string;
  onClose: () => void;
}

const MIN_LEN = 12;

export default function ResetPasswordDialog({
  open,
  userId,
  displayName,
  onClose,
}: ResetPasswordDialogProps) {
  const [password, setPassword] = useState("");
  const [touched, setTouched] = useState(false);
  const mutation = useResetPasswordMutation();
  const errorId = useId();
  const resetRef = useRef(mutation.reset);
  resetRef.current = mutation.reset;

  useEffect(() => {
    if (!open) {
      setPassword("");
      setTouched(false);
      resetRef.current();
    }
  }, [open]);

  const trimmed = password.trim();
  const passwordError =
    touched && trimmed.length < MIN_LEN
      ? `Senha precisa ter no mínimo ${MIN_LEN} caracteres.`
      : null;
  const canSubmit = trimmed.length >= MIN_LEN && !mutation.isPending;

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setTouched(true);
    if (!canSubmit) return;
    mutation.mutate(
      { userId, newPassword: trimmed },
      {
        onSuccess: () => {
          toast.success("Senha redefinida.");
          onClose();
        },
        onError: (err) => {
          toast.error(resolveQueryError(err, "Falha ao redefinir senha."));
        },
      },
    );
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Redefinir senha"
      description={`Definir nova senha temporária para ${displayName}.`}
      footer={
        <>
          <button type="button" className={styles.cancelBtn} onClick={onClose}>
            Cancelar
          </button>
          <button
            type="submit"
            form="reset-password-form"
            className={styles.submitBtn}
            disabled={!canSubmit}
          >
            {mutation.isPending ? "Redefinindo…" : "Redefinir senha"}
          </button>
        </>
      }
    >
      <form
        id="reset-password-form"
        className={styles.form}
        onSubmit={handleSubmit}
        noValidate
        aria-describedby={mutation.error ? errorId : undefined}
      >
        {mutation.error ? (
          <div id={errorId} role="alert" className={styles.error}>
            {resolveQueryError(mutation.error, "Falha ao redefinir senha.")}
          </div>
        ) : null}

        <div className={styles.field}>
          <label className={styles.label} htmlFor="reset-password-input">
            Nova senha
            <span className={styles.required} aria-hidden="true">
              *
            </span>
          </label>
          <input
            id="reset-password-input"
            type="password"
            className={`${styles.input} ${passwordError ? styles.inputInvalid : ""}`}
            value={password}
            minLength={MIN_LEN}
            autoComplete="new-password"
            autoCapitalize="off"
            spellCheck={false}
            required
            onChange={(e) => setPassword(e.target.value)}
            onBlur={() => setTouched(true)}
            aria-required="true"
            aria-invalid={passwordError ? true : undefined}
            aria-describedby={
              passwordError ? "reset-password-err" : "reset-password-hint"
            }
          />
          {passwordError ? (
            <span id="reset-password-err" className={styles.fieldError}>
              {passwordError}
            </span>
          ) : (
            <span id="reset-password-hint" className={styles.hint}>
              Mínimo {MIN_LEN} caracteres. Usuário deve trocar no próximo acesso.
            </span>
          )}
        </div>
      </form>
    </Dialog>
  );
}
