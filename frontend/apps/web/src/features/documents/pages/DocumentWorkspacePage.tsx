import { Link, useParams } from 'react-router-dom';

import { CodeChip, StatusPill } from '../../../components/ui';
import { useAuthStore } from '../../../store/auth.store';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';
import { deriveWorkspaceMode } from '../../approval/lib/workspaceMode';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useApprovalInstanceQuery } from '../queries/useApprovalInstanceQuery';
import { DocumentShell } from '../components/DocumentShell';
import { ModeChip } from '../components/workspace/ModeChip';
import { WorkspaceSidebar } from '../components/workspace/WorkspaceSidebar';
import { parseDocumentStatus } from '../lib/parseDocumentStatus';
import styles from './DocumentWorkspacePage.module.css';

/**
 * F2d.5 S2a — the mode-adaptive single working screen owner.
 *
 * Constant shell (header + ModeChip + canvas + WorkspaceSidebar) across every
 * `WorkspaceMode` (F2d.3, `deriveWorkspaceMode`) — only the canvas content and
 * the sidebar's decision footer vary by mode. This slice covers the READ
 * modes only: every mode renders the same read-only DocumentShell canvas
 * (mirrors the ApprovalCockpitPage "documento" tab pattern — no writer
 * session, no autosave).
 *
 * S2b: swaps the canvas for the lazy editor on author-editing /
 * author-changes-requested and for the disclosure/signature seed on
 * approving. Marked inline with `// S2b:` — do not build those here.
 */
export function DocumentWorkspacePage() {
  const { documentId = '' } = useParams<{ documentId: string }>();
  const currentUser = useAuthStore((s) => s.user);

  const docQuery = useDocumentDetailQuery(documentId);
  const instanceQuery = useApprovalInstanceQuery(documentId);

  const doc = docQuery.data ?? null;
  const instance = instanceQuery.data ?? null;
  const viewer = instance?.viewer ?? null;

  // §6 loading — shell skeleton, no central spinner.
  if (docQuery.isLoading) {
    return (
      <div className={styles.page} data-testid="workspace-page">
        <div className={styles.body}>
          <div className={styles.rail} aria-hidden="true" />
          <div className={styles.main}>
            <div className={styles.header} data-testid="workspace-skeleton">
              <div className={styles.skeletonBlock} />
              <div className={styles.skeletonChip} />
            </div>
            <div className={styles.canvas}>
              <div className={styles.skeletonCanvas} />
            </div>
          </div>
          <div className={styles.sidebarSkeleton} aria-hidden="true" />
        </div>
      </div>
    );
  }

  // Doc-not-found — teaching-copy empty state (doc query error, or settled
  // with no data).
  if (docQuery.isError || !doc) {
    return (
      <div className={styles.page} data-testid="workspace-page">
        <div className={styles.emptyState} role="alert">
          <p className={styles.emptyStateTitle}>Não foi possível localizar este documento.</p>
          <p className={styles.emptyStateHint}>
            Verifique se o link está correto ou se você ainda tem acesso a este documento.
          </p>
        </div>
      </div>
    );
  }

  const mode = deriveWorkspaceMode(doc, instance, viewer);
  const activeStage = instance?.stages?.find((s) => s.status === 'active');
  // isFetching (not isLoading): mirrors useDocumentApprovalArtifact — a
  // manual retry after an error should still surface as "loading", which v5
  // isLoading (first-fetch-only) would miss.
  const instanceLoading = instanceQuery.isFetching && instance == null;
  // §6 instance error — the canvas stays readable; only the sidebar
  // surfaces the error (instance is optional context, not a blocking read).
  const instanceError = instanceQuery.isError ? 'Não foi possível carregar os dados de aprovação.' : null;
  const statusForPill = parseDocumentStatus(doc.status);

  return (
    <div className={styles.page} data-testid="workspace-page">
      <div className={styles.body}>
        <Link to="/documents" className={styles.rail} aria-label="Voltar para a biblioteca">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="15 18 9 12 15 6" />
          </svg>
        </Link>

        <div className={styles.main}>
          <header className={styles.header}>
            <div className={styles.headerLeft}>
              {doc.code ? <CodeChip>{doc.code}</CodeChip> : null}
              <h1 className={styles.title}>{doc.name}</h1>
              {statusForPill ? <StatusPill status={statusForPill} /> : null}
              {doc.revision_number != null ? (
                <span className={styles.revision}>{formatRevisionCode(doc.revision_number)}</span>
              ) : null}
            </div>
            <ModeChip mode={mode} viewer={viewer} />
          </header>

          <main className={styles.canvas} data-testid="workspace-canvas">
            {
              // S2a: every mode renders this read-only canvas.
              // S2b: author-editing / author-changes-requested mount the lazy
              // writable editor here instead; approving mounts the
              // disclosure/signature seed. Do not branch on `mode` yet.
            }
            {doc.current_revision_id ? (
              <DocumentShell
                documentId={documentId}
                currentRevisionId={doc.current_revision_id}
                editorMode="readonly"
                author={currentUser?.displayName ?? ''}
              />
            ) : (
              <div className={styles.canvasEmpty}>Este documento ainda não possui conteúdo para exibir.</div>
            )}
          </main>
        </div>

        <WorkspaceSidebar
          documentId={documentId}
          doc={doc}
          instance={instance}
          instanceLoading={instanceLoading}
          instanceError={instanceError}
          onRetryInstance={() => void instanceQuery.refetch()}
          activeStage={activeStage}
          onRefetchInstance={async () => {
            await instanceQuery.refetch();
          }}
        />
      </div>
    </div>
  );
}
