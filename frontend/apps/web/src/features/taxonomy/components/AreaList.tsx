import { useState } from "react";
import type { ProcessArea } from "../types";
import { useArchiveAreaMutation } from "../queries/useTaxonomyMutations";
import { AreaEditDialog } from "./AreaEditDialog";
import { resolveErrorMessage } from "../../../lib/api/problem";
import styles from "./TaxonomyList.module.css";

type Props = {
  areas: ProcessArea[];
  includeArchived: boolean;
  onToggleArchived: (value: boolean) => void;
};

export function AreaList({ areas, includeArchived, onToggleArchived }: Props) {
  const [dialogMode, setDialogMode] = useState<"create" | "edit" | null>(null);
  const [selectedArea, setSelectedArea] = useState<ProcessArea | undefined>(undefined);
  const [actionError, setActionError] = useState("");
  const archiveMutation = useArchiveAreaMutation();

  function openCreate() {
    setSelectedArea(undefined);
    setDialogMode("create");
  }

  function openEdit(area: ProcessArea) {
    setSelectedArea(area);
    setDialogMode("edit");
  }

  function closeDialog() {
    setDialogMode(null);
    setSelectedArea(undefined);
  }

  async function handleArchive(area: ProcessArea) {
    if (!window.confirm(`Arquivar área "${area.name}" (${area.code})?`)) return;
    setActionError("");
    try {
      await archiveMutation.mutateAsync(area.code);
    } catch (err) {
      setActionError(resolveErrorMessage(err));
    }
  }

  const areaByCode = new Map(areas.map((a) => [a.code, a]));

  return (
    <div>
      <div className={styles.toolbar}>
        <button type="button" onClick={openCreate} className={styles.addButton}>
          + Nova Área
        </button>
        <label className={styles.toggleLabel}>
          <input
            type="checkbox"
            checked={includeArchived}
            onChange={(e) => onToggleArchived(e.target.checked)}
          />
          Mostrar arquivadas
        </label>
      </div>

      {actionError && <p className={styles.errorState} role="alert">{actionError}</p>}

      <table className={styles.table}>
        <thead>
          <tr className={styles.headRow}>
            <th className={styles.headCell}>Código</th>
            <th className={styles.headCell}>Nome</th>
            <th className={styles.headCell}>Área pai</th>
            <th className={styles.headCell}>Role aprovador</th>
            <th className={styles.headCell}>Status</th>
            <th className={styles.headCell}>Ações</th>
          </tr>
        </thead>
        <tbody>
          {areas.length === 0 && (
            <tr>
              <td colSpan={6} className={styles.emptyState}>Nenhuma área encontrada.</td>
            </tr>
          )}
          {areas.map((area) => (
            <tr
              key={area.code}
              className={area.archivedAt ? `${styles.row} ${styles.rowArchived}` : styles.row}
            >
              <td className={styles.cellMono}>{area.code}</td>
              <td className={styles.cell}>{area.name}</td>
              <td className={styles.cellMonoMuted}>
                {area.parentCode ? (areaByCode.get(area.parentCode)?.name ?? area.parentCode) : "-"}
              </td>
              <td className={styles.cellMuted}>{area.defaultApproverRole ?? "-"}</td>
              <td className={styles.cell}>
                {area.archivedAt ? (
                  <span className={styles.statusArchived}>Arquivada</span>
                ) : (
                  <span className={styles.statusActive}>Ativa</span>
                )}
              </td>
              <td className={styles.cellActions}>
                <button type="button" onClick={() => openEdit(area)} className={styles.actionButton}>
                  Editar
                </button>
                {!area.archivedAt && (
                  <button
                    type="button"
                    onClick={() => void handleArchive(area)}
                    disabled={archiveMutation.isPending}
                    className={styles.actionButtonDanger}
                  >
                    Arquivar
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {dialogMode && (
        <AreaEditDialog
          mode={dialogMode}
          area={selectedArea}
          areas={areas}
          onClose={closeDialog}
        />
      )}
    </div>
  );
}
