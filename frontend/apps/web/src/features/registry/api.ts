import { apiFetch, ApiError } from "../../lib/api";
import type { ControlledDocument, CreateControlledDocumentRequest } from "./types";

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

export async function createControlledDocument(req: CreateControlledDocumentRequest): Promise<ControlledDocument> {
  return apiFetch<ControlledDocument>(BASE, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
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

export interface ActiveDocumentInstance {
  documentId: string;
  approvalState: string;
  contentHash: string;
  revisionVersion: number;
  publishedDocumentId?: string;
  approvalInstanceId?: string;
}

export async function fetchActiveDocumentInstance(
  controlledDocumentId: string,
): Promise<ActiveDocumentInstance | null> {
  try {
    return await apiFetch<ActiveDocumentInstance>(
      `${BASE}/${encodeURIComponent(controlledDocumentId)}/active-document`,
    );
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}
