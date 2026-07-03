import { apiFetch, ApiError } from "../../../lib/api";
import type { components, operations } from "../../../lib/api-types";
import type { ControlledDocument } from "../types";

const BASE = "/api/v1/controlled-documents";

// FE-03: response type derived from the generated operation instead of a
// hand-rolled `{ items: ControlledDocument[] }` inline type. The hand-rolled
// version silently dropped the `page` cursor field from its type (the field
// was present on the wire but invisible to callers/tsc). Verified against
// internal/modules/controlleddocuments/delivery/http/routes.go
// ListControlledDocuments, which writes ListControlledDocuments200JSONResponse
// { Items, Page: {NextCursor, HasMore} } — a flat, non-enveloped body matching
// operations['listControlledDocuments'] exactly.
type ListControlledDocumentsResponse =
  operations["listControlledDocuments"]["responses"][200]["content"]["application/json"];

export async function fetchControlledDocuments(filter?: {
  profileCode?: string;
  processAreaCode?: string;
  status?: string;
  limit?: number;
  cursor?: string;
}): Promise<ControlledDocument[]> {
  const params = new URLSearchParams();
  if (filter?.profileCode) params.set("profile_code", filter.profileCode);
  if (filter?.processAreaCode) params.set("process_area_code", filter.processAreaCode);
  if (filter?.status) params.set("status", filter.status);
  if (filter?.limit != null) params.set("limit", String(filter.limit));
  if (filter?.cursor) params.set("cursor", filter.cursor);
  const qs = params.toString() ? `?${params.toString()}` : "";
  // FD-2: response is { items, page:{ next_cursor, has_more } }; this helper
  // returns the first page's items (no paginated CD UI consumes the cursor yet).
  const res = await apiFetch<ListControlledDocumentsResponse>(`${BASE}${qs}`);
  return res.items;
}

export async function fetchControlledDocument(id: string): Promise<ControlledDocument> {
  return apiFetch<ControlledDocument>(`${BASE}/${encodeURIComponent(id)}`);
}

export type CreateAtomicRequest = components["schemas"]["CreateAtomicRequest"];
export type AtomicCreateResponse = components["schemas"]["AtomicCreateResponse"];

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

export type CreateRevisionRequest = components["schemas"]["CreateRevisionRequest"];
export type RevisionResponse = components["schemas"]["RevisionResponse"];

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

export type PreviewCodeResponse = components["schemas"]["PreviewCodeResponse"];

export async function previewCode(
  profileCode: string,
  areaCode: string,
): Promise<PreviewCodeResponse> {
  const qs = new URLSearchParams({ profile_code: profileCode, area_code: areaCode }).toString();
  return apiFetch<PreviewCodeResponse>(`${BASE}/preview-code?${qs}`);
}

export async function obsoleteControlledDocument(id: string, idempotencyKey: string): Promise<void> {
  await apiFetch<void>(`${BASE}/${encodeURIComponent(id)}/obsolete`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    idempotencyKey,
  });
}

export async function supersedeControlledDocument(id: string, idempotencyKey: string): Promise<void> {
  await apiFetch<void>(`${BASE}/${encodeURIComponent(id)}/supersede`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    idempotencyKey,
  });
}

export type ActiveDocumentResponse = components["schemas"]["ActiveDocumentResponse"];

export async function fetchActiveDocumentInstance(
  controlledDocumentId: string,
): Promise<components["schemas"]["ActiveDocumentResponse"] | null> {
  try {
    return await apiFetch<components["schemas"]["ActiveDocumentResponse"]>(
      `${BASE}/${encodeURIComponent(controlledDocumentId)}/active-document`,
    );
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}
