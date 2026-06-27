import React from 'react';
import type { PlaceholderCatalogEntry } from './api/catalog';
import styles from './AvailableTokensPanel.module.css';

export interface AvailableTokensPanelProps {
  catalog: PlaceholderCatalogEntry[];
  usedKeys: Set<string>;
  unknownTokens: string[];
  onInsert: (key: string) => void;
}

export function AvailableTokensPanel({
  catalog,
  usedKeys,
  unknownTokens,
  onInsert,
}: AvailableTokensPanelProps): React.ReactElement {
  return (
    <aside className={styles.panel} aria-label="Tokens disponíveis">
      <header className={styles.head}>
        <h3 className={styles.title}>Preenchido pelo sistema (seguro)</h3>
        <p className={styles.hint}>
          Clique para inserir. O valor é preenchido automaticamente no documento — você não digita.
        </p>
      </header>
      <ul className={styles.list}>
        {catalog.map((it) => {
          const used = usedKeys.has(it.key);
          return (
            <li key={it.key} className={styles.item} data-testid={`token-${it.key}`} data-used={used}>
              <button type="button" className={styles.insertBtn} onClick={() => onInsert(it.key)}>
                <code className={styles.code}>{`{${it.key}}`}</code>
                <span className={styles.label}>{it.label}</span>
                {used && <span className={styles.usedBadge}>em uso</span>}
              </button>
            </li>
          );
        })}
      </ul>
      {unknownTokens.length > 0 && (
        <div className={styles.unknown} role="status">
          <p className={styles.unknownTitle}>Tokens não reconhecidos</p>
          <ul className={styles.unknownList}>
            {unknownTokens.map((t) => (
              <li key={t} data-testid={`unknown-${t}`} className={styles.unknownItem}>
                <code className={styles.code}>{`{${t}}`}</code> — não será preenchido, verifique o nome
              </li>
            ))}
          </ul>
        </div>
      )}
    </aside>
  );
}
