import { describe, expect, it } from 'vitest';
import { validateName, validateEntry } from './validation';

const COMPUTED = ['author', 'doc_code', 'effective_date'];

describe('validateName', () => {
  it('accepts a valid snake_case name', () => {
    expect(validateName('company_slogan', COMPUTED)).toBeNull();
  });
  it('rejects empty', () => {
    expect(validateName('', COMPUTED)).toBe('Nome obrigatório.');
  });
  it('rejects > 64 chars', () => {
    expect(validateName('a'.repeat(65), COMPUTED)).toBe('Nome deve ter no máximo 64 caracteres.');
  });
  it('rejects invalid grammar (leading digit)', () => {
    expect(validateName('1slogan', COMPUTED)).toBe('Nome inválido: use letras, números e _ , começando por letra ou _.');
  });
  it('rejects a reserved ident', () => {
    expect(validateName('constructor', COMPUTED)).toBe('Nome reservado pelo sistema.');
  });
  it('rejects a computed-catalog collision', () => {
    expect(validateName('author', COMPUTED)).toBe('Nome já é um token do sistema (catálogo computado).');
  });
});

describe('validateEntry', () => {
  it('returns no errors for a valid entry', () => {
    expect(
      validateEntry({ name: 'company_slogan', value: 'Qualidade', label: 'Slogan', description: '' }, COMPUTED),
    ).toEqual({});
  });
  it('flags required value and label', () => {
    expect(
      validateEntry({ name: 'company_slogan', value: '', label: '', description: '' }, COMPUTED),
    ).toEqual({ value: 'Valor obrigatório.', label: 'Rótulo obrigatório.' });
  });
  it('flags value over 4096 and description over 1024', () => {
    const errs = validateEntry(
      { name: 'company_slogan', value: 'a'.repeat(4097), label: 'L', description: 'b'.repeat(1025) },
      COMPUTED,
    );
    expect(errs.value).toBe('Valor deve ter no máximo 4096 caracteres.');
    expect(errs.description).toBe('Descrição deve ter no máximo 1024 caracteres.');
  });
});
