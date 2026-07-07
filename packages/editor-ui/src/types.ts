import type { ReactNode } from 'react';
import type { EditorComment } from './comment-mapping';
import type { DetectedToken } from '@metaldocs/shared-tokens';

export type EditorMode = 'template-draft' | 'document-edit' | 'readonly' | 'review';

/**
 * The kind of a tracked change. MetalDocs-owned literal union (NOT imported from
 * the vendor) so the ACL wall stays closed while consumers keep exhaustiveness
 * checking in sidebar rendering. Members mirror the Word revision kinds:
 * inline (`insertion`/`deletion`/`replacement`), paragraph-mark, and the table
 * row/cell/table structural variants.
 */
export type TrackedChangeType =
  | 'insertion'
  | 'deletion'
  | 'replacement'
  | 'paragraphMarkInsertion'
  | 'paragraphMarkDeletion'
  | 'paragraphPropertiesChanged'
  | 'rowInserted'
  | 'rowDeleted'
  | 'rowPropertiesChanged'
  | 'cellInserted'
  | 'cellDeleted'
  | 'cellMerged'
  | 'cellPropertiesChanged'
  | 'tableInserted'
  | 'tableDeleted'
  | 'tablePropertiesChanged';

/**
 * A tracked change (Word revision) surfaced across the ACL wall as a neutral,
 * vendor-free shape. `revisionId` is a string on this public surface even though
 * the underlying Word `w:id` is numeric — the wall never leaks the vendor's
 * numeric-id contract. `excerpt` is a plain-text preview for sidebar cards.
 */
export interface TrackedChange {
  revisionId: string;
  author: string;
  type: TrackedChangeType;
  excerpt: string;
}

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
  /**
   * Fired whenever the tracked-change set changes (edit, accept, reject) — the
   * review sidebar and the author request-changes panel both re-render off this.
   */
  onTrackedChangesChange?: (changes: TrackedChange[]) => void;
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
  /** Tracked changes in the body, neutral-mapped and sorted by document position. */
  getTrackedChanges(): TrackedChange[];
  /** Accept one conceptual change — resolves every site sharing its revision (incl. replacement + coalesced ids). */
  acceptChange(revisionId: string): void;
  /** Reject one conceptual change — inverse of {@link acceptChange}. */
  rejectChange(revisionId: string): void;
  /** Accept every tracked change in the body. */
  acceptAllChanges(): void;
  /** Reject every tracked change in the body. */
  rejectAllChanges(): void;
  /** Remove a comment mark by the library comment id (used when a thread is deleted). */
  removeCommentMark(libraryCommentId: string): void;
}

export type { DetectedToken } from '@metaldocs/shared-tokens';
