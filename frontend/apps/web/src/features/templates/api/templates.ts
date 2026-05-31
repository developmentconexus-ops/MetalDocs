import { apiFetch } from '../../../lib/api/client';
import type { components, paths } from '../../../lib/api-types';
import type { Placeholder, CompositionConfig } from '../placeholder-types';
export type { Placeholder, CompositionConfig };

type CreateTemplateRequest =
  paths['/api/v1/templates']['post']['requestBody']['content']['application/json'];
type CreateTemplateResponse =
  paths['/api/v1/templates']['post']['responses'][201]['content']['application/json'];
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
  placeholder_schema: Record<string, unknown> | null;
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

export interface PublishError {
  valid: false;
  parse_errors: Array<{ type: string; element?: string; ident?: string }>;
  missing_tokens: string[];
  orphan_tokens: string[];
}

export interface PublishSuccess {
  published_version_id: string;
  next_draft_id: string;
  next_draft_version_num: number;
}

async function apiJson<T>(res: Response): Promise<T> {
  if (!res.ok) {
    try {
      const body = (await res.json()) as { error?: { code?: string; message?: string } };
      const message = body?.error?.message;
      throw new Error(message || `HTTP ${res.status}`);
    } catch (err) {
      if (err instanceof Error) {
        throw err;
      }
      throw new Error(`HTTP ${res.status}`);
    }
  }

  return (await res.json()) as T;
}

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
  const body = await apiFetch<{
    data?: { templates?: unknown; items?: unknown };
    meta?: { limit?: unknown; offset?: unknown };
  }>(`/api/v1/templates${suffix}`);

  const dataTemplates = body?.data?.templates;
  const dataItems = body?.data?.items;
  const templates = Array.isArray(dataTemplates)
    ? (dataTemplates as TemplateDTO[])
    : Array.isArray(dataItems)
      ? (dataItems as TemplateDTO[])
      : [];

  const defaultLimit = params?.limit ?? 50;
  const defaultOffset = params?.offset ?? 0;
  const limit = toFiniteNumber(body?.meta?.limit) ?? defaultLimit;
  const offset = toFiniteNumber(body?.meta?.offset) ?? defaultOffset;

  return { templates, meta: { limit, offset } };
}

export async function getTemplate(id: string): Promise<{ template: TemplateDTO; latest_version: VersionDTO }> {
  const res = await fetch(`/api/v1/templates/${id}`);
  const body = await apiJson<{ data: { template: TemplateDTO; latest_version: VersionDTO } }>(res);
  return body.data;
}

export async function getVersion(templateId: string, n: number): Promise<VersionDTO> {
  const res = await fetch(`/api/v1/templates/${templateId}/versions/${n}`);
  const body = await apiJson<{ data: { version: VersionDTO } }>(res);
  return body.data.version;
}

export async function presignAutosave(
  templateId: string,
  versionNum: number,
): Promise<{ upload_url: string; storage_key: string; expires_at: string }> {
  const res = await fetch(`/api/v1/templates/${templateId}/versions/${versionNum}/autosave/presign`, {
    method: 'POST',
  });
  const body = await apiJson<{
    data: { upload_url: string; storage_key: string; expires_at: string };
  }>(res);
  return body.data;
}

export async function commitAutosave(
  templateId: string,
  versionNum: number,
  expectedContentHash: string,
): Promise<VersionDTO> {
  const res = await fetch(`/api/v1/templates/${templateId}/versions/${versionNum}/autosave/commit`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ expected_content_hash: expectedContentHash }),
  });
  const body = await apiJson<{ data: { version: VersionDTO } }>(res);
  return body.data.version;
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

export async function presignDocxUpload(
  templateId: string,
  versionNum: number,
): Promise<{ url: string; storage_key: string }> {
  const r = await presignAutosave(templateId, versionNum);
  return { url: r.upload_url, storage_key: r.storage_key };
}

export async function presignSchemaUpload(
  templateId: string,
  versionNum: number,
): Promise<{ url: string; storage_key: string }> {
  const r = await presignAutosave(templateId, versionNum);
  return { url: r.upload_url, storage_key: r.storage_key };
}

