import React from 'react';
import type { Token } from './tokens/tokenCatalog';
import styles from './AvailableTokensPanel.module.css';

export interface AvailableTokensPanelProps {
  catalog: Token[];
  /** True when the placeholder catalog failed to load. With no catalog, every
   * token would otherwise classify as "unknown"; warn instead of misleading. */
  catalogError?: boolean;
  usedKeys: Set<string>;
  unknownTokens: string[];
  invalidTokens: string[];
  onInsert: (key: string) => void;
}

function TokenList({
  items,
  usedKeys,
  onInsert,
}: {
  items: Token[];
  usedKeys: Set<string>;
  onInsert: (key: string) => void;
}): React.ReactElement {
  return (
    <ul className={styles.list}>
      {items.map((it) => {
        const used = usedKeys.has(it.key);
        return (
          <li key={it.key} className={styles.item} data-testid={`token-${it.key}`} data-used={used}>
            <button type="button" className={styles.insertBtn} onMouseDown={(e) => e.preventDefault()} onClick={() => onInsert(it.key)}>
              <code className={styles.code}>{`{${it.key}}`}</code>
              <span className={styles.label}>{it.label}</span>
              {used && <span className={styles.usedBadge}>em uso</span>}
            </button>
          </li>
        );
      })}
    </ul>
  );
}

export function AvailableTokensPanel({
  catalog,
  catalogError = false,
  usedKeys,
  unknownTokens,
  invalidTokens,
  onInsert,
}: AvailableTokensPanelProps): React.ReactElement {
  const computed = catalog.filter((t) => t.kind === 'computed');
  const dictionary = catalog.filter((t) => t.kind === 'dictionary');

  return (
    <aside className={styles.panel} aria-label="Tokens disponíveis">
      <header className={styles.head}>
        <h3 className={styles.title}>Preenchido pelo sistema (seguro)</h3>
        <p className={styles.hint}>
          Clique para inserir. O valor é preenchido automaticamente no documento — você não digita.
        </p>
      </header>
      {catalogError && (
        <div className={styles.unknown} role="alert" data-testid="catalog-error">
          <p className={styles.unknownTitle}>Não foi possível carregar os tokens</p>
          <ul className={styles.unknownList}>
            <li className={styles.unknownItem}>
              Recarregue a página. Sem o catálogo, os tokens do documento não podem ser validados.
            </li>
          </ul>
        </div>
      )}
      <TokenList items={computed} usedKeys={usedKeys} onInsert={onInsert} />
      {dictionary.length > 0 && (
        <>
          <header className={styles.head}>
            <h3 className={styles.title}>Definido pela sua organização</h3>
            <p className={styles.hint}>
              Constantes da sua organização — inseridas e preenchidas automaticamente.
            </p>
          </header>
          <TokenList items={dictionary} usedKeys={usedKeys} onInsert={onInsert} />
        </>
      )}
      {!catalogError && unknownTokens.length > 0 && (
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
      {invalidTokens.length > 0 && (
        <div className={styles.unknown} role="status">
          <p className={styles.unknownTitle}>Tokens inválidos</p>
          <ul className={styles.unknownList}>
            {invalidTokens.map((t) => (
              <li key={t} data-testid={`invalid-${t}`} className={styles.unknownItem}>
                <code className={styles.code}>{`{${t}}`}</code> — formato inválido (use letras, números e _)
              </li>
            ))}
          </ul>
        </div>
      )}
    </aside>
  );
}
