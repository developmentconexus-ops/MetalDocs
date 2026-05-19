import styles from './DocPaperPreview.module.css';

export type DocPaperPreviewProps = {
  /** Number of horizontal text-line bars to render. */
  lines: number;
  /** Optional code label rendered in the top-right corner (e.g. "POP-001 REV00"). */
  code?: string;
  /** Style variant — `thumbnail` (Step 4 large) or `template` (Step 3 small). */
  variant?: 'thumbnail' | 'template';
};

/**
 * Decorative paper-document preview used in Step 3 (template tile) + Step 4
 * (confirm summary). Pure visual — no semantics. The line-width formula is
 * intentionally deterministic (idx-based) so previews stay stable across
 * renders without React keys flickering.
 */
export function DocPaperPreview({
  lines,
  code,
  variant = 'thumbnail',
}: DocPaperPreviewProps): JSX.Element {
  const containerClass =
    variant === 'thumbnail' ? styles.thumbnail : styles.template;
  return (
    <div className={containerClass} aria-hidden="true">
      <div className={styles.titleBar} />
      {code ? <div className={styles.code}>{code}</div> : null}
      {Array.from({ length: lines }).map((_, idx) => (
        <div
          key={idx}
          className={styles.line}
          style={{ width: `${55 + ((idx * 11) % 38)}%` }}
        />
      ))}
    </div>
  );
}