export async function saveDraft(
  templateId: string,
  versionNum: number,
  body: {
    expected_lock_version: number;
    docx_storage_key: string;
    schema_storage_key: string;
    docx_content_hash: string;
    schema_content_hash: string;
  },
): Promise<void> {
  await commitAutosave(templateId, versionNum, body.docx_content_hash || body.schema_content_hash);
}

export async function publishVersion(
  templateId: string,
  versionNum: number,
  docxKey: string,
  schemaKey: string,
): Promise<PublishSuccess | PublishError> {
  const res = await fetch(`/api/v1/templates/${templateId}/versions/${versionNum}/publish`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ docx_key: docxKey, schema_key: schemaKey }),
  });
  if (res.status === 422) {
    return (await res.json()) as PublishError;
  }
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  return (await res.json()) as PublishSuccess;
}

export async function getDocxURL(templateId: string, versionNum: number): Promise<string> {
  const res = await fetch(`/api/v1/templates/${templateId}/versions/${versionNum}/docx-url`);
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error((body as any)?.error?.message || `HTTP ${res.status}`);
  }
  const body = (await res.json()) as { data: { url: string } };
  return body.data.url;
}

export async function submitForReview(templateId: string, versionNum: number): Promise<VersionDTO> {
  const data = await apiFetch<{ data: { version: VersionDTO } }>(
    `/api/v1/templates/${templateId}/versions/${versionNum}/submit`,
    { method: 'POST' },
  );
  return data.data.version;
}

export async function reviewVersion(
  templateId: string,
  versionNum: number,
  accept: boolean,
  reason?: string,
): Promise<VersionDTO> {
  const res = await fetch(`/api/v1/templates/${templateId}/versions/${versionNum}/review`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ accept, reason: reason || '' }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error((body as any)?.error?.message || `HTTP ${res.status}`);
  }
  const data = (await res.json()) as { data: { version: VersionDTO } };
  return data.data.version;
}

export async function approveVersion(
  templateId: string,
  versionNum: number,
  accept: boolean,
  reason?: string,
): Promise<VersionDTO> {
  const res = await fetch(`/api/v1/templates/${templateId}/versions/${versionNum}/approve`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ accept, reason: reason || '' }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error((body as any)?.error?.message || `HTTP ${res.status}`);
  }
  const data = (await res.json()) as { data: { version: VersionDTO } };
  return data.data.version;
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
  readonly code = 'stale_lock_version';
  constructor(message?: string) {
    super(message ?? 'stale_lock_version');
    this.name = 'StaleLockVersionError';
  }
}

export async function getTemplateSchemas(
  templateId: string,
  versionNum: number,
): Promise<{ schemas: TemplateSchemas; lockVersion: number }> {
  const res = await fetch(`/api/v1/templates/${templateId}/versions/${versionNum}`);
  const body = await apiJson<{
    data: {
      version: VersionDTO & {
        placeholder_schema: WirePlaceholder[] | null;
        lock_version?: number;
      };
    };
  }>(res);
  const v = body.data.version;
  return {
    schemas: {
      placeholders: Array.isArray(v.placeholder_schema) ? v.placeholder_schema.map(placeholderFromWire) : [],
      composition: null,
    },
    lockVersion: typeof v.lock_version === 'number' ? v.lock_version : 0,
  };
}

export async function putTemplateSchemas(
  templateId: string,
  versionNum: number,
  schemas: TemplateSchemas,
  expectedLockVersion: number,
): Promise<{ lockVersion: number }> {
  const res = await fetch(`/api/v1/templates/${templateId}/versions/${versionNum}/schema`, {
    method: 'PUT',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({
      metadata_schema: {},
      placeholder_schema: schemas.placeholders.map(placeholderToWire),
      expected_lock_version: expectedLockVersion,
    }),
  });
  if (!res.ok) {
    let payload: { code?: string; title?: string; error?: { code?: string; message?: string } } = {};
    try {
      payload = await res.json();
    } catch {
      // fall through
    }
    const code = payload.code ?? payload.error?.code;
    if (code === 'stale_lock_version') {
      throw new StaleLockVersionError(payload.title ?? payload.error?.message);
    }
    throw new Error(payload.title ?? payload.error?.message ?? `HTTP ${res.status}`);
  }
  const body = (await res.json()) as { data?: { version?: { lock_version?: number } } };
  const next = body?.data?.version?.lock_version;
  return { lockVersion: typeof next === 'number' ? next : expectedLockVersion + 1 };
}
