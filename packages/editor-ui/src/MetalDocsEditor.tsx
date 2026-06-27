import { forwardRef, useEffect, useImperativeHandle, useMemo, useRef } from 'react';
import { DocxEditor, createEmptyDocument, type DocxEditorRef } from '@eigenpal/docx-editor-react';
import { PluginHost, templatePlugin, type EditorPlugin } from '@eigenpal/docx-editor-react/plugin-api';
import '@eigenpal/docx-editor-react/styles.css';
import type { MetalDocsEditorProps, MetalDocsEditorRef } from './types';
import { filterTransactionGuard } from './plugins/filter-transaction-guard';
import { toEigenpalComment, fromEigenpalComment } from './comment-mapping';
import { detectTokens, type DetectedToken } from '@metaldocs/shared-tokens';

const AUTOSAVE_DEBOUNCE_MS = 1500;

// Structural seam over the vendor's ProseMirror EditorView — declared locally so
// no `@eigenpal`/`prosemirror` type leaks across the ACL wall (see public-surface
// guard test). Only the surface the adapter actually touches is modeled.
type PmView = {
  dom: HTMLElement;
  state: {
    selection: { from: number; to: number };
    doc: {
      content: { size: number };
      textBetween: (from: number, to: number, blockSep: string, leafSep: string) => string;
    };
    tr: { insertText: (text: string, from: number, to: number) => unknown };
  };
  dispatch: (tr: unknown) => void;
  focus: () => void;
};

// Vendor exposes no getActiveView(); the body view plus every header/footer rId
// view are gathered here so insert + token detection treat all bands uniformly.
function collectPmViews(editorRef: ReturnType<DocxEditorRef['getEditorRef']>): PmView[] {
  if (!editorRef) return [];
  const body = editorRef.getView();
  const hf = editorRef.getHfPmViews();
  return [body, ...(hf ? Array.from(hf.values()) : [])].filter(Boolean) as unknown as PmView[];
}

