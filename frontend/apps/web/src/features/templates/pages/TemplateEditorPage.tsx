import * as React from 'react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { MetalDocsEditor, type MetalDocsEditorRef } from '@metaldocs/editor-ui';
import { type TemplateSchemas, type VersionDTO, submitForReview } from '../api/templates';
import { useTemplateDraft } from '../hooks/useTemplateDraft';
import { useTemplateAutosave } from '../hooks/useTemplateAutosave';
import { useTemplateSchemas } from '../hooks/useTemplateSchemas';
import { AvailableTokensPanel } from '../AvailableTokensPanel';
import { useTemplateArtifact } from '../adapters/useTemplateArtifact';
import { ArtifactMetaSidebar } from '../../shared/controlled-artifact/ArtifactMetaSidebar';
import { partitionDetected } from '../lib/tokens';
import { canSubmit, type ActorContext } from '../lib/canActOnVersion';
import { useAuthStore } from '../../../store/auth.store';
import { useTokenCatalog } from '../tokens/useTokenCatalog';
import {
  EditorChrome,
  editorChromeStyles,
  VersionBadge,
  AutosaveStatus,
  type AutosaveState,
} from '../../shared/components/editor-chrome';
import { StatusPill, type DocumentStatus } from '../../../components/ui';
import { resolveQueryError } from '../../../lib/api';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';
import styles from './styles/TemplateEditorPage.module.css';

export type TemplateEditorPageProps = {
  templateId: string;
  versionNum: number;
  onNavigateToVersion?: (templateId: string, versionNum: number) => void;
  onBack?: () => void;
};

type LeftPanel = 'variables' | null;

const VARIABLE_SYNC_DEBOUNCE_MS = 400;

const AUTOSAVE_LABELS_PT = {
  idle: '',
  saving: 'Salvando...',
  saved: 'Salvo',
  error: 'Falha ao salvar',
};

