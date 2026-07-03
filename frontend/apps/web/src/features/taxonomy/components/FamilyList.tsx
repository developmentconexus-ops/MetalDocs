import { useState } from "react";
import type { DocumentFamily } from "../types";
import { useDeactivateFamilyMutation } from "../queries/useTaxonomyMutations";
import { FamilyEditDialog } from "./FamilyEditDialog";
import { resolveErrorMessage } from "../../../lib/api/problem";
import styles from "./TaxonomyList.module.css";

type Props = {
  families: DocumentFamily[];
  includeInactive: boolean;
  onToggleInactive: (value: boolean) => void;
};

export function FamilyList({ families, includeInactive, onToggleInactive }: Props) {
  const [dialogMode, setDialogMode] = useState<"create" | "edit" | null>(null);
  const [selectedFamily, setSelectedFamily] = useState<DocumentFamily | undefined>(undefined);
  const [actionError, setActionError] = useState("");
  const deactivateMutation = useDeactivateFamilyMutation();

  function openCreate() {
    setSelectedFamily(undefined);
    setDialogMode("create");
  }

  function openEdit(family: DocumentFamily) {
    setSelectedFamily(family);
    setDialogMode("edit");
  }

  function closeDialog() {
    setDialogMode(null);
    setSelectedFamily(undefined);
  }

  async function handleDeactivate(family: DocumentFamily) {
    if (!window.confirm(`Desativar família "${family.name}" (${family.code})?`)) return;
    setActionError("");
    try {
      await deactivateMutation.mutateAsync(family.code);
    } catch (err) {
      setActionError(resolveErrorMessage(err));
    }
  }

  return (
    <div>
      <div className={styles.toolbar}>
        <button type="button" onClick={openCreate} className={styles.addButton}>
          + Nova Família
        </button>
        <label className={styles.toggleLabel}>
          <input
            type="checkbox"
            checked={includeInactive}
            onChange={(e) => onToggleInactive(e.target.checked)}
          />
          Mostrar inativas
        </label>
      </div>

      {actionError && <p className={styles.errorState} role="alert">{actionError}</p>}

      <table className={styles.table}>
        <thead>
          <tr className={styles.headRow}>
            <th className={styles.headCell}>Código</th>
            <th className={styles.headCell}>Nome</th>
            <th className={styles.headCell}>Descrição</th>
            <th className={styles.headCell}>Status</th>
            <th className={styles.headCell}>Ações</th>
          </tr>
        </thead>
        <tbody>
          {families.length === 0 && (
            <tr>
              <td colSpan={5} className={styles.emptyState}>Nenhuma família encontrada.</td>
            </tr>
          )}
          {families.map((f) => (
            <tr key={f.code} className={styles.row}>
              <td className={styles.cellMono}>{f.code}</td>
              <td className={styles.cell}>{f.name}</td>
              <td className={styles.cellMuted}>{f.description || "—"}</td>
              <td className={styles.cell}>
                <span className={f.isActive ? styles.statusActive : styles.statusArchived}>
                  {f.isActive ? "Ativa" : "Inativa"}
                </span>
              </td>
              <td className={styles.cellActions}>
                <button type="button" onClick={() => openEdit(f)} className={styles.actionButton}>
                  Editar
                </button>
                {f.isActive && (
                  <button
                    type="button"
                    onClick={() => void handleDeactivate(f)}
                    disabled={deactivateMutation.isPending}
                    className={styles.actionButtonDanger}
                  >
                    Desativar
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {dialogMode && (
        <FamilyEditDialog
          mode={dialogMode}
          family={selectedFamily}
          onClose={closeDialog}
        />
      )}
    </div>
  );
}
