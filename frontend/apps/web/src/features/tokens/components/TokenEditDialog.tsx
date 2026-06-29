import { useState } from 'react';
import { validateEntry, type TokenFieldErrors, type TokenFormValues } from '../validation';
import styles from './TokenEditDialog.module.css';

export interface TokenEditDialogProps {
  mode: 'create' | 'edit';
  computedKeys: string[];
  submitting: boolean;
  initial?: TokenFormValues;
  onSubmit: (values: TokenFormValues) => void;
  onClose: () => void;
}

const EMPTY: TokenFormValues = { name: '', value: '', label: '', description: '' };

export function TokenEditDialog(props: TokenEditDialogProps) {
  const { mode, computedKeys, submitting, initial, onSubmit, onClose } = props;
  const [values, setValues] = useState<TokenFormValues>(initial ?? EMPTY);
  const [errors, setErrors] = useState<TokenFieldErrors>({});

  function set<K extends keyof TokenFormValues>(key: K, v: string) {
    setValues((prev) => ({ ...prev, [key]: v }));
  }

  function handleSubmit() {
    const errs = validateEntry(values, computedKeys);
    setErrors(errs);
    if (Object.keys(errs).length === 0) onSubmit(values);
  }

  return (
    <div className={styles.overlay} role="dialog" aria-modal="true">
      <div className={styles.dialog}>
        <h2 className={styles.title}>{mode === 'create' ? 'Novo token' : 'Editar token'}</h2>

        <div className={styles.field}>
          <label htmlFor="token-name">Nome</label>
          <input
            id="token-name"
            value={values.name}
            disabled={mode === 'edit'}
            onChange={(e) => set('name', e.target.value)}
          />
          {errors.name && <span className={styles.error}>{errors.name}</span>}
        </div>

        <div className={styles.field}>
          <label htmlFor="token-value">Valor</label>
          <textarea id="token-value" rows={3} value={values.value} onChange={(e) => set('value', e.target.value)} />
          {errors.value && <span className={styles.error}>{errors.value}</span>}
        </div>

        <div className={styles.field}>
          <label htmlFor="token-label">Rótulo</label>
          <input id="token-label" value={values.label} onChange={(e) => set('label', e.target.value)} />
          {errors.label && <span className={styles.error}>{errors.label}</span>}
        </div>

        <div className={styles.field}>
          <label htmlFor="token-description">Descrição</label>
          <textarea
            id="token-description"
            rows={2}
            value={values.description}
            onChange={(e) => set('description', e.target.value)}
          />
          {errors.description && <span className={styles.error}>{errors.description}</span>}
        </div>

        <div className={styles.actions}>
          <button type="button" className={styles.secondary} onClick={onClose}>Cancelar</button>
          <button type="button" className={styles.primary} disabled={submitting} onClick={handleSubmit}>Salvar</button>
        </div>
      </div>
    </div>
  );
}
