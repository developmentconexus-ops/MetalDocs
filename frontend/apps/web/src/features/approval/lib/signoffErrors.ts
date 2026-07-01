// Pure sign-off error classifier.
//
// Lifted verbatim from the former SignoffDialog.mapErrorToState + ERROR_MESSAGES
// so the inline decision panel (shared cockpit) and any legacy caller classify a
// failed sign-off IDENTICALLY. No React, no state setters — a raw error in, a
// { kind, message, stale } record out. Do NOT fork a second mapping.
//
// `stale` (HTTP 412 / conflict.stale) is not a terminal error in the UI sense:
// the dialog surfaced it as a "refresh the page" banner and reset to idle rather
// than a red error box. Callers render it as a banner and keep the form usable.

import { ApprovalError } from '../api/mutationClient';
import { resolveErrorMessage } from '../../../lib/api';

export type SignoffErrorKind =
  | 'stale'
  | 'bad_password'
  | 'not_eligible'
  | 'session_expired'
  | 'rate_limited'
  | 'network'
  | 'server'
  | 'sod_submitter'
  | 'sod_duplicate';

/**
 * A classified sign-off failure. Extends `Error` so it can be thrown from the
 * decision model's `submit` and caught by the shared decision panel, which
 * renders `.message` and treats `.stale` as a refresh banner rather than a hard
 * error. `.kind` is kept for callers/tests that branch on the specific cause.
 */
export class SignoffError extends Error {
  readonly kind: SignoffErrorKind;
  /** True only for the 412/stale case — render as a refresh banner, not an error. */
  readonly stale: boolean;

  constructor(kind: SignoffErrorKind, message: string) {
    super(message);
    this.name = 'SignoffError';
    this.kind = kind;
    this.stale = kind === 'stale';
  }
}

// The strings the panel renders per error kind. The retired dialog carried Latin-1
// mojibake / ASCII-stripped copy for two of these; corrected to proper UTF-8 PT-BR
// now that the mapping lives in a pure, testable module.
const MESSAGES: Record<Exclude<SignoffErrorKind, 'server'>, string> = {
  stale: 'Documento foi alterado. Atualize a página antes de tentar novamente.',
  bad_password: 'Senha incorreta. Verifique e tente novamente.',
  not_eligible: 'Você não está mais elegível para assinar esta etapa.',
  session_expired: 'Sessão expirada. Autentique novamente para assinar.',
  rate_limited: 'Muitas tentativas. Aguarde 30 segundos antes de tentar novamente.',
  network: 'Erro de conexão. Verifique sua internet e tente novamente.',
  // E2: SoD codes get Portuguese messages from the shared errorMessages map.
  sod_submitter: resolveErrorMessage(
    'sod.submitter_cannot_sign',
    'Segregação de funções: submissão e aprovação pelo mesmo usuário não é permitida.',
  ),
  sod_duplicate: resolveErrorMessage(
    'sod.cross_stage_duplicate',
    'Você já assinou este documento em uma etapa anterior.',
  ),
};

const SERVER_FALLBACK =
  'Fluxo de aprovação indisponível no momento. Verifique se o documento ainda está elegível para sua revisão.';

function of(kind: Exclude<SignoffErrorKind, 'server'>): SignoffError {
  return new SignoffError(kind, MESSAGES[kind]);
}

/**
 * Classify a failed sign-off. Mirrors SignoffDialog.mapErrorToState exactly:
 * 412/stale → refresh banner; known ApprovalError codes → their specific copy;
 * a resolvable server code → its resolved message; TypeError / fetch failure →
 * network; anything else → the generic server fallback.
 */
export function mapSignoffError(error: unknown): SignoffError {
  if (error instanceof ApprovalError) {
    if (error.status === 412 || error.code === 'conflict.stale') return of('stale');
    if (error.code === 'authn.signature_invalid' || error.code === 'AUTH_INVALID_CREDENTIALS') return of('bad_password');
    if (error.code === 'signoff.not_eligible' || error.code === 'SIGNOFF_NOT_ELIGIBLE') return of('not_eligible');
    if (error.status === 401) return of('session_expired');
    if (error.status === 429 || error.code === 'authn.rate_limited') return of('rate_limited');
    if (error.code === 'sod.submitter_cannot_sign') return of('sod_submitter');
    if (error.code === 'sod.cross_stage_duplicate') return of('sod_duplicate');
    const resolved = resolveErrorMessage(error.code);
    return new SignoffError('server', resolved === 'Erro inesperado.' ? SERVER_FALLBACK : resolved);
  }

  if (error instanceof TypeError) return of('network');
  if (error instanceof Error && /network|failed to fetch|fetch/i.test(error.message)) return of('network');

  return new SignoffError('server', SERVER_FALLBACK);
}
