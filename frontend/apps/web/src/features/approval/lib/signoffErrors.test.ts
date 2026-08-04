import { describe, expect, it } from 'vitest';

import { ApprovalError } from '../api/mutationClient';
import { mapSignoffError, SignoffError } from './signoffErrors';

// These assertions are the contract guard for sign-off error classification. They
// were ported verbatim from the retired SignoffDialog.test (which exercised the
// same mapping through the modal) so the PT copy per error code cannot silently
// drift now that the logic lives in the pure `mapSignoffError`.
describe('mapSignoffError', () => {
  it('any 412 → stale refresh banner (stale=true), whatever code carries it', () => {
    // Classification is by STATUS. The codes that ride a 412 differ per surface
    // (precondition.content_hash_mismatch, conflict.stale_revision, ...), so
    // pinning the branch to one code would silently stop firing for the others.
    for (const code of ['precondition.content_hash_mismatch', 'conflict.stale_revision']) {
      const err = mapSignoffError(new ApprovalError(code, 412, 'stale'));
      expect(err.kind).toBe('stale');
      expect(err.stale).toBe(true);
      expect(err.message).toBe('Documento foi alterado. Atualize a página antes de tentar novamente.');
    }
  });

  it('authn.signature_invalid → bad_password', () => {
    const err = mapSignoffError(new ApprovalError('auth.signature_invalid', 403, 'invalid signature'));
    expect(err.kind).toBe('bad_password');
    expect(err.message).toBe('Senha incorreta. Verifique e tente novamente.');
    expect(err.stale).toBe(false);
  });

  it('AUTH_INVALID_CREDENTIALS (401) → bad_password (takes precedence over session_expired)', () => {
    const err = mapSignoffError(new ApprovalError('auth.invalid_credentials', 401, 'invalid credentials'));
    expect(err.kind).toBe('bad_password');
    expect(err.message).toBe('Senha incorreta. Verifique e tente novamente.');
  });

  it('signoff.not_eligible → not_eligible', () => {
    const err = mapSignoffError(new ApprovalError('permission.signoff_actor_not_eligible', 403, 'not eligible'));
    expect(err.kind).toBe('not_eligible');
    expect(err.message).toContain('Você não está mais elegível para assinar esta etapa.');
  });

  it('status 401 (unmatched code) → session_expired', () => {
    const err = mapSignoffError(new ApprovalError('unmatched.code', 401, 'expired'));
    expect(err.kind).toBe('session_expired');
    expect(err.message).toBe('Sessão expirada. Autentique novamente para assinar.');
  });

  it('429 / ratelimit.exceeded → rate_limited', () => {
    const byStatus = mapSignoffError(new ApprovalError('unmatched.code', 429, 'too many attempts'));
    expect(byStatus.kind).toBe('rate_limited');
    expect(byStatus.message).toBe('Muitas tentativas. Aguarde 30 segundos antes de tentar novamente.');

    const byCode = mapSignoffError(new ApprovalError('ratelimit.exceeded', 400, 'too many'));
    expect(byCode.kind).toBe('rate_limited');
  });

  it('sod.submitter_cannot_sign → sod_submitter', () => {
    const err = mapSignoffError(new ApprovalError('permission.sod_submitter_cannot_sign', 403, 'forbidden'));
    expect(err.kind).toBe('sod_submitter');
    expect(err.message).toContain('submeteu este documento');
  });

  it('sod.cross_stage_duplicate → sod_duplicate', () => {
    const err = mapSignoffError(new ApprovalError('permission.sod_cross_stage_duplicate', 403, 'forbidden'));
    expect(err.kind).toBe('sod_duplicate');
    expect(err.message).toContain('outra etapa');
  });

  it('resolvable business code → server with the resolved message', () => {
    const err = mapSignoffError(new ApprovalError('state.approval_blocked_unresolved_comments', 409, 'pending comments'));
    expect(err.kind).toBe('server');
    expect(err.message).toContain(
      'Resolva os comentários pendentes antes de aprovar ou liberar este documento.',
    );
  });

  it('unresolvable code → server kind carrying the resolved fallback message', () => {
    // resolveErrorMessage returns a generic code-bearing string for unknown codes;
    // the mapper passes it through (the mojibake SERVER_FALLBACK only applies when
    // resolveErrorMessage yields the legacy 'Erro inesperado.' sentinel).
    const err = mapSignoffError(new ApprovalError('unmatched.code', 500, 'boom'));
    expect(err.kind).toBe('server');
    expect(err.message.length).toBeGreaterThan(0);
    expect(err.stale).toBe(false);
  });

  it('TypeError / fetch failure → network', () => {
    expect(mapSignoffError(new TypeError('Failed to fetch')).kind).toBe('network');
    expect(mapSignoffError(new Error('network request failed')).kind).toBe('network');
  });

  it('unknown non-Error → server fallback', () => {
    const err = mapSignoffError('nope');
    expect(err).toBeInstanceOf(SignoffError);
    expect(err.kind).toBe('server');
  });
});
