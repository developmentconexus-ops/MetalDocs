import type { ReactNode } from 'react';
import type { Comment, EditorPlugin } from '@eigenpal/docx-js-editor';
import type { SidebarModel } from './plugins/mergefieldPlugin';

export type EditorMode = 'template-draft' | 'document-edit' | 'readonly';

export interface MetalDocsEditorProps {
  documentId?: string;
  documentBuffer?: ArrayBuffer;
  mode: EditorMode;
  author?: string;
  documentName?: string;
  documentNameEditable?: boolean;
  onDocumentNameChange?: (name: string) => void;
  comments?: Comment[];
  onCommentsChange?: (comments: Comment[]) => void;
  onCommentAdd?: (c: Comment) => void;
  onCommentResolve?: (c: Comment) => void;
  onCommentDelete?: (c: Comment) => void;
  onCommentReply?: (reply: Comment, parent: Comment) => void;
  renderTitleBarRight?: () => ReactNode;
  sidebarModel?: SidebarModel;
  externalPlugins?: EditorPlugin[];
  onAutoSave?: (buf: ArrayBuffer) => Promise<void>;
  /**
   * Called synchronously on every editor change event (before the autosave
   * debounce fires). Use for lightweight side-effects (placeholder sync,
   * outline refresh). Do NOT trigger heavy async work here.
   */
  onChange?: () => void;
  onLockLost?: () => void;
  showRuler?: boolean;
}

export interface MetalDocsEditorRef {
  getDocumentBuffer(): Promise<ArrayBuffer | null>;
  focus(): void;
}
