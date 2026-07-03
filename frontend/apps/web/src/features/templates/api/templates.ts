import { api, apiFetch } from '../../../lib/api/client';
import { ApiError } from '../../../lib/api/errors';
import type { components, operations, paths } from '../../../lib/api-types';
import type { Placeholder, CompositionConfig } from '../placeholder-types';
export type { Placeholder, CompositionConfig };

type CreateTemplateRequest =
  paths['/templates']['post']['requestBody']['content']['application/json'];
type CreateTemplateResponse =
  paths['/templates']['post']['responses'][201]['content']['application/json'];
// Query params derived from the generated contract so the snake_case wire keys
// (limit/offset/doc_type) can never drift from the spec (F-C2).
type ListTemplatesQuery = NonNullable<operations['listTemplates']['parameters']['query']>;
// ADR 0065: nested version-ref value object (latest_version/published_version).
// Re-exported so consumers don't reach into the generated api-types directly.
export type VersionRef = components['schemas']['TemplateVersionRef'];
// FE-15: response envelopes derived from the generated operations so the
// envelope-vs-flat shape can never drift from api/openapi. Verified 1:1
// against the Go handlers (internal/modules/templates/delivery/http/*.go):
// getTemplate -> {data:{template,latest_version}}; getDocxURL -> {data:{url}};
// submit/review -> TemplateVersionEnvelope {data:{version}}; approve ->
// ApproveTemplateVersionResponse {data:{version}}; updateTemplateSchema ->
// {data:{version}}.
type GetTemplateResponse = operations['getTemplate']['responses'][200]['content']['application/json'];
type GetTemplateDocxUrlResponse =
  operations['getTemplateDocxUrl']['responses'][200]['content']['application/json'];
type TemplateVersionEnvelope =
  operations['submitTemplateVersion']['responses'][200]['content']['application/json'];
type ApproveTemplateVersionResponse =
  operations['approveTemplateVersion']['responses'][200]['content']['application/json'];
type UpdateTemplateSchemaResponse =
  operations['updateTemplateSchema']['responses'][200]['content']['application/json'];

export type TemplateDTO = components['schemas']['TemplateDTO'];
export type VersionDTO = components['schemas']['VersionDTO'];
export type VersionStatus = VersionDTO['status'];

export interface TemplateSchemas {
  placeholders: Placeholder[];
  composition: CompositionConfig | null;
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
  const { template, version } = body.data;
  if (!template || !version) {
    throw new Error('Resposta de criação de template não trouxe os dados esperados.');
  }
  return { template: template as TemplateDTO, version: version as VersionDTO };
}

export async function listTemplates(params?: {
  limit?: number;
  offset?: number;
  doc_type?: string;
  published?: boolean;
}): Promise<{ templates: TemplateDTO[]; meta: { limit: number; offset: number } }> {
  const query: ListTemplatesQuery = {};
  if (params?.limit !== undefined) query.limit = params.limit;
  if (params?.offset !== undefined) query.offset = params.offset;
  if (params?.doc_type) query.doc_type = params.doc_type;
  if (params?.published) query.published = true;

  const { data, error } = await api.GET('/templates', { params: { query } });
  if (error) {
    throw error instanceof ApiError
      ? error
      : new ApiError('templates.list_failed', 0, 'Falha ao listar templates.');
  }
  if (!data) {
    throw new ApiError('templates.empty_response', 0, 'Resposta vazia ao listar templates.');
  }

  const templates = data.data.templates as TemplateDTO[];
  return { templates, meta: { limit: data.meta.limit, offset: data.meta.offset } };
}

export async function getTemplate(id: string): Promise<{ template: TemplateDTO; latest_version: VersionDTO }> {
  const body = await apiFetch<GetTemplateResponse>(`/api/v1/templates/${id}`);
  return body.data as { template: TemplateDTO; latest_version: VersionDTO };
}

// F1.2 / ADR 0035 — flat typed bodies, no { data: { ... } } envelope.
export async function getVersion(templateId: string, n: number): Promise<VersionDTO> {
  return apiFetch<VersionDTO>(`/api/v1/templates/${templateId}/versions/${n}`);
}

// F1.2 / ADR 0035 — manual next-version: POST with empty body, flat VersionDTO response, no idempotency key.
// M1 removed auto-spawn on approve/publish; this is the only path to a new template version.
export async function createNextVersion(templateId: string): Promise<VersionDTO> {
  return apiFetch<VersionDTO>(`/api/v1/templates/${templateId}/versions`, { method: 'POST' });
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
  const body = await apiFetch<GetTemplateDocxUrlResponse>(
    `/api/v1/templates/${templateId}/versions/${versionNum}/docx-url`,
  );
  return body.data.url;
}

export async function submitForReview(
  templateId: string,
  versionNum: number,
  idempotencyKey: string,
): Promise<VersionDTO> {
  const data = await apiFetch<TemplateVersionEnvelope>(
    `/api/v1/templates/${templateId}/versions/${versionNum}/submit`,
    { method: 'POST', idempotencyKey },
  );
  return data.data.version as VersionDTO;
}

export async function reviewVersion(
  templateId: string,
  versionNum: number,
  accept: boolean,
  idempotencyKey: string,
  reason?: string,
): Promise<VersionDTO> {
  const data = await apiFetch<TemplateVersionEnvelope>(
    `/api/v1/templates/${templateId}/versions/${versionNum}/review`,
    { method: 'POST', idempotencyKey, body: JSON.stringify({ accept, reason: reason || '' }) },
  );
  return data.data.version as VersionDTO;
}

export async function approveVersion(
  templateId: string,
  versionNum: number,
  accept: boolean,
  idempotencyKey: string,
  reason?: string,
): Promise<VersionDTO> {
  const data = await apiFetch<ApproveTemplateVersionResponse>(
    `/api/v1/templates/${templateId}/versions/${versionNum}/approve`,
    {
      method: 'POST',
      idempotencyKey,
      body: JSON.stringify({ accept, reason: reason || '' }),
    },
  );
  return data.data.version as VersionDTO;
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
  // The generated DTO types placeholder_schema elements as an untyped JSON
  // object ({[key: string]: unknown}) because the spec can't statically
  // declare the placeholder shape; WirePlaceholder is the local wire-format
  // view type for this call site only (not an Omit<> override of the DTO).
  const ph = version.placeholder_schema as WirePlaceholder[] | null;
  return {
    schemas: {
      placeholders: Array.isArray(ph) ? ph.map(placeholderFromWire) : [],
      composition: null,
    },
    lockVersion: typeof version.lock_version === 'number' ? version.lock_version : 0,
  };
}

// Thrown when the PUT .../schema 200 response doesn't carry a numeric
// lock_version. FE-15: previously this case silently guessed
// `expectedLockVersion + 1`, which can mask real concurrent-write drift (the
// caller would believe its optimistic lock advanced by exactly one when the
// server's actual value could differ). Fail loud instead so the editor
// surfaces a hard error rather than silently trusting a fabricated CAS token.
export class SchemaSaveResponseShapeError extends Error {
  constructor(message?: string) {
    super(message ?? 'Resposta inesperada ao salvar o schema do template: lock_version ausente.');
    this.name = 'SchemaSaveResponseShapeError';
  }
}

export async function putTemplateSchemas(
  templateId: string,
  versionNum: number,
  schemas: TemplateSchemas,
  expectedLockVersion: number,
): Promise<{ lockVersion: number }> {
  try {
    const body = await apiFetch<UpdateTemplateSchemaResponse>(
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
    if (typeof next !== 'number') {
      throw new SchemaSaveResponseShapeError();
    }
    return { lockVersion: next };
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
