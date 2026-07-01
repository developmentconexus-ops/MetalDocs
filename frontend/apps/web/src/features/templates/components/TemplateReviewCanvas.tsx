import { MetalDocsEditor } from '@metaldocs/editor-ui';
import { useTemplateDraft } from '../hooks/useTemplateDraft';
import styles from './TemplateReviewCanvas.module.css';

interface TemplateReviewCanvasProps {
  templateId: string;
  versionNum: number;
}

/**
 * Read-only rendering of a template version's .docx content for the approval
 * screen's `main` slot — the parity equivalent of the document ReviewDocumentCanvas.
 * The reviewer/approver reads the content here before acting. Mounted by the route
 * only once a concrete versionNum is resolved (so useTemplateDraft never fetches v0).
 */
export function TemplateReviewCanvas({ templateId, versionNum }: TemplateReviewCanvasProps) {
  const draft = useTemplateDraft(templateId, versionNum);

  if (draft.loading) {
    return <div className={styles.state} role="status" aria-live="polite">Carregando conteúdo do modelo…</div>;
  }
  if (draft.error) {
    return (
      <div className={styles.state} role="alert">
        <span>{draft.error}</span>
        <button type="button" className={styles.retry} onClick={draft.refetch}>Tentar novamente</button>
      </div>
    );
  }
  return (
    <div className={styles.canvas}>
      {draft.docxError ? (
        <div className={styles.docxBanner} role="alert">
          <span>{draft.docxError}</span>
          <button type="button" className={styles.retry} onClick={draft.refetch}>Tentar novamente</button>
        </div>
      ) : null}
      <MetalDocsEditor mode="readonly" documentBuffer={draft.docxBytes ?? undefined} showRuler />
    </div>
  );
}
