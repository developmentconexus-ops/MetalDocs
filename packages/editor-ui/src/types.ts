import type { ReactNode } from 'react';
import type { EditorComment } from './comment-mapping';
import type { DetectedToken } from '@metaldocs/shared-tokens';

export type EditorMode = 'template-draft' | 'document-edit' | 'readonly' | 'review';

export interface MetalDocsEditorProps {
  documentBuffer?: ArrayBuffer;
  mode: EditorMode;
  author?: string;
  documentName?: string;
  documentNameEditable?: boolean;
  onDocumentNameChange?: (name: string) => void;
  comments?: EditorComment[];
  onCommentsChange?: (comments: EditorComment[]) => void;
  onCommentAdd?: (c: EditorComment) => void;
  onCommentResolve?: (c: EditorComment) => void;
  onCommentDelete?: (c: EditorComment) => void;
  onCommentReply?: (reply: EditorComment, parent: EditorComment) => void;
  renderTitleBarRight?: () => ReactNode;
  onAutoSave?: (buf: ArrayBuffer) => Promise<void>;
  onChange?: () => void;
  showRuler?: boolean;
  showMarginGuides?: boolean;
}

export interface MetalDocsEditorRef {
  getDocumentBuffer(): Promise<ArrayBuffer | null>;
  saveNow(): Promise<ArrayBuffer | null>;
  getPageCount(): number | null;
  focus(): void;
  insertToken(key: string): void;
  getUsedTokens(): string[];
  /** Every variable tag across all bands, broad-detected + validity-classified. */
  getDetectedTokens(): DetectedToken[];
}

export type { DetectedToken } from '@metaldocs/shared-tokens';
