import { useState } from 'react';

import { etagCache } from '../../api/etagCache';
import styles from './IntegrityDisclosure.module.css';

interface IntegrityDisclosureProps {
  documentId: string;
  contentHash: string | null;
  frozenContentHash: string | null;
}

/**
 * Collapsed-by-default auditor disclosure (spec §1.3). Renders the hash + ETag
 * ONLY while `open` is true — a controlled `<details>` via `onToggle`, since
 * jsdom mounts `<details>` children regardless of the `open` attribute. This
 * gives a hard guarantee that `queryByText(hash)` is null while collapsed.
 */
export function IntegrityDisclosure({ documentId, contentHash, frozenContentHash }: IntegrityDisclosureProps) {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState<'hash' | 'etag' | null>(null);

  const hash = frozenContentHash ?? contentHash;
  const etag = etagCache.get(documentId) ?? '-';

  const copy = async (value: string, which: 'hash' | 'etag') => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(which);
      window.setTimeout(() => setCopied(null), 1200);
    } catch (_error) {
      setCopied(null);
    }
  };

  return (
    <details
      className={styles.disclosure}
      open={open}
      onToggle={(e) => setOpen((e.target as HTMLDetailsElement).open)}
    >
      <summary className={styles.summary}>Conteúdo verificado ✓ · detalhes</summary>

      {open && (
        <div className={styles.body}>
          <div className={styles.row}>
            <span className={styles.label}>Hash do conteúdo</span>
            <div className={styles.valueRow}>
              <code className={styles.value}>{hash ?? '—'}</code>
              {hash != null && (
                <button type="button" className={styles.copyButton} onClick={() => void copy(hash, 'hash')}>
                  {copied === 'hash' ? 'Copiado' : 'Copiar'}
                </button>
              )}
            </div>
          </div>

          <div className={styles.row}>
            <span className={styles.label}>ETag</span>
            <div className={styles.valueRow}>
              <code className={styles.value}>{etag}</code>
              <button type="button" className={styles.copyButton} onClick={() => void copy(etag, 'etag')}>
                {copied === 'etag' ? 'Copiado' : 'Copiar'}
              </button>
            </div>
          </div>
        </div>
      )}
    </details>
  );
}
