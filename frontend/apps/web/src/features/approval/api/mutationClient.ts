// @ts-expect-error O pacote existe no workspace, mas este app não expõe typings de uuid.
import { v4 as uuidv4 } from 'uuid';
import { toast } from 'sonner';

import { ApiError, dispatchAuthExpired } from '../../../lib/api';
import { parseProblem } from '../../../lib/api/problem';
import { etagCache } from './etagCache';

// Re-export ApiError subclass for backwards compatibility with existing import sites.
export class ApprovalError extends ApiError {
  constructor(code: string, status: number, message: string) {
    super(code, status, message);
    this.name = 'ApprovalError';
  }
}

export interface MutateOptions {
  idempotencyKey?: string;
  resourceId?: string;
  ifMatch?: string;
  on412?: (resourceId: string) => void;
}

export async function mutate<TReq, TRes>(
  method: 'POST' | 'PUT' | 'PATCH' | 'DELETE',
  url: string,
  body?: TReq,
  opts: MutateOptions = {},
): Promise<TRes> {
  const idempotencyKey = opts.idempotencyKey ?? uuidv4();
  const ifMatch = opts.ifMatch ?? (opts.resourceId ? etagCache.get(opts.resourceId) : undefined);

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Idempotency-Key': idempotencyKey,
  };
  if (ifMatch) headers['If-Match'] = ifMatch;

  // Same-origin /api/v1; `credentials: 'include'` is set explicitly to mirror the
  // shared apiFetch transport (do not rely on the same-origin default). The
  // ETag/If-Match + Idempotency-Key + 412 semantics below are why this stays a
  // dedicated client rather than folding into apiFetch (which returns parsed JSON
  // and does not expose response headers).
  const res = await fetch(url, {
    method,
    credentials: 'include',
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  const newETag = res.headers.get('ETag');
  if (newETag && opts.resourceId) {
    etagCache.set(opts.resourceId, newETag);
  }

  if (res.ok) {
    if (res.status === 204) return undefined as TRes;
    return res.json() as Promise<TRes>;
  }

  // The API emits only RFC 9457 problem+json (AD-2). Parse it; a missing body
  // means a non-handler failure (infra 502/504/bodyless) → synthesize generic.
  const prob = await parseProblem(res.clone());
  if (res.status === 401 && (!prob || prob.code === 'AUTH_UNAUTHORIZED')) {
    dispatchAuthExpired(window.location.pathname + window.location.search);
    throw new ApprovalError('authn.expired', 401, 'Sessão expirada');
  }
  if (res.status === 429) {
    throw new ApprovalError('authn.rate_limited', 429, 'Muitas tentativas. Aguarde 30 segundos.');
  }
  if (res.status === 412) {
    if (opts.on412 && opts.resourceId) {
      opts.on412(opts.resourceId);
    } else {
      toast.error('Documento foi alterado. Por favor, atualize a página.');
    }
  }
  if (prob) {
    throw new ApprovalError(prob.code, prob.status, prob.title ?? prob.detail ?? '');
  }
  const code = `http_${res.status}`;
  throw new ApprovalError(code, res.status, 'Erro interno');
}
