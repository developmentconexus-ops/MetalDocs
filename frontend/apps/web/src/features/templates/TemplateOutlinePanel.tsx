import * as React from 'react';
import type { Heading } from './lib/readHeadings';
import styles from './TemplateOutlinePanel.module.css';

export interface TemplateOutlinePanelProps {
  headings: Heading[];
}

export function TemplateOutlinePanel({ headings }: TemplateOutlinePanelProps): React.ReactElement {
  return (
    <aside className={styles.panel} aria-label="Estrutura do documento">
      <header className={styles.head}>
        <h3 className={styles.title}>Estrutura</h3>
        <p className={styles.hint}>Títulos detectados no documento.</p>
      </header>
      <div className={styles.body}>
        {headings.length === 0 ? (
          <p className={styles.empty}>
            Aplique estilos de título (Título 1, Título 2…) no editor para listar a estrutura aqui.
          </p>
        ) : (
          <ul className={styles.list}>
            {headings.map((h) => (
              <li key={h.id} className={styles.item} data-level={h.level}>
                <span className={styles.text}>{h.text}</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </aside>
  );
}
