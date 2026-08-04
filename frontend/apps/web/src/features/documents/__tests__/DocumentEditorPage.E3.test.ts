import { describe, it, expect } from 'vitest';
import { resolveErrorMessage } from '../../../lib/api';

describe('E3 — submit error message resolution', () => {
  it('not_found.route maps to Portuguese route missing message', () => {
    expect(resolveErrorMessage('notfound.approval_route', 'no route')).toContain('Nenhuma rota de aprovação');
  });

  it('unknown code falls back to backend message', () => {
    expect(resolveErrorMessage('unmatched.code', 'server error')).toBe('server error');
  });
});
