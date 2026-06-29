import { useMemo, useState } from 'react';
import { usePlaceholderCatalogQuery } from '../../templates';
import { useHasCapability } from '../../iam/hooks/useHasCapability';
import { resolveQueryError } from '../../../lib/api';
import { TokenList } from '../components/TokenList';
import { TokenEditDialog } from '../components/TokenEditDialog';
import { useTokensQuery } from '../queries/useTokensQuery';
import { useTokenMutations } from '../queries/useTokenMutations';
import type { TokenDictionaryEntry } from '../api/tokensTypes';
import type { TokenFormValues } from '../validation';
import styles from './TokensRoutePage.module.css';

type DialogState =
  | { open: false }
  | { open: true; mode: 'create' }
  | { open: true; mode: 'edit'; entry: TokenDictionaryEntry };

export function Component() {
  const canManage = useHasCapability('token_dictionary.manage');
  const tokensQuery = useTokensQuery();
  const catalogQuery = usePlaceholderCatalogQuery();
  const { create, update, remove } = useTokenMutations();
  const [dialog, setDialog] = useState<DialogState>({ open: false });

  const computedKeys = useMemo(
    () => (catalogQuery.data ?? []).map((c) => c.key),
    [catalogQuery.data],
  );

  function handleSubmit(values: TokenFormValues) {
    const req = {
      name: values.name,
      value: values.value,
      label: values.label,
      description: values.description.length > 0 ? values.description : undefined,
    };
    if (dialog.open && dialog.mode === 'edit') {
      update.mutate({ id: dialog.entry.id, req }, { onSuccess: () => setDialog({ open: false }) });
    } else {
      create.mutate(req, { onSuccess: () => setDialog({ open: false }) });
    }
  }

  function handleDelete(entry: TokenDictionaryEntry) {
    if (window.confirm(`Excluir o token {${entry.name}}?`)) remove.mutate(entry.id);
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Dicionário de tokens</h1>
          <p className={styles.subtitle}>Constantes reutilizáveis preenchidas automaticamente nos documentos.</p>
        </div>
        {canManage && (
          <button type="button" className={styles.newBtn} onClick={() => setDialog({ open: true, mode: 'create' })}>
            Novo token
          </button>
        )}
      </div>

      {tokensQuery.isLoading && <div className={styles.state}>Carregando tokens...</div>}
      {tokensQuery.isError && (
        <div className={styles.state} role="alert">{resolveQueryError(tokensQuery.error)}</div>
      )}
      {tokensQuery.data && (
        <TokenList
          entries={tokensQuery.data}
          canManage={canManage}
          onEdit={(entry) => setDialog({ open: true, mode: 'edit', entry })}
          onDelete={handleDelete}
        />
      )}

      {dialog.open && (
        <TokenEditDialog
          mode={dialog.mode}
          computedKeys={computedKeys}
          submitting={create.isPending || update.isPending}
          initial={
            dialog.mode === 'edit'
              ? {
                  name: dialog.entry.name,
                  value: dialog.entry.value,
                  label: dialog.entry.label,
                  description: dialog.entry.description ?? '',
                }
              : undefined
          }
          onSubmit={handleSubmit}
          onClose={() => setDialog({ open: false })}
        />
      )}
    </div>
  );
}
