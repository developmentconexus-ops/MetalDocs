import { apiFetch } from '../../../lib/api/client';
import { ApiError } from '../../../lib/api/errors';
import type { components, paths } from '../../../lib/api-types';
import type { Placeholder, CompositionConfig } from '../placeholder-types';
export type { Placeholder, CompositionConfig };

type CreateTemplateRequest =
  paths['/templates']['post']['requestBody']['content']['application/json'];
type CreateTemplateResponse =
  paths['/templates']['post']['responses'][201]['content']['application/json'];
type ListTemplatesResponse =
  paths['/templates']['get']['responses'][200]['content']['application/json'];
type GeneratedTemplateDTO = components['schemas']['TemplateDTO'];
type GeneratedVersionDTO = components['schemas']['VersionDTO'];

export type TemplateDTO = Omit<
  GeneratedTemplateDTO,
  | 'archived_at'
  | 'description'
  | 'doc_type_code'
  | 'published_version_id'
> & {
  archived_at: string | null;
  description: string | null | undefined;
  doc_type_code: string | null;
  published_version_id: string | null;
};
export type VersionDTO = Omit<
  GeneratedVersionDTO,
  | 'approved_at'
  | 'approver_id'
  | 'content_hash'
  | 'docx_storage_key'
  | 'metadata_schema'
  | 'obsoleted_at'
  | 'pending_approver_role'
  | 'pending_reviewer_role'
  | 'placeholder_schema'
  | 'published_at'
  | 'reviewed_at'
  | 'reviewer_id'
  | 'submitted_at'
> & {
  approved_at: string | null;
  approver_id: string | null;
  content_hash: string | null;
  docx_storage_key: string | null;
  metadata_schema: Record<string, unknown> | null;
  obsoleted_at: string | null;
  pending_approver_role: string | null;
  pending_reviewer_role: string | null;
  // ADR 0035 / F1.2: backend always emits the placeholder schema as an ordered
  // array. The generated DTO already types it as an array; this override keeps
  // the precise element shape instead of widening it back to an object.
  placeholder_schema: WirePlaceholder[] | null;
  published_at: string | null;
  reviewed_at: string | null;
  reviewer_id: string | null;
  submitted_at: string | null;
};
export type VersionStatus = VersionDTO['status'];

export interface TemplateSchemas {
  placeholders: Placeholder[];
  composition: CompositionConfig | null;
}

export type TemplateListRow = {
  id: string;
  key: string;
  name: string;
  description?: string;
  latest_version: number;
  latest_version_id?: string;
  published_version_id?: string | null;
  updated_at?: string;
  doc_type_code?: string | null;
  archived_at: string | null;
};

function toFiniteNumber(value: unknown): number | null {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === 'string') {
    const n = Number(value);
    if (Number.isFinite(n)) {
      return n;
    }
  }
  return null;
}

export async function createTemplate(cmd: {
  key: string;
  name: string;
  description?: string;
  doc_type_code?: string;
  idempotencyKey: string;
}): Promise<{ template: TemplateDTO; version: VersionDTO }> {
  const payload: CreateTemplateRequest = {
    key: cmd.key,
    name: cmd.name,
    ...(cmd.description ? { description: cmd.description } : {}),
    ...(cmd.doc_type_code ? { doc_type_code: cmd.doc_type_code } : {}),
  };
  const body = await apiFetch<CreateTemplateResponse>('/api/v1/templates', {
    method: 'POST',
    idempotencyKey: cmd.idempotencyKey,
    body: JSON.stringify(payload),
  });
  const template = body.data.template as TemplateDTO;
  const version = body.data.version as VersionDTO;
  if (!template || !version) {
    throw new Error('Resposta de criação de template não trouxe os dados esperados.');
  }
  return { template, version };
}

export async function listTemplates(params?: {
  limit?: number;
  offset?: number;
  doc_type?: string;
}): Promise<{ templates: TemplateDTO[]; meta: { limit: number; offset: number } }> {
  const qs = new URLSearchParams();
  if (params?.limit !== undefined) qs.set('limit', String(params.limit));
  if (params?.offset !== undefined) qs.set('offset', String(params.offset));
  if (params?.doc_type) qs.set('doc_type', params.doc_type);

  const suffix = qs.toString() ? `?${qs.toString()}` : '';
  const body = await apiFetch<ListTemplatesResponse>(`/api/v1/templates${suffix}`);

  const templates = body.data.templates as TemplateDTO[];

  const defaultLimit = params?.limit ?? 50;
  const defaultOffset = params?.offset ?? 0;
  const limit = toFiniteNumber(body.meta.limit) ?? defaultLimit;
  const offset = toFiniteNumber(body.meta.offset) ?? defaultOffset;

  return { templates, meta: { limit, offset } };
}

