import * as React from 'react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { MetalDocsEditor, type MetalDocsEditorRef } from '@metaldocs/editor-ui';
import { filterTransactionGuard } from '../../../editor-adapters/filter-transaction-guard';
import { type TemplateSchemas, type VersionDTO, submitForReview } from '../api/templates';
import { useTemplateDraft } from '../hooks/useTemplateDraft';
import { useTemplateAutosave } from '../hooks/useTemplateAutosave';
import { useTemplateSchemas } from '../hooks/useTemplateSchemas';
import { VersionActionPanel } from '../VersionActionPanel';
import { PlaceholderCatalogPanel } from '../PlaceholderCatalogPanel';
import { TemplateOutlinePanel } from '../TemplateOutlinePanel';
import { canSubmit, type ActorContext } from '../lib/canActOnVersion';
import { useAuthStore } from '../../../store/auth.store';
import { readHeadings, type Heading } from '../lib/readHeadings';
import { fetchPlaceholderCatalog, type PlaceholderCatalogEntry } from '../api/catalog';
import {
  EditorChrome,
  editorChromeStyles,
  VersionBadge,
  AutosaveStatus,
  type AutosaveState,
} from '../../shared/components/editor-chrome';
import { StatusPill, type DocumentStatus } from '../../../components/ui';
import { ApiError, resolveErrorMessage } from '../../../lib/api';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';
import styles from './styles/TemplateEditorPage.module.css';

export type TemplateEditorPageProps = {
  templateId: string;
  versionNum: number;
  onNavigateToVersion?: (templateId: string, versionNum: number) => void;
  onBack?: () => void;
};

type LeftPanel = 'variables' | 'outline' | null;

const VARIABLE_SYNC_DEBOUNCE_MS = 400;
const OUTLINE_REFRESH_DEBOUNCE_MS = 600;

const AUTOSAVE_LABELS_PT = {
  idle: '',
  saving: 'Salvando...',
  saved: 'Salvo',
  error: 'Falha ao salvar',
};

function resolveError(err: unknown, fallback: string): string {
  if (err instanceof ApiError) return resolveErrorMessage(err.code, err.message);
  if (err instanceof Error) return err.message;
  return fallback;
}

