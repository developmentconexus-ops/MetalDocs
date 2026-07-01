import { describe, expect, it } from 'vitest';

import { ApprovalError } from '../api/mutationClient';
import { mapSignoffError, SignoffError } from './signoffErrors';

// These assertions are the contract guard for sign-off error classification. They
// were ported verbatim from the retired SignoffDialog.test (which exercised the
// same mapping through the modal) so the PT copy per error code cannot silently
// drift now that the logic lives in the pure `mapSignoffError`.
describe('mapSignoffError', () => {
  it('412 / conflict.stale → stale refresh banner (stale=true)', () => {
    const byStatus = mapSignoffError(new ApprovalError('whatever', 412, 'stale'));
    expect(byStatus.kind).toBe('stale');
    expect(byStatus.stale).toBe(true);
    expect(byStatus.message).toBe('Documento foi alterado. Atualize a página antes de tentar novamente.');

    const byCode = mapSignoffError(new ApprovalError('conflict.stale', 409, 'stale'));
    expect(byCode.kind).toBe('stale');
    expect(byCode.stale).toBe(true);
  });

  it('authn.signature_invalid → bad_password', () => {
    const err = mapSignoffError(new ApprovalError('authn.signature_invalid', 403, 'invalid signature'));
    expect(err.kind).toBe('bad_password');
    expect(err.message).toBe('Senha incorreta. Verifique e tente novamente.');
    expect(err.stale).toBe(false);
  });

  it('AUTH_INVALID_CREDENTIALS (401) → bad_password (takes precedence over session_expired)', () => {
    const err = mapSignoffError(new ApprovalError('AUTH_INVALID_CREDENTIALS', 401, 'invalid credentials'));
    expect(err.kind).toBe('bad_password');
    expect(err.message).toBe('Senha incorreta. Verifique e tente novamente.');
  });

  it('signoff.not_eligible → not_eligible', () => {
    const err = mapSignoffError(new ApprovalError('signoff.not_eligible', 403, 'not eligible'));
    expect(err.kind).toBe('not_eligible');
    expect(err.message).toContain('Voce nao esta mais elegivel para assinar esta etapa.');
  });

  it('status 401 (unmatched code) → session_expired', () => {
    const err = mapSignoffError(new ApprovalError('authn.session_expired', 401, 'expired'));
    expect(err.kind).toBe('session_expired');
    expect(err.message).toBe('Sessão expirada. Autentique novamente para assinar.');
  });

  it('429 / authn.rate_limited → rate_limited', () => {
    const byStatus = mapSignoffError(new ApprovalError('too_many', 429, 'too many attempts'));
    expect(byStatus.kind).toBe('rate_limited');
    expect(byStatus.message).toBe('Muitas tentativas. Aguarde 30 segundos antes de tentar novamente.');

    const byCode = mapSignoffError(new ApprovalError('authn.rate_limited', 400, 'too many'));
    expect(byCode.kind).toBe('rate_limited');
  });

  it('sod.submitter_cannot_sign → sod_submitter', () => {
    const err = mapSignoffError(new ApprovalError('sod.submitter_cannot_sign', 403, 'forbidden'));
    expect(err.kind).toBe('sod_submitter');
    expect(err.message).toContain('submeteu este documento');
  });

  it('sod.cross_stage_duplicate → sod_duplicate', () => {
    const err = mapSignoffError(new ApprovalError('sod.cross_stage_duplicate', 403, 'forbidden'));
    expect(err.kind).toBe('sod_duplicate');
    expect(err.message).toContain('outra etapa');
  });

  it('resolvable business code → server with the resolved message', () => {
    const err = mapSignoffError(new ApprovalError('approval.unresolved_comments', 409, 'pending comments'));
    expect(err.kind).toBe('server');
    expect(err.message).toContain(
      'Resolva os comentários pendentes antes de aprovar ou liberar este documento.',
    );
  });

  it('unresolvable code → server fallback copy', () => {
    const err = mapSignoffError(new ApprovalError('totally.unknown.code', 500, 'boom'));
    expect(err.kind).toBe('server');
    expect(err.message).toContain('Fluxo de aprova');
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
