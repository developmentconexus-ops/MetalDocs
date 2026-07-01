import { useState } from "react";
import type { ReactNode } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { Icon } from "../../../components/ui/Icon";
import { ArtifactDetailView } from "../../shared/controlled-artifact/ArtifactDetailView";
import { useTemplateArtifact } from "../adapters/useTemplateArtifact";
import { createNextVersion } from "../api/templates";
import { QK } from "../../../lib/queryKeys";
import { resolveQueryError } from "../../../lib/api";
import styles from "./TemplateDetailRoute.module.css";

/**
 * Template-specific route wrapper for the shared ArtifactDetailView. Owns the
 * manual "Criar nova versão" action (the only path to a new template version
 * since M1 removed auto-spawn on approve/publish — ADR 0052).
 *
 * No aside slot — templates have no coverage card.
 */
export function TemplateDetailRoute() {
  const { templateId = "" } = useParams<{ templateId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { model, isLoading, isError } = useTemplateArtifact(templateId);

  const [isCreating, setIsCreating] = useState(false);
  const [createError, setCreateError] = useState("");

  // ── Loading state ──────────────────────────────────────────────────────────
  if (isLoading) {
    return (
      <div className={styles.state} role="status" aria-live="polite">
        <Icon name="docs" size={24} />
        <span>Carregando modelo…</span>
      </div>
    );
  }

  // ── Error state ────────────────────────────────────────────────────────────
  if (isError) {
    return (
      <div className={styles.state} role="alert">
        <Icon name="x" size={20} />
        <span>Modelo não encontrado ou sem permissão de acesso.</span>
      </div>
    );
  }

  const handleEdit = () => navigate(`/templates/${templateId}/edit`);
  const handleReview = () => navigate(`/templates/${templateId}/approval`);

  const handleCreateVersion = async () => {
    if (isCreating) return;
    setIsCreating(true);
    setCreateError("");
    try {
      await createNextVersion(templateId);
      await queryClient.invalidateQueries({ queryKey: QK.templates.detail(templateId) });
      void queryClient.invalidateQueries({ queryKey: QK.templates.list() });
      navigate(`/templates/${templateId}/edit`);
    } catch (err) {
      setCreateError(resolveQueryError(err, "Falha ao criar nova versão."));
    } finally {
      setIsCreating(false);
    }
  };

  // ── Hero actions slot ──────────────────────────────────────────────────────
  // Detail screen is the lifecycle hub: each status routes to its natural next step.
  // draft → edit the working draft; under_review/approved → the dedicated approval
  // screen (parity with documents → /approvals/:id); published → manually spawn the
  // next version (the only revision path since M1 dropped auto-spawn — ADR 0052).
  let heroActions: ReactNode;
  switch (model.status) {
    case "under_review":
      heroActions = (
        <div className={styles.heroActions}>
          <button className="btn btn-primary btn-lg" type="button" onClick={handleReview}>
            <Icon name="check" size={15} />
            Revisar versão
          </button>
          <button className="btn" type="button" onClick={handleEdit}>
            <Icon name="eye" size={13} />
            Ver no editor
          </button>
        </div>
      );
      break;
    case "approved":
      heroActions = (
        <div className={styles.heroActions}>
          <button className="btn btn-primary btn-lg" type="button" onClick={handleReview}>
            <Icon name="upload" size={15} />
            Publicar versão
          </button>
          <button className="btn" type="button" onClick={handleEdit}>
            <Icon name="eye" size={13} />
            Ver no editor
          </button>
        </div>
      );
      break;
    case "published":
      heroActions = (
        <div className={styles.heroActions}>
          <button
            className="btn btn-primary btn-lg"
            type="button"
            onClick={() => void handleCreateVersion()}
            disabled={isCreating}
          >
            <Icon name="edit" size={15} />
            {isCreating ? "Criando…" : "Criar nova versão"}
          </button>
          <button className="btn" type="button" onClick={handleEdit}>
            <Icon name="eye" size={13} />
            Ver no editor
          </button>
        </div>
      );
      break;
    case "draft":
      heroActions = (
        <div className={styles.heroActions}>
          <button className="btn btn-primary btn-lg" type="button" onClick={handleEdit}>
            <Icon name="edit" size={15} />
            Editar modelo
          </button>
          <button
            className="btn"
            type="button"
            aria-disabled
            disabled
            title="Já existe uma versão em edição. Publique a versão atual antes de criar a próxima."
          >
            <Icon name="edit" size={13} />
            Criar nova versão
          </button>
        </div>
      );
      break;
    default:
      // obsolete / unknown: read-only entry only.
      heroActions = (
        <div className={styles.heroActions}>
          <button className="btn" type="button" onClick={handleEdit}>
            <Icon name="eye" size={13} />
            Ver no editor
          </button>
        </div>
      );
      break;
  }

  // ── Extras slot: create-version error banner ───────────────────────────────
  const extras = createError ? (
    <div className={styles.errorBanner} role="alert">
      {createError}
    </div>
  ) : null;

  return (
    <ArtifactDetailView
      model={model}
      heroActions={heroActions}
      extras={extras}
    />
  );
}

export { TemplateDetailRoute as Component };
