import { apiFetch, ApiError } from "../../../lib/api";
import type { ControlledDocument } from "../types";

const BASE = "/api/v2/controlled-documents";

export async function fetchControlledDocuments(filter?: {
  profileCode?: string;
  processAreaCode?: string;
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<ControlledDocument[]> {
  const params = new URLSearchParams();
  if (filter?.profileCode) params.set("profileCode", filter.profileCode);
  if (filter?.processAreaCode) params.set("processAreaCode", filter.processAreaCode);
  if (filter?.status) params.set("status", filter.status);
  if (filter?.limit != null) params.set("limit", String(filter.limit));
  if (filter?.offset != null) params.set("offset", String(filter.offset));
  const qs = params.toString() ? `?${params.toString()}` : "";
  const res = await apiFetch<{ items: ControlledDocument[] }>(`${BASE}${qs}`);
  return res.items;
}

export async function fetchControlledDocument(id: string): Promise<ControlledDocument> {
  return apiFetch<ControlledDocument>(`${BASE}/${encodeURIComponent(id)}`);
}

export interface CreateAtomicRequest {
  profileCode: string;
  processAreaCode: string;
  title: string;
  ownerUserId: string;
  documentName: string;
  templateVersionId?: string;
  formData?: Record<string, unknown>;
  manualCode?: string;
  manualCodeReason?: string;
}

export interface AtomicCreateResponse {
  controlledDocument: ControlledDocument;
  document: { id: string; contentHash: string };
}

export async function createControlledDocumentAtomic(
  req: CreateAtomicRequest,
  idempotencyKey: string,
): Promise<AtomicCreateResponse> {
  return apiFetch<AtomicCreateResponse>(BASE, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
    idempotencyKey,
  });
}

export interface CreateRevisionRequest {
  name: string;
  formData?: Record<string, unknown>;
  templateVersionId?: string;
}

export interface RevisionResponse {
  document: { id: string; contentHash: string };
}

export async function createRevision(
  cdID: string,
  req: CreateRevisionRequest,
  idempotencyKey: string,
): Promise<RevisionResponse> {
  return apiFetch<RevisionResponse>(
    `${BASE}/${encodeURIComponent(cdID)}/revisions`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(req),
      idempotencyKey,
    },
  );
}

export interface PreviewCodeResponse {
  profileCode: string;
  areaCode: string;
  nextSeq: number;
  code: string;
}

export async function previewCode(
  profileCode: string,
  areaCode: string,
): Promise<PreviewCodeResponse> {
  const qs = new URLSearchParams({ profileCode, areaCode }).toString();
  return apiFetch<PreviewCodeResponse>(`${BASE}/preview-code?${qs}`);
}

export async function obsoleteControlledDocument(id: string): Promise<void> {
  await apiFetch<void>(`${BASE}/${encodeURIComponent(id)}/obsolete`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
  });
}

export async function supersedeControlledDocument(id: string): Promise<void> {
  await apiFetch<void>(`${BASE}/${encodeURIComponent(id)}/supersede`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
  });
}

// All fields are optional — the backend may return only publishedDocumentId
// when the controlled document has no active draft (published-only state, E10).
export interface ActiveDocumentResponse {
  documentId?: string;
  approvalState?: string;
  contentHash?: string;
  revisionVersion?: number;
  publishedDocumentId?: string;
  approvalInstanceId?: string;
}

/** @deprecated use ActiveDocumentResponse */
export type ActiveDocumentInstance = ActiveDocumentResponse;

export async function fetchActiveDocumentInstance(
  controlledDocumentId: string,
): Promise<ActiveDocumentResponse | null> {
  try {
    return await apiFetch<ActiveDocumentResponse>(
      `${BASE}/${encodeURIComponent(controlledDocumentId)}/active-document`,
    );
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}
