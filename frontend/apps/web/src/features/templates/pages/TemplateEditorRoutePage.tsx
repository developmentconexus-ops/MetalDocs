import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { TemplateEditorPage } from "./TemplateEditorPage";
import { getTemplate } from "../api/templates";
import styles from "./styles/TemplateEditorPage.module.css";

// ADR 0013 coherence: the URL carries only the templateId (mirrors
// `documents/:documentId/edit`). version_number is internal and never shown to
// humans. We resolve the working version from the template's latest_version here
// and pass the concrete number to the editor, which still keys its API calls on
// version_number internally — the same identity documents keeps in its storage
// hash. The user only ever sees the REV{nn} chip.
export function Component() {
  const navigate = useNavigate();
  const { templateId } = useParams();

  // null = still resolving the working version; a number once known.
  const [versionNum, setVersionNum] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [tick, setTick] = useState(0);
  const retry = useCallback(() => setTick((t) => t + 1), []);

  useEffect(() => {
    if (!templateId) return;
    let cancelled = false;
    setVersionNum(null);
    setError(null);
    getTemplate(templateId)
      .then((r) => {
        if (!cancelled) setVersionNum(r.template.latest_version.number);
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e));
      });
    return () => {
      cancelled = true;
    };
  }, [templateId, tick]);

  if (!templateId) {
    return null;
  }

  if (error) {
    return (
      <div role="alert" className={styles.error}>
        <span>{error}</span>
        <button type="button" className={styles.retryBtn} onClick={retry}>
          Tentar novamente
        </button>
      </div>
    );
  }

  if (versionNum == null) {
    return <div className={styles.loading}>Carregando template...</div>;
  }

  return (
    <TemplateEditorPage
      templateId={templateId}
      versionNum={versionNum}
      // A new working draft spawned by approve/publish keeps the same /edit URL —
      // we just re-point the editor at the new version_number internally.
      onNavigateToVersion={(_nextTemplateId, nextVersionNum) => setVersionNum(nextVersionNum)}
      onBack={() => navigate("/templates")}
    />
  );
}