export const MetalDocsEditor = forwardRef<MetalDocsEditorRef, MetalDocsEditorProps>(
  function MetalDocsEditor(props, ref) {
    const inner = useRef<DocxEditorRef>(null);
    const onAutoSaveRef = useRef(props.onAutoSave);
    const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const inFlightRef = useRef(false);
    const rootRef = useRef<HTMLDivElement>(null);
    // The Eigenpal paged editor runs one ProseMirror view per band (body + each
    // header/footer rId). Vendor exposes no getActiveView(), so the adapter tracks
    // the last-focused band via a delegated focusin listener and re-resolves the
    // live view at insert time. See wiki/modules/editor-ui-eigenpal.md.
    const lastFocusedPmRef = useRef<HTMLElement | null>(null);

    onAutoSaveRef.current = props.onAutoSave;

    useImperativeHandle(ref, () => {
      // getDocumentBuffer and saveNow are the same operation under two names the
      // app uses interchangeably — one impl, no drift.
      const save = async () => (inner.current ? ((await inner.current.save()) ?? null) : null);
      // Extracted so getUsedTokens can delegate without a `this` binding issue
      // (TypeScript cannot infer the return type of the object literal while it
      // is still being constructed).
      const getDetectedTokens = (): DetectedToken[] => {
        const views = collectPmViews(inner.current?.getEditorRef() ?? null);
        const seen = new Set<string>();
        const out: DetectedToken[] = [];
        for (const v of views) {
          const doc = v.state.doc;
          const text = doc.textBetween(0, doc.content.size, '\n', '\n');
          for (const d of detectTokens(text)) {
            if (seen.has(d.name)) continue;
            seen.add(d.name);
            out.push(d);
          }
        }
        return out;
      };
      return {
        getDocumentBuffer: save,
        saveNow: save,
        getPageCount() {
          if (!inner.current) return null;
          const total = inner.current.getTotalPages();
          return Number.isInteger(total) && total > 0 ? total : null;
        },
        focus() {
          inner.current?.focus();
        },
        insertToken(key: string) {
          const views = collectPmViews(inner.current?.getEditorRef() ?? null);
          if (views.length === 0) return;
          // Resolve the live caret band fresh: document.activeElement is never
          // stale (the palette button's mousedown-preventDefault keeps focus in
          // the band), whereas a cached focusin node can be detached by an HF
          // painter re-render. Fall back to the tracked node, then the body.
          const active =
            (document.activeElement as HTMLElement | null)?.closest?.('.ProseMirror') as
              | HTMLElement
              | null;
          const dom = active ?? lastFocusedPmRef.current;
          const view =
            (dom && views.find((v) => v.dom === dom || v.dom.contains(dom))) || views[0];
          const { from, to } = view.state.selection;
          view.dispatch(view.state.tr.insertText(`{${key}}`, from, to));
          view.focus();
        },
        getDetectedTokens,
        getUsedTokens() {
          return getDetectedTokens()
            .filter((d) => d.valid)
            .map((d) => d.name);
        },
      };
    }, []);

    useEffect(() => () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    }, []);

    useEffect(() => {
      const root = rootRef.current;
      if (!root) return;
      const onFocusIn = (e: FocusEvent) => {
        const pm = (e.target as HTMLElement | null)?.closest?.('.ProseMirror') as HTMLElement | null;
        if (pm) lastFocusedPmRef.current = pm;
      };
      root.addEventListener('focusin', onFocusIn);
      return () => root.removeEventListener('focusin', onFocusIn);
    }, []);

    const handleChange = () => {
      if (props.mode === 'readonly') return;
      // Lightweight change notify is independent of autosave: consumers use it to
      // refresh derived state (e.g. token usage) and must not be silently coupled
      // to an onAutoSave callback being present.
      props.onChange?.();
      if (!onAutoSaveRef.current) return;
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(async () => {
        if (inFlightRef.current) return;
        if (!inner.current) return;
        const cb = onAutoSaveRef.current;
        if (!cb) return;
        try {
          inFlightRef.current = true;
          const buf = await inner.current.save();
          if (buf) await cb(buf);
        } finally {
          inFlightRef.current = false;
        }
      }, AUTOSAVE_DEBOUNCE_MS);
    };

    // The vendor's onChange fires for the body view only — header/footer PM
    // transactions never reach it, so HF edits would silently miss token-usage
    // refresh and autosave. ProseMirror input events bubble out of every band,
    // so a delegated `input` listener on the root surfaces them. It is scoped to
    // header/footer bands (their PM parent carries `data-hf-kind`) precisely so
    // body edits are NOT double-counted — those already arrive via the vendor
    // onChange prop. The ref keeps the mount-once listener pointed at the latest
    // closure. See wiki/modules/editor-ui-eigenpal.md.
    const handleChangeRef = useRef(handleChange);
    handleChangeRef.current = handleChange;
    useEffect(() => {
      const root = rootRef.current;
      if (!root) return;
      const onInput = (e: Event) => {
        const pm = (e.target as HTMLElement | null)?.closest?.('.ProseMirror') as HTMLElement | null;
        if (pm?.closest('[data-hf-kind]')) handleChangeRef.current();
      };
      root.addEventListener('input', onInput);
      return () => root.removeEventListener('input', onInput);
    }, []);

    const libMode =
      props.mode === 'readonly' ? 'viewing' : props.mode === 'review' ? 'suggesting' : 'editing';
    const blankDocument = useMemo(
      () => (!props.documentBuffer && props.mode !== 'readonly' ? createEmptyDocument() : undefined),
      [props.documentBuffer, props.mode],
    );
    // templatePlugin renders the `template-annotation-chip` items into the
    // unified sidebar (used for variable authoring in template-draft mode).
    // Document editing renders fully-substituted output — skip the plugin so
    // the sidebar stays empty (no chips, canvas centered) when there are also
    // no comments to display. See wiki/modules/editor-ui-eigenpal.md.
    const plugins: EditorPlugin[] = [
      ...(props.mode === 'template-draft'
        ? [templatePlugin, { id: 'filter-transaction-guard', name: 'filter-transaction-guard', proseMirrorPlugins: [filterTransactionGuard()] }]
        : []),
    ];

    return (
      <div ref={rootRef} style={{ display: 'contents' }}>
        <PluginHost plugins={plugins}>
          <DocxEditor
            ref={inner}
            documentBuffer={props.documentBuffer}
            document={blankDocument}
            mode={libMode}
            author={props.author}
            documentName={props.documentName}
            documentNameEditable={props.documentNameEditable ?? (libMode === 'editing')}
            onDocumentNameChange={props.onDocumentNameChange}
            comments={props.comments?.map(toEigenpalComment)}
            onCommentsChange={
              props.onCommentsChange
                ? (cs) => props.onCommentsChange!(cs.map(fromEigenpalComment))
                : undefined
            }
            onCommentAdd={
              props.onCommentAdd ? (c) => props.onCommentAdd!(fromEigenpalComment(c)) : undefined
            }
            onCommentResolve={
              props.onCommentResolve ? (c) => props.onCommentResolve!(fromEigenpalComment(c)) : undefined
            }
            onCommentDelete={
              props.onCommentDelete ? (c) => props.onCommentDelete!(fromEigenpalComment(c)) : undefined
            }
            onCommentReply={
              props.onCommentReply
                ? (reply, parent) => props.onCommentReply!(fromEigenpalComment(reply), fromEigenpalComment(parent))
                : undefined
            }
            renderTitleBarRight={props.renderTitleBarRight}
            showRuler={props.showRuler ?? true}
            // Margin guides are a capability distinct from the ruler; default to the
            // ruler's resolved value for backward compatibility, but allow callers to
            // control them independently.
            showMarginGuides={props.showMarginGuides ?? props.showRuler ?? true}
            showOutlineButton
            showZoomControl
            onChange={handleChange}
          />
        </PluginHost>
      </div>
    );
  }
);
