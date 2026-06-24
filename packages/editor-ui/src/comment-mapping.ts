import type { Comment } from '@eigenpal/docx-editor-core/types/content';

/**
 * EditorComment — MetalDocs-owned neutral comment type that crosses the ACL wall.
 *
 * This is the ONLY comment shape the app sees. It deliberately hides the eigenpal
 * `Comment` realization (Class B) and treats the comment body as opaque (Class C:
 * `body: unknown` is a ProseMirror Paragraph[] under the hood — the app never
 * inspects it). Mapping to/from the vendor type happens ONLY here and in
 * MetalDocsEditor at the <DocxEditor> boundary.
 */
export interface EditorComment {
  /** eigenpal library-comment id (number) — NOT the DB uuid. */
  id: number;
  parentId?: number;
  author: string;
  /** ISO timestamp; maps to eigenpal `date`. */
  createdAt?: string;
  /** Opaque comment body (ProseMirror Paragraph[]); never inspected above the wall. */
  body: unknown;
  /** Maps to eigenpal `done`. */
  resolved: boolean;
}

/** Class B: derive author initials for vendor display. Internal to the wall. */
function toInitials(author: string): string {
  return author
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]!.toUpperCase())
    .join('');
}

/** Map a MetalDocs EditorComment to the eigenpal Comment the vendor editor expects. */
export function toEigenpalComment(ec: EditorComment): Comment {
  return {
    id: ec.id,
    parentId: ec.parentId,
    author: ec.author,
    date: ec.createdAt,
    content: ec.body as Comment['content'],
    done: ec.resolved,
    initials: toInitials(ec.author),
  };
}

/** Map an eigenpal Comment (from vendor callbacks) back to a MetalDocs EditorComment. */
export function fromEigenpalComment(c: Comment): EditorComment {
  return {
    id: c.id,
    parentId: c.parentId,
    author: c.author,
    createdAt: c.date,
    body: c.content,
    resolved: Boolean(c.done),
  };
}