export function TemplateEditorPage({
  templateId,
  versionNum,
  onNavigateToVersion: _nav,
  onBack,
}: TemplateEditorPageProps) {
  const draft = useTemplateDraft(templateId, versionNum);
  const autosave = useTemplateAutosave(templateId, versionNum);
  const schemaState = useTemplateSchemas(templateId, versionNum);
  const editorRef = useRef<MetalDocsEditorRef>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const schemaSnapshotRef = useRef<string | null>(null);
  const variableSyncTimerRef = useRef<number | null>(null);
  const outlineSyncTimerRef = useRef<number | null>(null);
  // filterTransactionGuard is a raw ProseMirror plugin; wrap in EditorPlugin shell so
  // MetalDocsEditor can register it via PluginHost (proseMirrorPlugins field).
  const editorPlugins = useMemo(() => [{
    id: 'filter-transaction-guard',
    name: 'filter-transaction-guard',
    proseMirrorPlugins: [filterTransactionGuard()],
  }], []);

  const [submitting, setSubmitting] = useState(false);
  const [submitMsg, setSubmitMsg] = useState<{ kind: 'success' | 'error'; text: string } | null>(null);
  const [importing, setImporting] = useState(false);
  const [importErr, setImportErr] = useState<string | null>(null);
  const [liveVersion, setLiveVersion] = useState<VersionDTO | null>(null);
  const [leftActive, setLeftActive] = useState<LeftPanel>('variables');
  const [localSchemas, setLocalSchemas] = useState<TemplateSchemas | null>(null);
  const [catalog, setCatalog] = useState<PlaceholderCatalogEntry[]>([]);
  const [detectedVariables, setDetectedVariables] = useState<string[]>([]);
  const [headings, setHeadings] = useState<Heading[]>([]);

  useEffect(() => { void fetchPlaceholderCatalog().then(setCatalog); }, []);
  const catalogByKey = useMemo(() => new Map(catalog.map((c) => [c.key, c])), [catalog]);

  const currentVersion = liveVersion ?? draft.version ?? null;
  const isDraft = currentVersion?.status === 'draft';

  const user = useAuthStore((s) => s.user);
  const actor: ActorContext = user
    ? { roles: user.roles ?? [], capabilities: user.capabilities ?? [] }
    : { roles: [], capabilities: [] };
  const submitGate = currentVersion ? canSubmit(currentVersion, actor) : null;

  const syncPlaceholdersFromDocument = useCallback(() => {
    if (!isDraft) return;
    // getAgent is a deferred ACL violation - MetalDocsEditorRef does not expose it;
    // optional chaining returns undefined silently. Follow-up: forward getAgent in the adapter.
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const rawVariables: string[] = (editorRef.current as any)?.getAgent?.()?.getVariables?.() ?? [];
    const variables = Array.from(new Set(rawVariables));
    setDetectedVariables(variables.filter((name) => catalogByKey.has(name)));
    const valid = variables.filter((name) => catalogByKey.has(name));
    const placeholders = valid.map((name) => {
      const entry = catalogByKey.get(name)!;
      return {
        id: crypto.randomUUID(),
        name,
        label: entry.label,
        type: 'computed' as const,
        resolverKey: name,
      };
    });
    setLocalSchemas((prev) => (prev ? { ...prev, placeholders } : prev));
  }, [isDraft, catalogByKey]);

  const syncOutline = useCallback(() => {
    // Cast: MetalDocsEditorRef doesn't expose getAgent; readHeadings guards internally.
    setHeadings(readHeadings(editorRef as unknown as React.RefObject<{ getAgent?(): unknown }>));
  }, []);

  const handleEditorChange = useCallback(() => {
    if (variableSyncTimerRef.current) window.clearTimeout(variableSyncTimerRef.current);
    variableSyncTimerRef.current = window.setTimeout(syncPlaceholdersFromDocument, VARIABLE_SYNC_DEBOUNCE_MS);
    if (outlineSyncTimerRef.current) window.clearTimeout(outlineSyncTimerRef.current);
    outlineSyncTimerRef.current = window.setTimeout(syncOutline, OUTLINE_REFRESH_DEBOUNCE_MS);
  }, [syncPlaceholdersFromDocument, syncOutline]);

  useEffect(() => () => {
    if (variableSyncTimerRef.current) window.clearTimeout(variableSyncTimerRef.current);
    if (outlineSyncTimerRef.current) window.clearTimeout(outlineSyncTimerRef.current);
  }, []);

  useEffect(() => { setLiveVersion(draft.version ?? null); }, [draft.version]);

  // Sync detected tokens + outline once catalog + draft content are both ready.
  const editorContentReady = draft.version != null && catalog.length > 0;
  useEffect(() => {
    if (!editorContentReady) return;
    const t = window.setTimeout(() => {
      syncPlaceholdersFromDocument();
      syncOutline();
    }, OUTLINE_REFRESH_DEBOUNCE_MS);
    return () => window.clearTimeout(t);
  }, [editorContentReady, syncPlaceholdersFromDocument, syncOutline]);

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
      if (autosave.hasPending()) await autosave.flush();
      const updated = await submitForReview(templateId, versionNum);
      setLiveVersion(updated);
      setSubmitMsg({ kind: 'success', text: 'Enviado para revisão.' });
    } catch (err) {
      setSubmitMsg({ kind: 'error', text: resolveError(err, 'Falha ao submeter para revisão.') });
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
      setImportErr(resolveError(err, 'Falha ao importar arquivo .docx.'));
    } finally {
      setImporting(false);
    }
  }

  if (draft.loading || (schemaState.loading && !localSchemas)) {
    return <div className={styles.loading}>Carregando template...</div>;
  }
  if (draft.error) return <div role="alert" className={styles.error}>{draft.error}</div>;
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
    if (s === 'in_review') return 'under_review';
    if (s === 'draft' || s === 'approved' || s === 'published') return s;
    return null;
  })();

  const autosaveState: AutosaveState =
    autosave.status === 'saving' ? 'saving' :
    autosave.status === 'error' ? 'error' :
    autosave.status === 'saved' ? 'saved' :
    'idle';

  const togglePanel = (key: 'variables' | 'outline') =>
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
          <button
            type="button"
            aria-label="Estrutura"
            aria-pressed={leftActive === 'outline'}
            className={`${styles.railBtn} ${leftActive === 'outline' ? styles.isActive : ''}`}
            onClick={() => togglePanel('outline')}
          >
            {ICON_OUTLINE}
            <span className={styles.railTip}>Estrutura</span>
          </button>
        </aside>

        {leftActive === 'variables' && (
          <PlaceholderCatalogPanel detected={detectedVariables} />
        )}
        {leftActive === 'outline' && (
          <TemplateOutlinePanel headings={headings} />
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
              submitMsg ? (
                <div
                  role={submitMsg.kind === 'error' ? 'alert' : 'status'}
                  className={submitMsg.kind === 'error' ? styles.alertError : styles.alertSuccess}
                >
                  {submitMsg.text}
                </div>
              ) : importErr ? (
                <div role="alert" className={styles.alertError}>{importErr}</div>
              ) : undefined
            }
          >
            <MetalDocsEditor
              ref={editorRef}
              mode={isDraft ? 'template-draft' : 'readonly'}
              documentBuffer={draft.docxBytes ?? undefined}
              onAutoSave={async (buf) => { autosave.queueDocx(buf); }}
              onChange={handleEditorChange}
              externalPlugins={editorPlugins}
              showRuler
            />
          </EditorChrome>
        </main>
      </div>

      {currentVersion && ['in_review', 'approved', 'published'].includes(currentVersion.status) && (
        <VersionActionPanel
          version={currentVersion}
          onVersionUpdate={(v) => setLiveVersion(v)}
        />
      )}
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

const ICON_OUTLINE = (
  <svg {...SVG_BASE}>
    <path d="M21 12h-8" />
    <path d="M21 6h-8" />
    <path d="M21 18h-8" />
    <path d="M3 6h.01" />
    <path d="M3 12h.01" />
    <path d="M3 18h.01" />
  </svg>
);

const ICON_CHEVRON_LEFT = (
  <svg {...SVG_BASE} width={14} height={14}>
    <path d="M15 18l-6-6 6-6" />
  </svg>
);
