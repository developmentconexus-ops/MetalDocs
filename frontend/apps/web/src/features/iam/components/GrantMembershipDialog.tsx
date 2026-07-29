import { useEffect, useId, useState } from "react";
import type { FormEvent } from "react";
import { Dialog } from "../../../components/ui/Dialog";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { IAM_AREA_ROLES } from "../constants";
import { isAreaRole, type AreaRole } from "../../../lib/iam/role-vocabulary";
import styles from "./GrantMembershipDialog.module.css";

export interface GrantUserOption {
  userId: string;
  label: string;
}

export interface GrantMembershipPayload {
  userId: string;
  areaCode: string;
  // AreaRole, not UserRole: this dialog writes a user_process_areas row, which
  // cannot hold system_admin (F-QA4-2 — the contract now says so too).
  role: AreaRole;
}

interface GrantMembershipDialogProps {
  open: boolean;
  onClose: () => void;
  users: ReadonlyArray<GrantUserOption>;
  areaSuggestions: ReadonlyArray<string>;
  onSubmit: (payload: GrantMembershipPayload) => void;
  isPending: boolean;
  error: unknown;
}

interface FormState {
  userId: string;
  areaCode: string;
  role: AreaRole;
}

const INITIAL: FormState = {
  userId: "",
  areaCode: "",
  role: "viewer",
};

export default function GrantMembershipDialog({
  open,
  onClose,
  users,
  areaSuggestions,
  onSubmit,
  isPending,
  error,
}: GrantMembershipDialogProps) {
  const [form, setForm] = useState<FormState>(INITIAL);
  const [touched, setTouched] = useState({ userId: false, areaCode: false });
  const errorId = useId();
  // Literal id (not useId): useId() emits colons which some browsers/AT
  // misparse inside the <input list="…"> ↔ <datalist id="…"> association.
  const areaListId = "grant-membership-area-list";

  useEffect(() => {
    if (!open) {
      setForm(INITIAL);
      setTouched({ userId: false, areaCode: false });
    }
  }, [open]);

  const userIdError =
    touched.userId && form.userId.trim().length === 0
      ? "Selecione um usuário."
      : null;
  const areaError =
    touched.areaCode && form.areaCode.trim().length === 0
      ? "Informe o código da área."
      : null;

  const canSubmit =
    form.userId.trim().length > 0 &&
    form.areaCode.trim().length > 0 &&
    !isPending;

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setTouched({ userId: true, areaCode: true });
    if (!canSubmit) return;
    onSubmit({
      userId: form.userId.trim(),
      areaCode: form.areaCode.trim(),
      role: form.role,
    });
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Conceder membership de área"
      description="Concede a um usuário uma função em uma área de processo. A ação é registrada em auditoria."
      footer={
        <>
          <button type="button" className={styles.cancelBtn} onClick={onClose}>
            Cancelar
          </button>
          <button
            type="submit"
            form="grant-membership-form"
            className={styles.submitBtn}
            disabled={!canSubmit}
          >
            {isPending ? "Concedendo…" : "Conceder"}
          </button>
        </>
      }
    >
      <form
        id="grant-membership-form"
        className={styles.form}
        onSubmit={handleSubmit}
        noValidate
        aria-describedby={error ? errorId : undefined}
      >
        {error ? (
          <div id={errorId} role="alert" className={styles.error}>
            {resolveQueryError(error, "Falha ao conceder membership.")}
          </div>
        ) : null}

        <div className={styles.field}>
          <label className={styles.label} htmlFor="grant-user">
            Usuário
            <span className={styles.required} aria-hidden="true">
              *
            </span>
          </label>
          <select
            id="grant-user"
            className={`${styles.select} ${userIdError ? styles.inputInvalid : ""}`}
            value={form.userId}
            onChange={(e) => setForm((f) => ({ ...f, userId: e.target.value }))}
            onBlur={() => setTouched((t) => ({ ...t, userId: true }))}
            aria-required="true"
            aria-invalid={userIdError ? true : undefined}
            aria-describedby={userIdError ? "grant-user-err" : undefined}
          >
            <option value="">Selecione um usuário…</option>
            {users.map((u) => (
              <option key={u.userId} value={u.userId}>
                {u.label}
              </option>
            ))}
          </select>
          {userIdError ? (
            <span id="grant-user-err" className={styles.fieldError}>
              {userIdError}
            </span>
          ) : null}
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="grant-area">
            Área
            <span className={styles.required} aria-hidden="true">
              *
            </span>
          </label>
          <input
            id="grant-area"
            type="text"
            list={areaListId}
            className={`${styles.input} ${areaError ? styles.inputInvalid : ""}`}
            value={form.areaCode}
            autoComplete="off"
            autoCapitalize="characters"
            spellCheck={false}
            onChange={(e) => setForm((f) => ({ ...f, areaCode: e.target.value }))}
            onBlur={() => setTouched((t) => ({ ...t, areaCode: true }))}
            aria-required="true"
            aria-invalid={areaError ? true : undefined}
            aria-describedby={areaError ? "grant-area-err" : undefined}
          />
          <datalist id={areaListId}>
            {areaSuggestions.map((code) => (
              <option key={code} value={code} />
            ))}
          </datalist>
          {areaError ? (
            <span id="grant-area-err" className={styles.fieldError}>
              {areaError}
            </span>
          ) : (
            <span className={styles.hint}>
              Código da área de processo (ex.: QUA, PRD). Selecione ou digite.
            </span>
          )}
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="grant-role">
            Função
            <span className={styles.required} aria-hidden="true">
              *
            </span>
          </label>
          <select
            id="grant-role"
            className={styles.select}
            value={form.role}
            onChange={(e) => {
              const next = e.target.value;
              if (!isAreaRole(next)) return;
              setForm((f) => ({ ...f, role: next }));
            }}
            aria-required="true"
          >
            {IAM_AREA_ROLES.map(([v, label]) => (
              <option key={v} value={v}>
                {label}
              </option>
            ))}
          </select>
        </div>
      </form>
    </Dialog>
  );
}
