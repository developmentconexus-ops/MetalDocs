import { useState } from "react";
import type { DocumentFamily } from "../types";
import { useCreateFamilyMutation, useUpdateFamilyMutation } from "../queries/useTaxonomyMutations";
import { resolveErrorMessage } from "../../../lib/api/problem";
import styles from "./TaxonomyDialog.module.css";

type Props = {
  mode: "create" | "edit";
  family?: DocumentFamily;
  onClose: () => void;
};

export function FamilyEditDialog({ mode, family, onClose }: Props) {
  const [code, setCode] = useState(family?.code ?? "");
  const [name, setName] = useState(family?.name ?? "");
  const [description, setDescription] = useState(family?.description ?? "");
  const [error, setError] = useState("");
  const createMutation = useCreateFamilyMutation();
  const updateMutation = useUpdateFamilyMutation();
  const saving = createMutation.isPending || updateMutation.isPending;

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      if (mode === "create") {
        await createMutation.mutateAsync({
          code: code.trim(),
          name: name.trim(),
          description: description.trim() || undefined,
        });
      } else {
        await updateMutation.mutateAsync({
          code: family!.code,
          req: {
            name: name.trim(),
            description: description.trim() || undefined,
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
          {mode === "create" ? "Nova Família Documental" : "Editar Família Documental"}
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
              <input value={family?.code ?? ""} readOnly className={styles.inputReadOnly} />
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
          <div className={styles.fieldLast}>
            <label className={styles.label}>Descrição</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              className={styles.input}
            />
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
