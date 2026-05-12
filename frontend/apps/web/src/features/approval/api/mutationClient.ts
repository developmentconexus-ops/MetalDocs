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

  const res = await fetch(url, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  const newETag = res.headers.get('ETag');
  if (newETag && opts.resourceId) {
    etagCache.set(opts.resourceId, newETag);
  }

  if (res.ok) {
    return res.json() as Promise<TRes>;
  }

  // Try RFC 9457 problem+json first.
  const prob = await parseProblem(res.clone());
  if (prob) {
    if (prob.status === 401) {
      dispatchAuthExpired(window.location.pathname + window.location.search);
    }
    if (prob.status === 412 && opts.on412 && opts.resourceId) {
      opts.on412(opts.resourceId);
    } else if (prob.status === 412) {
      toast.error('Documento foi alterado. Por favor, atualize a página.');
    }
    throw new ApprovalError(prob.code, prob.status, prob.title ?? prob.detail ?? '');
  }

  // Legacy envelope fallback: { error: { code, message } }.
  const legacyBody = (await res.json().catch(() => ({}))) as {
    error?: { code?: string; message?: string };
  };
  const legacyCode = legacyBody?.error?.code ?? `http_${res.status}`;
  const legacyMessage = legacyBody?.error?.message ?? 'Erro interno';

  if (res.status === 401) {
    dispatchAuthExpired(window.location.pathname + window.location.search);
    throw new ApprovalError('authn.expired', 401, 'Não autorizado');
  }
  if (res.status === 412) {
    if (opts.on412 && opts.resourceId) {
      opts.on412(opts.resourceId);
    } else {
      toast.error('Documento foi alterado. Por favor, atualize a página.');
    }
    throw new ApprovalError(legacyCode, 412, legacyMessage);
  }
  if (res.status === 429) {
    throw new ApprovalError('authn.rate_limited', 429, 'Muitas tentativas. Aguarde 30 segundos.');
  }
  throw new ApprovalError(legacyCode, res.status, legacyMessage);
}