export async function getTemplate(id: string): Promise<{ template: TemplateDTO; latest_version: VersionDTO }> {
  const body = await apiFetch<{ data: { template: TemplateDTO; latest_version: VersionDTO } }>(
    `/api/v1/templates/${id}`,
  );
  return body.data;
}

// F1.2 / ADR 0035 — flat typed bodies, no { data: { ... } } envelope.
export async function getVersion(templateId: string, n: number): Promise<VersionDTO> {
  return apiFetch<VersionDTO>(`/api/v1/templates/${templateId}/versions/${n}`);
}

export async function presignAutosave(
  templateId: string,
  versionNum: number,
): Promise<{ upload_url: string; storage_key: string; expires_at: string }> {
  return apiFetch<{ upload_url: string; storage_key: string; expires_at: string }>(
    `/api/v1/templates/${templateId}/versions/${versionNum}/autosave/presign`,
    { method: 'POST' },
  );
}

export async function commitAutosave(
  templateId: string,
  versionNum: number,
  expectedContentHash: string,
): Promise<VersionDTO> {
  return apiFetch<VersionDTO>(
    `/api/v1/templates/${templateId}/versions/${versionNum}/autosave/commit`,
    { method: 'POST', body: JSON.stringify({ expected_content_hash: expectedContentHash }) },
  );
}

async function sha256Hex(buf: ArrayBuffer): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', buf);
  return Array.from(new Uint8Array(digest)).map((b) => b.toString(16).padStart(2, '0')).join('');
}

