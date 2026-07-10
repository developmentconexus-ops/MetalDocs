import { useState } from 'react';
import type { EditorComment } from '@metaldocs/editor-ui';
import { commentText } from './commentText';
import styles from './AuthorCommentsPanel.module.css';

export type AuthorCommentsPanelProps = {
  comments: EditorComment[];
  onReply: (text: string, parent: EditorComment) => void;
  onResolve: (comment: EditorComment) => void;
  onReopen: (comment: EditorComment) => void;
};

/**
 * AuthorCommentsPanel — F2d.6. Author-waiting sidebar surface (WorkspaceSidebar
 * contextualPanel slot). Read reviewer comment threads and reply / resolve —
 * NEVER edits document content. Pure presentational: every write is delegated
 * via callbacks (onReply/onResolve/onReopen); the panel owns only local composer
 * text. Reuses the existing useDocumentComments mutations at the call site.
 */
export function AuthorCommentsPanel({
  comments,
  onReply,
  onResolve,
  onReopen,
}: AuthorCommentsPanelProps): React.ReactElement {
  const roots = comments.filter((c) => c.parentId == null);
  const repliesOf = (rootId: EditorComment['id']) =>
    comments.filter((c) => c.parentId === rootId).sort((a, b) => Number(a.id) - Number(b.id));

  return (
    <aside className={styles.root} aria-label="Comentários da revisão">
      <div className={styles.header}>
        <span className={styles.title}>Comentários da revisão</span>
      </div>
      {roots.length === 0 ? (
        <p className={styles.emptyState}>Nenhum comentário nesta revisão.</p>
      ) : (
        <ul className={styles.list}>
          {roots.map((root) => (
            <CommentThread
              key={root.id}
              root={root}
              replies={repliesOf(root.id)}
              onReply={onReply}
              onResolve={onResolve}
              onReopen={onReopen}
            />
          ))}
        </ul>
      )}
    </aside>
  );
}

function CommentThread({
  root,
  replies,
  onReply,
  onResolve,
  onReopen,
}: {
  root: EditorComment;
  replies: EditorComment[];
  onReply: (text: string, parent: EditorComment) => void;
  onResolve: (comment: EditorComment) => void;
  onReopen: (comment: EditorComment) => void;
}): React.ReactElement {
  const [draft, setDraft] = useState('');

  const submit = () => {
    const text = draft.trim();
    if (!text) return;
    onReply(text, root);
    setDraft('');
  };

  return (
    <li className={styles.card}>
      <div className={styles.cardMeta}>
        <span className={styles.cardAuthor}>{root.author}</span>
        {root.resolved ? <span className={styles.resolvedTag}>Resolvido</span> : null}
      </div>
      <p className={styles.cardBody}>{commentText(root.body)}</p>

      {replies.length > 0 ? (
        <ul className={styles.replies}>
          {replies.map((reply) => (
            <li key={reply.id} className={styles.reply}>
              <span className={styles.cardAuthor}>{reply.author}</span>
              <p className={styles.cardBody}>{commentText(reply.body)}</p>
            </li>
          ))}
        </ul>
      ) : null}

      <div className={styles.cardActions}>
        {root.resolved ? (
          <button type="button" className={styles.actionBtn} onClick={() => onReopen(root)}>
            Reabrir
          </button>
        ) : (
          <button type="button" className={styles.actionBtn} onClick={() => onResolve(root)}>
            Resolver
          </button>
        )}
      </div>

      <div className={styles.composer}>
        <label className={styles.srOnly} htmlFor={`reply-${root.id}`}>
          Responder ao comentário de {root.author}
        </label>
        <textarea
          id={`reply-${root.id}`}
          className={styles.textarea}
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          rows={2}
          placeholder="Responder…"
        />
        <button
          type="button"
          className={`${styles.actionBtn} ${styles.replyBtn}`}
          onClick={submit}
          disabled={draft.trim().length === 0}
        >
          Responder
        </button>
      </div>
    </li>
  );
}
