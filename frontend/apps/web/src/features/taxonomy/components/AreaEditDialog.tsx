import { useState } from "react";
import type { ProcessArea } from "../types";
import { useCreateAreaMutation, useUpdateAreaMutation } from "../queries/useTaxonomyMutations";
import { resolveErrorMessage } from "../../../lib/api/problem";
import { AREA_ROLES, getRoleLabel, isAreaRole, type AreaRole } from "../../../lib/iam/role-vocabulary";
import styles from "./TaxonomyDialog.module.css";

type Props = {
  mode: "create" | "edit";
  area?: ProcessArea;
  areas: ProcessArea[];
  onClose: () => void;
};

export function AreaEditDialog({ mode, area, areas, onClose }: Props) {
  const [code, setCode] = useState(area?.code ?? "");
  const [name, setName] = useState(area?.name ?? "");
  const [description, setDescription] = useState(area?.description ?? "");
  const [parentCode, setParentCode] = useState(area?.parentCode ?? "");
  // default_approver_role is a bound vocabulary (F-QA4-2): the 7 AREA-assignable
  // roles, or "" for none. system_admin is deliberately NOT offered — it is
  // tenant-wide tier-1 and can never be an area membership, so configuring it
  // would resolve to an empty approver pool. An unrecognized stored value
  // (legacy free-text) falls back to "none" rather than pre-loading a value the
  // API would now reject with a 422.
  const [defaultApproverRole, setDefaultApproverRole] = useState<AreaRole | "">(
    isAreaRole(area?.defaultApproverRole) ? area.defaultApproverRole : "",
  );
  const [error, setError] = useState("");
  const createMutation = useCreateAreaMutation();
  const updateMutation = useUpdateAreaMutation();
  const saving = createMutation.isPending || updateMutation.isPending;

  const parentOptions = areas.filter((a) => a.code !== area?.code && !a.archivedAt);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      if (mode === "create") {
        await createMutation.mutateAsync({
          code: code.trim(),
          name: name.trim(),
          description: description.trim() || undefined,
          parentCode: parentCode.trim() || undefined,
          defaultApproverRole: defaultApproverRole || undefined,
        });
      } else {
        await updateMutation.mutateAsync({
          code: area!.code,
          req: {
            name: name.trim(),
            description: description.trim() || undefined,
            parentCode: parentCode.trim() || null,
            defaultApproverRole: defaultApproverRole || null,
          },
        });
      }
      onClose();
    } catch (err) {
      setError(resolveErrorMessage(err));
    }
  }

  return (
    <div className={styles.overlay} onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className={styles.panel}>
        <h2 className={styles.title}>
          {mode === "create" ? "Nova Área de Processo" : "Editar Área de Processo"}
        </h2>
        <form onSubmit={(e) => void handleSubmit(e)}>
          {mode === "create" ? (
            <div className={styles.field}>
              <label className={styles.label}>Código *</label>
              <input
                value={code}
                onChange={(e) => setCode(e.target.value.toLowerCase())}
                required
                className={styles.input}
              />
            </div>
          ) : (
            <div className={styles.field}>
              <label className={styles.label}>Código</label>
              <input value={area?.code ?? ""} readOnly className={styles.inputReadOnly} />
            </div>
          )}
          <div className={styles.field}>
            <label className={styles.label}>Nome *</label>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className={styles.input}
            />
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Descrição</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              className={styles.input}
            />
          </div>
          <div className={styles.field}>
            <label className={styles.label}>Área pai</label>
            <select
              value={parentCode}
              onChange={(e) => setParentCode(e.target.value)}
              className={styles.select}
            >
              <option value="">— Nenhuma —</option>
              {parentOptions.map((a) => (
                <option key={a.code} value={a.code}>{a.name} ({a.code})</option>
              ))}
            </select>
          </div>
          <div className={styles.fieldLast}>
            <label className={styles.label}>Role de aprovador padrão</label>
            <select
              value={defaultApproverRole}
              onChange={(e) => setDefaultApproverRole(e.target.value as AreaRole | "")}
              className={styles.select}
            >
              <option value="">— Nenhuma —</option>
              {AREA_ROLES.map((role) => (
                <option key={role} value={role}>{getRoleLabel(role)}</option>
              ))}
            </select>
          </div>
          {error && <p className={styles.error} role="alert">{error}</p>}
          <div className={styles.actions}>
            <button type="button" onClick={onClose} className={styles.buttonSecondary}>Cancelar</button>
            <button type="submit" disabled={saving} className={styles.buttonPrimary}>
              {saving ? "Salvando..." : "Salvar"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
