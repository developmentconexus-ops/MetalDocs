import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { Icon } from "../../../components/ui/Icon";
import { ArtifactDetailView } from "../../shared/controlled-artifact/ArtifactDetailView";
import { useTemplateArtifact } from "../adapters/useTemplateArtifact";
import { useTemplateDetailQuery } from "../queries/useTemplateDetailQuery";
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

  const detailQuery = useTemplateDetailQuery(templateId);
  const latestStatus = detailQuery.data?.latest_version.status;

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

  // Published latest ⇒ the only path forward is a brand-new version.
  // Non-published ⇒ a working draft exists; edit it.
  const isPublished = latestStatus === "published";

  const handleEdit = () => navigate(`/templates/${templateId}/edit`);

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
  const heroActions = (
    <div className={styles.heroActions}>
      {isPublished ? (
        <>
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
        </>
      ) : (
        <>
          <button
            className="btn btn-primary btn-lg"
            type="button"
            onClick={handleEdit}
          >
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
        </>
      )}
    </div>
  );

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