export function TemplateEditorPage({
  templateId,
  versionNum,
  onNavigateToVersion: onNavigateToVersion,
  onBack,
}: TemplateEditorPageProps) {
  const draft = useTemplateDraft(templateId, versionNum);
  const autosave = useTemplateAutosave(templateId, versionNum);
  const schemaState = useTemplateSchemas(templateId, versionNum, draft.version);
  const editorRef = useRef<MetalDocsEditorRef>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const schemaSnapshotRef = useRef<string | null>(null);
  const variableSyncTimerRef = useRef<number | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [submitMsg, setSubmitMsg] = useState<{ kind: 'success' | 'error'; text: string } | null>(null);
  const [importing, setImporting] = useState(false);
  const [importErr, setImportErr] = useState<string | null>(null);
  const [liveVersion, setLiveVersion] = useState<VersionDTO | null>(null);
  const [leftActive, setLeftActive] = useState<LeftPanel>('variables');
  const [localSchemas, setLocalSchemas] = useState<TemplateSchemas | null>(null);
  const { tokens: catalog, computedFailed: catalogFailed } = useTokenCatalog();
  const [usedKeys, setUsedKeys] = useState<Set<string>>(new Set());
  const [unknownTokens, setUnknownTokens] = useState<string[]>([]);
  const [invalidTokens, setInvalidTokens] = useState<string[]>([]);
  const catalogByKey = useMemo(() => new Map(catalog.map((c) => [c.key, c])), [catalog]);
  const { model: metaModel, isLoading: metaLoading } = useTemplateArtifact(templateId);
  const [sidebarOpen, setSidebarOpen] = useState<boolean>(
    () => localStorage.getItem('editor-sidebar-open') !== 'false',
  );

  const currentVersion = liveVersion ?? draft.version ?? null;
  const isDraft = currentVersion?.status === 'draft';

  const user = useAuthStore((s) => s.user);
  const actor: ActorContext = user
    ? { roles: user.roles ?? [], capabilities: user.capabilities ?? [] }
    : { roles: [], capabilities: [] };
  const submitGate = currentVersion ? canSubmit(currentVersion, actor) : null;

  const refreshTokenUsage = useCallback(() => {
    if (!isDraft) return;
    // No catalog (still loading, or load failed) => we cannot classify tokens.
    // Skip so we never write phantom "unknown" state for every token; the panel
    // surfaces the failure via catalogError instead.
    if (catalogFailed || catalogByKey.size === 0) return;
    const detected = editorRef.current?.getDetectedTokens() ?? [];
    const { usedKeys, unknownTokens, invalidTokens } = partitionDetected(detected, new Set(catalogByKey.keys()));
    setUsedKeys(usedKeys);
    setUnknownTokens(unknownTokens);
    setInvalidTokens(invalidTokens);
  }, [isDraft, catalogFailed, catalogByKey]);

  const handleEditorChange = useCallback(() => {
    if (variableSyncTimerRef.current) window.clearTimeout(variableSyncTimerRef.current);
    variableSyncTimerRef.current = window.setTimeout(refreshTokenUsage, VARIABLE_SYNC_DEBOUNCE_MS);
  }, [refreshTokenUsage]);

  useEffect(() => () => {
    if (variableSyncTimerRef.current) window.clearTimeout(variableSyncTimerRef.current);
  }, []);

  useEffect(() => { setLiveVersion(draft.version ?? null); }, [draft.version]);

  const editorContentReady = draft.version != null && catalog.length > 0;
  useEffect(() => {
    if (!editorContentReady) return;
    const t = window.setTimeout(() => {
      refreshTokenUsage();
    }, VARIABLE_SYNC_DEBOUNCE_MS);
    return () => window.clearTimeout(t);
  }, [editorContentReady, refreshTokenUsage]);

  useEffect(() => {
    if (!schemaState.schemas) return;
    setLocalSchemas(schemaState.schemas);
    schemaSnapshotRef.current = JSON.stringify(schemaState.schemas);
  }, [schemaState.schemas]);

  useEffect(() => {
    if (!localSchemas || !isDraft) return;
    const nextSnapshot = JSON.stringify(localSchemas);
    if (schemaSnapshotRef.current === nextSnapshot) return;
    const timer = window.setTimeout(() => {
      void schemaState.save(localSchemas).then(() => {
        schemaSnapshotRef.current = nextSnapshot;
      }).catch(() => { /* hook surfaces error state */ });
    }, 400);
    return () => window.clearTimeout(timer);
  }, [isDraft, localSchemas, schemaState]);

  // Eigenpal closes its popovers on capture-phase scroll. Swallow scroll
  // originating inside its dropdowns so font-size lists remain interactive.
  useEffect(() => {
    const guard = (e: Event) => {
      const t = e.target as Element | null;
      if (!t || typeof t.closest !== 'function') return;
      if (t.closest('[role="listbox"]') || t.closest('[data-testid$="-dropdown"]')) {
        e.stopImmediatePropagation();
      }
    };
    window.addEventListener('scroll', guard, true);
    return () => window.removeEventListener('scroll', guard, true);
  }, []);

  async function handleSubmitForReview() {
    setSubmitMsg(null);
    setSubmitting(true);
    try {
      if (autosave.hasPending()) {
        const flushed = await autosave.flush();
        if (!flushed) {
          setSubmitMsg({ kind: 'error', text: 'Falha ao salvar o conteúdo. Tente novamente antes de submeter.' });
          return;
        }
      }
      const updated = await submitForReview(templateId, versionNum, crypto.randomUUID());
      setLiveVersion(updated);
      setSubmitMsg({ kind: 'success', text: 'Enviado para revisão.' });
    } catch (err) {
      setSubmitMsg({ kind: 'error', text: resolveQueryError(err, 'Falha ao submeter para revisão.') });
    } finally {
      setSubmitting(false);
    }
  }

  async function handleImportFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    e.target.value = '';
    setImporting(true);
    setImportErr(null);
    try {
      await autosave.importDocx(await file.arrayBuffer());
      draft.refetch();
    } catch (err) {
      setImportErr(resolveQueryError(err, 'Falha ao importar arquivo .docx.'));
    } finally {
      setImporting(false);
    }
  }

  // Spinner ONLY while the version itself is loading. schemaState.loading is
  // derived synchronously from the version (it is just `version == null`), so
  // ORing it here is what let a fatal draft.error spin forever — the error
  // branch below was unreachable. Once the draft resolves, fall through to the
  // error checks, then render.
  if (draft.loading) {
    return <div className={styles.loading}>Carregando template...</div>;
  }
  if (draft.error) {
    return (
      <div role="alert" className={styles.error}>
        <span>{draft.error}</span>
        <button type="button" className={styles.retryBtn} onClick={draft.refetch}>
          Tentar novamente
        </button>
      </div>
    );
  }
  if (schemaState.error && !localSchemas) {
    return <div role="alert" className={styles.error}>{schemaState.error}</div>;
  }

  // ADR 0013: render backend-canonical REV{nn} from the version's revision_number,
  // mirroring documents. No FE lifecycle-counter math.
  const revisionBadge =
    currentVersion?.revision_number != null
      ? formatRevisionCode(currentVersion.revision_number)
      : null;

  const versionStatus: DocumentStatus | null = (() => {
    const s = currentVersion?.status;
    if (!s) return null;
    if (s === 'draft' || s === 'under_review' || s === 'approved' || s === 'published') return s;
    return null;
  })();

  const autosaveState: AutosaveState =
    autosave.status === 'saving' ? 'saving' :
    autosave.status === 'dirty' ? 'dirty' :
    autosave.status === 'error' ? 'error' :
    autosave.status === 'saved' ? 'saved' :
    'idle';

  const togglePanel = (key: 'variables') =>
    setLeftActive((prev) => (prev === key ? null : key));

  return (
    <div className={styles.page} data-editor-root>
      <div className={styles.body}>
        <aside className={styles.rail}>
          {onBack && (
            <>
              <button
                type="button"
                className={styles.railBackBtn}
                onClick={onBack}
                aria-label="Voltar"
              >
                {ICON_CHEVRON_LEFT}
                <span className={styles.railTip}>Voltar</span>
              </button>
              <div className={styles.railDivider} />
            </>
          )}
          <button
            type="button"
            aria-label="Variáveis"
            aria-pressed={leftActive === 'variables'}
            className={`${styles.railBtn} ${leftActive === 'variables' ? styles.isActive : ''}`}
            onClick={() => togglePanel('variables')}
          >
            {ICON_BRACES}
            <span className={styles.railTip}>Variáveis</span>
          </button>
        </aside>

        {leftActive === 'variables' && (
          <AvailableTokensPanel
            catalog={catalog}
            catalogError={catalogFailed}
            usedKeys={usedKeys}
            unknownTokens={unknownTokens}
            invalidTokens={invalidTokens}
            onInsert={(key) => editorRef.current?.insertToken(key)}
          />
        )}
        <main className={styles.canvas}>
          <EditorChrome
            center={
              <>
                <span className={editorChromeStyles.docTitle}>
                  {draft.template?.name ?? 'Template sem nome'}
                </span>
                <span className={editorChromeStyles.docSep}>·</span>
                <span className={editorChromeStyles.docMeta}>Template</span>
                {revisionBadge && <VersionBadge>{revisionBadge}</VersionBadge>}
                {versionStatus && <StatusPill status={versionStatus} />}
              </>
            }
            right={
              <>
                <AutosaveStatus status={autosaveState} labels={AUTOSAVE_LABELS_PT} />
                {isDraft && (
                  <>
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept=".docx,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                      style={{ display: 'none' }}
                      onChange={handleImportFile}
                    />
                    <button
                      type="button"
                      className={editorChromeStyles.primaryBtn}
                      onClick={() => fileInputRef.current?.click()}
                      disabled={importing}
                    >
                      {importing ? 'Importando...' : 'Importar .docx'}
                    </button>
                    <button
                      type="button"
                      className={editorChromeStyles.primaryBtn}
                      onClick={() => void handleSubmitForReview()}
                      disabled={submitting || !(submitGate?.allowed ?? false)}
                      title={submitGate && !submitGate.allowed ? submitGate.reason : undefined}
                    >
                      {submitting ? 'Enviando...' : 'Submeter para revisão'}
                    </button>
                  </>
                )}
              </>
            }
            alert={
              schemaState.staleConflict ? (
                <div role="alert" className={styles.alertError}>
                  <span>Outro usuário editou os metadados. Recarregue para continuar.</span>
                  <button type="button" className={styles.retryBtn} onClick={draft.refetch}>
                    Recarregar
                  </button>
                </div>
              ) : submitMsg ? (
                <div
                  role={submitMsg.kind === 'error' ? 'alert' : 'status'}
                  className={submitMsg.kind === 'error' ? styles.alertError : styles.alertSuccess}
                >
                  {submitMsg.text}
                </div>
              ) : importErr ? (
                <div role="alert" className={styles.alertError}>{importErr}</div>
              ) : draft.docxError ? (
                <div role="alert" className={styles.alertError}>
                  <span>{draft.docxError}</span>
                  <button type="button" className={styles.retryBtn} onClick={draft.refetch}>
                    Tentar novamente
                  </button>
                </div>
              ) : undefined
            }
          >
            <MetalDocsEditor
              ref={editorRef}
              mode={isDraft ? 'template-draft' : 'readonly'}
              documentBuffer={draft.docxBytes ?? undefined}
              onAutoSave={async (buf) => { autosave.queueDocx(buf); }}
              onChange={handleEditorChange}
              showRuler
            />
          </EditorChrome>
        </main>
        <ArtifactMetaSidebar
          open={sidebarOpen}
          onToggle={() => {
            setSidebarOpen((prev) => {
              const next = !prev;
              localStorage.setItem('editor-sidebar-open', String(next));
              return next;
            });
          }}
          loading={metaLoading}
          code={metaModel.code}
          meta={metaModel.meta}
          approvalChain={null}
          lineage={metaModel.lineage}
        />
      </div>
    </div>
  );
}

/* Inline SVG icons (Lucide-style). Kept colocated - single-use. */

const SVG_BASE = {
  width: 18,
  height: 18,
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 1.75,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
};

const ICON_BRACES = (
  <svg {...SVG_BASE}>
    <path d="M8 3H7a2 2 0 0 0-2 2v5a2 2 0 0 1-2 2 2 2 0 0 1 2 2v5a2 2 0 0 0 2 2h1" />
    <path d="M16 21h1a2 2 0 0 0 2-2v-5a2 2 0 0 1 2-2 2 2 0 0 1-2-2V5a2 2 0 0 0-2-2h-1" />
  </svg>
);

const ICON_CHEVRON_LEFT = (
  <svg {...SVG_BASE} width={14} height={14}>
    <path d="M15 18l-6-6 6-6" />
  </svg>
);
