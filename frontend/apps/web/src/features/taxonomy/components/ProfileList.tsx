import { useState } from "react";
import { Link } from "react-router-dom";
import type { DocumentProfile } from "../types";
import { useArchiveProfileMutation } from "../queries/useTaxonomyMutations";
import { useHasCapability } from "../../iam/hooks/useHasCapability";
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
  // Affordance gate only — the "Configurar rota" shortcut points at a route the
  // AppShell guard would bounce without this capability. Capabilities, never
  // roles (ADR 0022); the backend stays the sole authz enforcer.
  const canManageRoutes = useHasCapability("route.manage");

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
                {/* Readiness is orthogonal to archived/active, and the two route
                    kinds are orthogonal to each other (ADR 0086): an ACTIVE
                    profile with no active DOCUMENT route cannot receive
                    documents (create rejects 409 state.approval_route_missing),
                    and one with no active TEMPLATE route cannot receive
                    templates (POST /templates rejects 409
                    APPROVAL_ROUTE_MISSING). Either gap earns its own badge; the
                    shortcut is rendered once for both. */}
                {!profile.archivedAt &&
                  (!profile.hasActiveRoute || !profile.hasActiveTemplateRoute) && (
                  <span className={styles.readinessCell}>
                    {!profile.hasActiveRoute && (
                      <span
                        className="pill pill-review"
                        title="Sem rota de aprovação ativa"
                      >
                        INCOMPLETO
                      </span>
                    )}
                    {!profile.hasActiveTemplateRoute && (
                      <span
                        className="pill pill-review"
                        title="Sem rota de aprovação de template ativa"
                      >
                        SEM ROTA DE TEMPLATE
                      </span>
                    )}
                    {canManageRoutes && (
                      <Link to="/approval-routes" className={styles.readinessLink}>
                        Configurar rota
                      </Link>
                    )}
                  </span>
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
