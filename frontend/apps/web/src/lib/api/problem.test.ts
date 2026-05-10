import { describe, expect, it } from 'vitest';

import { ApiError, parseProblem, resolveErrorMessage } from './problem';

describe('problem', () => {
  it('parseProblem returns null on application/json', async () => {
    const res = new Response(JSON.stringify({ code: 'VALIDATION_ERROR', title: 'x', status: 400 }), {
      headers: { 'Content-Type': 'application/json' },
    });

    await expect(parseProblem(res)).resolves.toBeNull();
  });

  it('parseProblem returns null on application/problem+json with malformed body', async () => {
    const res = new Response('not-json', {
      headers: { 'Content-Type': 'application/problem+json' },
    });

    await expect(parseProblem(res)).resolves.toBeNull();
  });

  it('parseProblem returns null when code is missing', async () => {
    const res = new Response(JSON.stringify({ title: 'Erro', status: 400 }), {
      headers: { 'Content-Type': 'application/problem+json' },
    });

    await expect(parseProblem(res)).resolves.toBeNull();
  });

  it('parseProblem returns Problem for valid problem body with case-insensitive content-type', async () => {
    const payload = {
      code: 'VALIDATION_ERROR',
      title: 'Erro de validação',
      status: 400,
      detail: 'Há erros no formulário.',
    };

    const res = new Response(JSON.stringify(payload), {
      headers: { 'Content-Type': 'Application/Problem+JSON; charset=utf-8' },
    });

    await expect(parseProblem(res)).resolves.toEqual(payload);
  });

  it('ApiError.getFieldErrors returns map keyed by field', () => {
    const err = new ApiError({
      code: 'VALIDATION_ERROR',
      title: 'Erro de validação',
      status: 400,
      errors: [
        { field: 'name', code: 'REQUIRED', message: 'name is required' },
        { field: 'email', code: 'INVALID_FORMAT', message: 'email is invalid' },
      ],
    });

    expect(err.getFieldErrors()).toEqual({
      name: { field: 'name', code: 'REQUIRED', message: 'name is required' },
      email: { field: 'email', code: 'INVALID_FORMAT', message: 'email is invalid' },
    });
  });

  it('ApiError.hasFieldError returns true/false correctly', () => {
    const err = new ApiError({
      code: 'VALIDATION_ERROR',
      title: 'Erro de validação',
      status: 400,
      errors: [{ field: 'name', code: 'REQUIRED', message: 'name is required' }],
    });

    expect(err.hasFieldError('name')).toBe(true);
    expect(err.hasFieldError('email')).toBe(false);
  });

  it('ApiError.fromLegacy creates sensible defaults', () => {
    const err = ApiError.fromLegacy('legacy.code', 418, 'Mensagem legado');

    expect(err.code).toBe('legacy.code');
    expect(err.status).toBe(418);
    expect(err.title).toBe('Mensagem legado');
    expect(err.message).toBe('Mensagem legado');
    expect(err.errors).toEqual([]);
  });

  it('resolveErrorMessage returns mapped pt-BR for known code', () => {
    const err = new ApiError({
      code: 'VALIDATION_ERROR',
      title: 'Validation failed',
      status: 400,
      detail: 'detail',
    });

    expect(resolveErrorMessage(err)).toBe('Há campos inválidos. Revise os dados informados.');
  });

  it('resolveErrorMessage returns detail for unknown code when ApiError has detail', () => {
    const err = new ApiError({
      code: 'CUSTOM_UNKNOWN_CODE',
      title: 'Fallback title',
      status: 400,
      detail: 'Detalhe específico',
    });

    expect(resolveErrorMessage(err)).toBe('Detalhe específico');
  });

  it("resolveErrorMessage falls back to 'Erro inesperado.' for non-Error input", () => {
    expect(resolveErrorMessage(123)).toBe('Erro inesperado.');
  });
});
