import { useState } from "react";
import type { DocumentProfile } from "../types";
import { useArchiveProfileMutation } from "../queries/useTaxonomyMutations";
import { ProfileEditDialog } from "./ProfileEditDialog";
import { resolveErrorMessage } from "../../../lib/api/problem";
import styles from "./TaxonomyList.module.css";

type Props = {
  profiles: DocumentProfile[];
  includeArchived: boolean;
  onToggleArchived: (value: boolean) => void;
};

export function ProfileList({ profiles, includeArchived, onToggleArchived }: Props) {
  const [dialogMode, setDialogMode] = useState<"create" | "edit" | null>(null);
  const [selectedProfile, setSelectedProfile] = useState<DocumentProfile | undefined>(undefined);
  const [actionError, setActionError] = useState("");
  const archiveMutation = useArchiveProfileMutation();

  function openCreate() {
    setSelectedProfile(undefined);
    setDialogMode("create");
  }

  function openEdit(profile: DocumentProfile) {
    setSelectedProfile(profile);
    setDialogMode("edit");
  }

  function closeDialog() {
    setDialogMode(null);
    setSelectedProfile(undefined);
  }

  async function handleArchive(profile: DocumentProfile) {
    if (!window.confirm(`Arquivar perfil "${profile.name}" (${profile.code})?`)) return;
    setActionError("");
    try {
      await archiveMutation.mutateAsync(profile.code);
    } catch (err) {
      setActionError(resolveErrorMessage(err));
    }
  }

  return (
    <div>
      <div className={styles.toolbar}>
        <button type="button" onClick={openCreate} className={styles.addButton}>
          + Novo Perfil
        </button>
        <label className={styles.toggleLabel}>
          <input
            type="checkbox"
            checked={includeArchived}
            onChange={(e) => onToggleArchived(e.target.checked)}
          />
          Mostrar arquivados
        </label>
      </div>

      {actionError && <p className={styles.errorState} role="alert">{actionError}</p>}

      <table className={styles.table}>
        <thead>
          <tr className={styles.headRow}>
            <th className={styles.headCell}>Código</th>
            <th className={styles.headCell}>Nome</th>
            <th className={styles.headCell}>Família</th>
            <th className={styles.headCell}>Template padrão</th>
            <th className={styles.headCell}>Status</th>
            <th className={styles.headCell}>Ações</th>
          </tr>
        </thead>
        <tbody>
          {profiles.length === 0 && (
            <tr>
              <td colSpan={6} className={styles.emptyState}>Nenhum perfil encontrado.</td>
            </tr>
          )}
          {profiles.map((profile) => (
            <tr
              key={profile.code}
              className={profile.archivedAt ? `${styles.row} ${styles.rowArchived}` : styles.row}
            >
              <td className={styles.cellMono}>{profile.code}</td>
              <td className={styles.cell}>{profile.name}</td>
              <td className={styles.cellMono}>{profile.familyCode}</td>
              <td className={styles.cellMonoMuted}>{profile.defaultTemplateVersionId ?? "-"}</td>
              <td className={styles.cell}>
                {profile.archivedAt ? (
                  <span className={styles.statusArchived}>Arquivado</span>
                ) : (
                  <span className={styles.statusActive}>Ativo</span>
                )}
              </td>
              <td className={styles.cellActions}>
                <button type="button" onClick={() => openEdit(profile)} className={styles.actionButton}>
                  Editar
                </button>
                {!profile.archivedAt && (
                  <button
                    type="button"
                    onClick={() => void handleArchive(profile)}
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
        <ProfileEditDialog
          mode={dialogMode}
          profile={selectedProfile}
          onClose={closeDialog}
        />
      )}
    </div>
  );
}