export async function importTemplateDocx(
  templateId: string,
  versionNum: number,
  file: File,
): Promise<VersionDTO> {
  const buf = await file.arrayBuffer();
  const { upload_url } = await presignAutosave(templateId, versionNum);
  const upload = await fetch(upload_url, {
    method: 'PUT',
    headers: { 'content-type': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' },
    body: buf,
  });
  if (!upload.ok) {
    throw new Error(`Falha ao enviar DOCX: HTTP ${upload.status}`);
  }
  const hash = await sha256Hex(buf);
  return commitAutosave(templateId, versionNum, hash);
}

export async function getDocxURL(templateId: string, versionNum: number): Promise<string> {
  const body = await apiFetch<{ data: { url: string } }>(
    `/api/v1/templates/${templateId}/versions/${versionNum}/docx-url`,
  );
  return body.data.url;
}

export async function submitForReview(
  templateId: string,
  versionNum: number,
  idempotencyKey: string,
): Promise<VersionDTO> {
  const data = await apiFetch<{ data: { version: VersionDTO } }>(
    `/api/v1/templates/${templateId}/versions/${versionNum}/submit`,
    { method: 'POST', idempotencyKey },
  );
  return data.data.version;
}

export async function reviewVersion(
  templateId: string,
  versionNum: number,
  accept: boolean,
  idempotencyKey: string,
  reason?: string,
): Promise<VersionDTO> {
  const data = await apiFetch<{ data: { version: VersionDTO } }>(
    `/api/v1/templates/${templateId}/versions/${versionNum}/review`,
    { method: 'POST', idempotencyKey, body: JSON.stringify({ accept, reason: reason || '' }) },
  );
  return data.data.version;
}

export interface NextDraftRef {
  id: string;
  versionNumber: number;
}

export interface ApproveVersionResult {
  version: VersionDTO;
  nextDraft: NextDraftRef | null;
}

export async function approveVersion(
  templateId: string,
  versionNum: number,
  accept: boolean,
  idempotencyKey: string,
  reason?: string,
): Promise<ApproveVersionResult> {
  const data = await apiFetch<{
    data: {
      version: VersionDTO;
      next_draft?: { id: string; version_number: number } | null;
    };
  }>(`/api/v1/templates/${templateId}/versions/${versionNum}/approve`, {
    method: 'POST',
    idempotencyKey,
    body: JSON.stringify({ accept, reason: reason || '' }),
  });
  const raw = data.data.next_draft;
  const nextDraft: NextDraftRef | null =
    raw && raw.id && Number.isFinite(raw.version_number)
      ? { id: raw.id, versionNumber: raw.version_number }
      : null;
  return { version: data.data.version, nextDraft };
}

// Wire-format types (backend snake_case)
interface WirePlaceholder { id: string; name?: string; label: string; type: string; required: boolean; options?: string[]; regex?: string; min_number?: number; max_number?: number; min_date?: string; max_date?: string; max_length?: number; resolver_key?: string; visible_if?: { placeholder_id: string; op: string; value?: unknown }; }

function placeholderFromWire(w: WirePlaceholder): Placeholder {
  return {
    id: w.id,
    ...(w.name != null ? { name: w.name } : {}),
    label: w.label,
    type: w.type as Placeholder['type'],
    ...(w.required ? { required: true } : {}),
    ...(w.options ? { options: w.options } : {}),
    ...(w.regex != null ? { regex: w.regex } : {}),
    ...(w.min_number != null ? { minNumber: w.min_number } : {}),
    ...(w.max_number != null ? { maxNumber: w.max_number } : {}),
    ...(w.min_date != null ? { minDate: w.min_date } : {}),
    ...(w.max_date != null ? { maxDate: w.max_date } : {}),
    ...(w.max_length != null ? { maxLength: w.max_length } : {}),
    ...(w.resolver_key != null ? { resolverKey: w.resolver_key } : {}),
    ...(w.visible_if ? { visibleIf: { placeholderID: w.visible_if.placeholder_id, operator: w.visible_if.op as NonNullable<Placeholder['visibleIf']>['operator'], value: w.visible_if.value as string | undefined } } : {}),
  };
}

function placeholderToWire(p: Placeholder): WirePlaceholder {
  return {
    id: p.id,
    ...(p.name != null ? { name: p.name } : {}),
    label: p.label,
    type: p.type,
    required: p.required ?? false,
    ...(p.options ? { options: p.options } : {}),
    ...(p.regex != null ? { regex: p.regex } : {}),
    ...(p.minNumber != null ? { min_number: p.minNumber } : {}),
    ...(p.maxNumber != null ? { max_number: p.maxNumber } : {}),
    ...(p.minDate != null ? { min_date: p.minDate } : {}),
    ...(p.maxDate != null ? { max_date: p.maxDate } : {}),
    ...(p.maxLength != null ? { max_length: p.maxLength } : {}),
    ...(p.resolverKey != null ? { resolver_key: p.resolverKey } : {}),
    ...(p.visibleIf ? { visible_if: { placeholder_id: p.visibleIf.placeholderID, op: p.visibleIf.operator, value: p.visibleIf.value } } : {}),
  };
}

export class StaleLockVersionError extends Error {
  readonly code = 'CONCURRENT_MODIFICATION';
  constructor(message?: string) {
    super(message ?? 'Concurrent modification: the template schema changed since you loaded it.');
    this.name = 'StaleLockVersionError';
  }
}

// Project the editor's schema view from a version already loaded via getVersion.
// The version GET is the single source of truth (ADR 0035 flat body); deriving
// here avoids a second round-trip and removes any chance of response-shape drift.
export function deriveTemplateSchemas(
  version: VersionDTO,
): { schemas: TemplateSchemas; lockVersion: number } {
  const ph = version.placeholder_schema;
  return {
    schemas: {
      placeholders: Array.isArray(ph) ? ph.map(placeholderFromWire) : [],
      composition: null,
    },
    lockVersion: typeof version.lock_version === 'number' ? version.lock_version : 0,
  };
}

export async function putTemplateSchemas(
  templateId: string,
  versionNum: number,
  schemas: TemplateSchemas,
  expectedLockVersion: number,
): Promise<{ lockVersion: number }> {
  try {
    const body = await apiFetch<{ data?: { version?: { lock_version?: number } } }>(
      `/api/v1/templates/${templateId}/versions/${versionNum}/schema`,
      {
        method: 'PUT',
        body: JSON.stringify({
          metadata_schema: {},
          placeholder_schema: schemas.placeholders.map(placeholderToWire),
          expected_lock_version: expectedLockVersion,
        }),
      },
    );
    const next = body?.data?.version?.lock_version;
    return { lockVersion: typeof next === 'number' ? next : expectedLockVersion + 1 };
  } catch (err) {
    // The optimistic-lock conflict surfaces as RFC 9457 412/CONCURRENT_MODIFICATION
    // through the shared transport; re-raise as the typed domain error so the
    // editor can prompt a refresh.
    if (err instanceof ApiError && (err.code === 'CONCURRENT_MODIFICATION' || err.status === 412)) {
      throw new StaleLockVersionError(err.message);
    }
    throw err;
  }
}
