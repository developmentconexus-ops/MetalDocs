import { isReservedIdent, isValidIdent } from '@metaldocs/shared-tokens';

export interface TokenFormValues {
  name: string;
  value: string;
  label: string;
  description: string;
}

export type TokenFieldErrors = Partial<Record<keyof TokenFormValues, string>>;

export function validateName(name: string, computedKeys: string[]): string | null {
  if (name.length === 0) return 'Nome obrigatório.';
  if (name.length > 64) return 'Nome deve ter no máximo 64 caracteres.';
  if (!isValidIdent(name)) return 'Nome inválido: use letras, números e _ , começando por letra ou _.';
  if (isReservedIdent(name)) return 'Nome reservado pelo sistema.';
  if (computedKeys.includes(name)) return 'Nome já é um token do sistema (catálogo computado).';
  return null;
}

export function validateEntry(values: TokenFormValues, computedKeys: string[]): TokenFieldErrors {
  const errors: TokenFieldErrors = {};

  const nameErr = validateName(values.name, computedKeys);
  if (nameErr) errors.name = nameErr;

  if (values.value.length === 0) errors.value = 'Valor obrigatório.';
  else if (values.value.length > 4096) errors.value = 'Valor deve ter no máximo 4096 caracteres.';

  if (values.label.length === 0) errors.label = 'Rótulo obrigatório.';
  else if (values.label.length > 256) errors.label = 'Rótulo deve ter no máximo 256 caracteres.';

  if (values.description.length > 1024) errors.description = 'Descrição deve ter no máximo 1024 caracteres.';

  return errors;
}
