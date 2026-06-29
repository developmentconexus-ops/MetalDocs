import type { TokenDictionaryEntry } from '../api/tokensTypes';
import styles from './TokenList.module.css';

export interface TokenListProps {
  entries: TokenDictionaryEntry[];
  canManage: boolean;
  onEdit: (entry: TokenDictionaryEntry) => void;
  onDelete: (entry: TokenDictionaryEntry) => void;
}

export function TokenList(props: TokenListProps) {
  const { entries, canManage, onEdit, onDelete } = props;

  if (entries.length === 0) {
    return <div className={styles.empty}>Nenhum token cadastrado.</div>;
  }

  return (
    <table className={styles.table}>
      <thead>
        <tr>
          <th>Nome</th>
          <th>Rótulo</th>
          <th>Valor</th>
          <th>Descrição</th>
          {canManage && <th aria-label="Ações" />}
        </tr>
      </thead>
      <tbody>
        {entries.map((e) => (
          <tr key={e.id}>
            <td className={styles.code}>{e.name}</td>
            <td>{e.label}</td>
            <td className={styles.value} title={e.value}>{e.value}</td>
            <td>{e.description ?? ''}</td>
            {canManage && (
              <td>
                <div className={styles.actions}>
                  <button type="button" className={styles.link} onClick={() => onEdit(e)}>Editar</button>
                  <button type="button" className={styles.link} onClick={() => onDelete(e)}>Excluir</button>
                </div>
              </td>
            )}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
