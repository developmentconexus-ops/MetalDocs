import { useEffect, useId, useRef, useState } from "react";
import type { FormEvent } from "react";
import { toast } from "sonner";
import { Dialog } from "../../../components/ui/Dialog";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useInviteUserMutation } from "../mutations/useInviteUserMutation";
import { ROLE_OPTIONS } from "../constants";
import type { IamRole } from "../types";
import styles from "./InviteUserDialog.module.css";

interface InviteUserDialogProps {
  open: boolean;
  onClose: () => void;
}

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

interface FormState {
  username: string;
  email: string;
  displayName: string;
  tenantRole: IamRole;
}

const INITIAL: FormState = {
  username: "",
  email: "",
  displayName: "",
  tenantRole: "viewer",
};

export default function InviteUserDialog({ open, onClose }: InviteUserDialogProps) {
  const [form, setForm] = useState<FormState>(INITIAL);
  const [touched, setTouched] = useState({
    username: false,
    email: false,
    displayName: false,
  });
  const mutation = useInviteUserMutation();
  const errorId = useId();
  const resetRef = useRef(mutation.reset);
  resetRef.current = mutation.reset;

  useEffect(() => {
    if (!open) {
      setForm(INITIAL);
      setTouched({ username: false, email: false, displayName: false });
      resetRef.current();
    }
  }, [open]);

  const usernameError =
    touched.username && form.username.trim().length < 3
      ? "Username deve ter no mínimo 3 caracteres."
      : null;
  const emailError =
    touched.email && !EMAIL_RE.test(form.email.trim())
      ? "E-mail inválido."
      : null;
  const displayNameError =
    touched.displayName && form.displayName.trim().length < 2
      ? "Nome de exibição obrigatório."
      : null;

  const canSubmit =
    form.username.trim().length >= 3 &&
    EMAIL_RE.test(form.email.trim()) &&
    form.displayName.trim().length >= 2 &&
    !mutation.isPending;

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setTouched({ username: true, email: true, displayName: true });
    if (!canSubmit) return;
    mutation.mutate(
      {
        username: form.username.trim(),
        email: form.email.trim(),
        displayName: form.displayName.trim(),
        tenantRole: form.tenantRole,
      },
      {
        onSuccess: (data) => {
          toast.success(`Convite enviado para ${form.displayName}`, {
            description: data?.tempPassword
              ? `Senha temporária: ${data.tempPassword}`
              : "Usuário deverá trocar a senha no primeiro acesso.",
            duration: 8000,
          });
          onClose();
        },
        onError: (err) => {
          toast.error(resolveQueryError(err, "Falha ao convidar usuário."));
        },
      },
    );
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Convidar usuário"
      description="Uma senha temporária será gerada e exibida apenas uma vez. O usuário deverá trocá-la no primeiro acesso."
      footer={
        <>
          <button type="button" className={styles.cancelBtn} onClick={onClose}>
            Cancelar
          </button>
          <button
            type="submit"
            form="invite-user-form"
            className={styles.submitBtn}
            disabled={!canSubmit}
          >
            {mutation.isPending ? "Enviando…" : "Enviar convite"}
          </button>
        </>
      }
    >
      <form
        id="invite-user-form"
        className={styles.form}
        onSubmit={handleSubmit}
        noValidate
        aria-describedby={mutation.error ? errorId : undefined}
      >
        {mutation.error ? (
          <div id={errorId} role="alert" className={styles.error}>
            {resolveQueryError(mutation.error, "Falha ao convidar usuário.")}
          </div>
        ) : null}

        <div className={styles.field}>
          <label className={styles.label} htmlFor="invite-username">
            Username
            <span className={styles.required} aria-hidden="true">
              *
            </span>
          </label>
          <input
            id="invite-username"
            type="text"
            className={`${styles.input} ${usernameError ? styles.inputInvalid : ""}`}
            value={form.username}
            autoComplete="off"
            autoCapitalize="off"
            spellCheck={false}
            onChange={(e) => setForm((f) => ({ ...f, username: e.target.value }))}
            onBlur={() => setTouched((t) => ({ ...t, username: true }))}
            aria-required="true"
            aria-invalid={usernameError ? true : undefined}
            aria-describedby={usernameError ? "invite-username-err" : undefined}
          />
          {usernameError ? (
            <span id="invite-username-err" className={styles.fieldError}>
              {usernameError}
            </span>
          ) : (
            <span className={styles.hint}>Identificador único, sem espaços.</span>
          )}
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="invite-email">
            E-mail
            <span className={styles.required} aria-hidden="true">
              *
            </span>
          </label>
          <input
            id="invite-email"
            type="email"
            className={`${styles.input} ${emailError ? styles.inputInvalid : ""}`}
            value={form.email}
            autoComplete="off"
            onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
            onBlur={() => setTouched((t) => ({ ...t, email: true }))}
            aria-required="true"
            aria-invalid={emailError ? true : undefined}
            aria-describedby={emailError ? "invite-email-err" : undefined}
          />
          {emailError ? (
            <span id="invite-email-err" className={styles.fieldError}>
              {emailError}
            </span>
          ) : null}
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="invite-display-name">
            Nome de exibição
            <span className={styles.required} aria-hidden="true">
              *
            </span>
          </label>
          <input
            id="invite-display-name"
            type="text"
            className={`${styles.input} ${displayNameError ? styles.inputInvalid : ""}`}
            value={form.displayName}
            onChange={(e) =>
              setForm((f) => ({ ...f, displayName: e.target.value }))
            }
            onBlur={() => setTouched((t) => ({ ...t, displayName: true }))}
            aria-required="true"
            aria-invalid={displayNameError ? true : undefined}
            aria-describedby={displayNameError ? "invite-displayname-err" : undefined}
          />
          {displayNameError ? (
            <span id="invite-displayname-err" className={styles.fieldError}>
              {displayNameError}
            </span>
          ) : null}
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="invite-role">
            Função no tenant
            <span className={styles.required} aria-hidden="true">
              *
            </span>
          </label>
          <select
            id="invite-role"
            className={styles.select}
            value={form.tenantRole}
            onChange={(e) =>
              setForm((f) => ({ ...f, tenantRole: e.target.value as IamRole }))
            }
            aria-required="true"
          >
            {ROLE_OPTIONS.map(([v, label]) => (
              <option key={v} value={v}>
                {label}
              </option>
            ))}
          </select>
          <span className={styles.hint}>
            Áreas adicionais podem ser atribuídas após criação no drawer do usuário.
          </span>
        </div>
      </form>
    </Dialog>
  );
}
