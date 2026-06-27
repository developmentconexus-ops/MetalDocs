import { forwardRef, useEffect, useImperativeHandle, useMemo, useRef } from 'react';
import { DocxEditor, createEmptyDocument, type DocxEditorRef } from '@eigenpal/docx-editor-react';
import { PluginHost, templatePlugin, getTemplatePluginTags, type EditorPlugin } from '@eigenpal/docx-editor-react/plugin-api';
import '@eigenpal/docx-editor-react/styles.css';
import type { MetalDocsEditorProps, MetalDocsEditorRef } from './types';
import { filterTransactionGuard } from './plugins/filter-transaction-guard';
import { toEigenpalComment, fromEigenpalComment } from './comment-mapping';

const AUTOSAVE_DEBOUNCE_MS = 1500;

export const MetalDocsEditor = forwardRef<MetalDocsEditorRef, MetalDocsEditorProps>(
  function MetalDocsEditor(props, ref) {
    const inner = useRef<DocxEditorRef>(null);
    const onAutoSaveRef = useRef(props.onAutoSave);
    const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const inFlightRef = useRef(false);

    onAutoSaveRef.current = props.onAutoSave;

    useImperativeHandle(ref, () => {
      // getDocumentBuffer and saveNow are the same operation under two names the
      // app uses interchangeably — one impl, no drift.
      const save = async () => (inner.current ? ((await inner.current.save()) ?? null) : null);
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
          const view = inner.current?.getEditorRef()?.getView();
          if (!view) return;
          const { from, to } = view.state.selection;
          view.dispatch(view.state.tr.insertText(`{${key}}`, from, to));
          view.focus();
        },
        getUsedTokens() {
          const state = inner.current?.getEditorRef()?.getState();
          if (!state) return [];
          return getTemplatePluginTags(state)
            .filter((t) => t.type === 'variable')
            .map((t) => t.name);
        },
      };
    }, []);

    useEffect(() => () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    }, []);

    const handleChange = () => {
      if (props.mode === 'readonly') return;
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
      // Notify consumers of change before debounce fires (for lightweight sync).
      props.onChange?.();
    };

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
    );
  }
);
