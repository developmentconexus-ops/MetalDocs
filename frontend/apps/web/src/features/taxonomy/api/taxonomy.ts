import { api } from "../../../lib/api/client";
import { ApiError } from "../../../lib/api/errors";
import type { components } from "../../../lib/api-types";
import type {
  CreateAreaRequest,
  CreateFamilyRequest,
  CreateProfileRequest,
  DocumentFamily,
  DocumentProfile,
  ProcessArea,
  SetDefaultTemplateRequest,
  UpdateAreaRequest,
  UpdateFamilyRequest,
  UpdateProfileRequest,
} from "../types";

// Wire (snake_case) schemas straight from the generated contract — see
// lib/api/client.ts transport hierarchy doc block (FE-13): `api` is
// canonical for every contracted route.
type DocumentProfileItem = components["schemas"]["DocumentProfileItem"];
type ProcessAreaItem = components["schemas"]["ProcessAreaItem"];
type DocumentFamilyItem = components["schemas"]["DocumentFamilyItem"];
type ProfileUpsertRequest = components["schemas"]["TaxonomyProfileUpsertRequest"];
type AreaUpsertRequest = components["schemas"]["TaxonomyAreaUpsertRequest"];
type FamilyUpsertRequest = components["schemas"]["TaxonomyFamilyUpsertRequest"];

function asApiError(error: unknown, fallbackCode: string): ApiError {
  if (error instanceof ApiError) return error;
  const env = error as { code?: string; message?: string; details?: unknown } | null;
  return new ApiError(env?.code ?? fallbackCode, 0, env?.message ?? "Erro inesperado", env?.details);
}

function toDocumentProfile(item: DocumentProfileItem): DocumentProfile {
  return {
    code: item.code,
    familyCode: item.family_code,
    name: item.name,
    description: item.description,
    reviewIntervalDays: item.review_interval_days,
    defaultTemplateVersionId: null,
    ownerUserId: null,
    editableByRole: item.workflow_profile,
    archivedAt: null,
    createdAt: "",
  };
}

function toProcessArea(item: ProcessAreaItem): ProcessArea {
  return {
    code: item.code,
    tenantId: item.tenant_id,
    name: item.name,
    description: item.description,
    parentCode: item.parent_code,
    ownerUserId: item.owner_user_id,
    defaultApproverRole: item.default_approver_role,
    archivedAt: item.archived_at,
    createdAt: item.created_at,
  };
}

function toDocumentFamily(item: DocumentFamilyItem): DocumentFamily {
  return {
    code: item.code,
    name: item.name,
    description: item.description,
    isActive: item.is_active,
    createdAt: item.created_at,
  };
}

export async function fetchProfiles(includeArchived?: boolean): Promise<DocumentProfile[]> {
  const { data, error } = await api.GET("/taxonomy/profiles", {
    params: { query: includeArchived ? { include_archived: true } : undefined },
  });
  if (error) throw asApiError(error, "taxonomy.profiles.list_failed");
  return (data?.items ?? []).map(toDocumentProfile);
}

export async function fetchProfile(code: string): Promise<DocumentProfile> {
  const { data, error } = await api.GET("/taxonomy/profiles/{code}", {
    params: { path: { code } },
  });
  if (error) throw asApiError(error, "taxonomy.profiles.get_failed");
  if (!data) throw new ApiError("taxonomy.profiles.empty_response", 0, "Resposta vazia ao carregar perfil.");
  return toDocumentProfile(data);
}

export async function createProfile(req: CreateProfileRequest): Promise<DocumentProfile> {
  const body: ProfileUpsertRequest = {
    code: req.code,
    family_code: req.familyCode,
    name: req.name,
    description: req.description ?? "",
    review_interval_days: req.reviewIntervalDays,
    editable_by_role: req.editableByRole ?? "admin",
  };
  const { data, error } = await api.POST("/taxonomy/profiles", {
    params: { header: { "Idempotency-Key": crypto.randomUUID() } },
    body,
  });
  if (error) throw asApiError(error, "taxonomy.profiles.create_failed");
  if (!data) throw new ApiError("taxonomy.profiles.empty_response", 0, "Resposta vazia ao criar perfil.");
  return toDocumentProfile(data);
}

export async function updateProfile(code: string, req: UpdateProfileRequest): Promise<DocumentProfile> {
  const body: ProfileUpsertRequest = {
    code,
    family_code: req.familyCode,
    name: req.name ?? "",
    description: req.description ?? "",
    review_interval_days: req.reviewIntervalDays ?? 0,
    editable_by_role: req.editableByRole ?? "admin",
  };
  const { data, error } = await api.PATCH("/taxonomy/profiles/{code}", {
    params: { path: { code } },
    body,
  });
  if (error) throw asApiError(error, "taxonomy.profiles.update_failed");
  if (!data) throw new ApiError("taxonomy.profiles.empty_response", 0, "Resposta vazia ao atualizar perfil.");
  return toDocumentProfile(data);
}

export async function setDefaultTemplate(code: string, req: SetDefaultTemplateRequest): Promise<void> {
  const { error } = await api.PUT("/taxonomy/profiles/{code}/default-template", {
    params: { path: { code } },
    body: { template_version_id: req.templateVersionId },
  });
  if (error) throw asApiError(error, "taxonomy.profiles.set_default_template_failed");
}

export async function archiveProfile(code: string): Promise<void> {
  const { error } = await api.DELETE("/taxonomy/profiles/{code}", {
    params: { path: { code } },
  });
  if (error) throw asApiError(error, "taxonomy.profiles.archive_failed");
}

export async function fetchAreas(includeArchived?: boolean): Promise<ProcessArea[]> {
  const { data, error } = await api.GET("/taxonomy/areas", {
    params: { query: includeArchived ? { include_archived: true } : undefined },
  });
  if (error) throw asApiError(error, "taxonomy.areas.list_failed");
  return (data?.items ?? []).map(toProcessArea);
}

export async function fetchArea(code: string): Promise<ProcessArea> {
  const { data, error } = await api.GET("/taxonomy/areas/{code}", {
    params: { path: { code } },
  });
  if (error) throw asApiError(error, "taxonomy.areas.get_failed");
  if (!data) throw new ApiError("taxonomy.areas.empty_response", 0, "Resposta vazia ao carregar área.");
  return toProcessArea(data);
}

export async function createArea(req: CreateAreaRequest): Promise<ProcessArea> {
  const body: AreaUpsertRequest = {
    code: req.code,
    name: req.name,
    description: req.description ?? "",
    parent_code: req.parentCode ?? null,
    default_approver_role: req.defaultApproverRole ?? null,
  };
  const { data, error } = await api.POST("/taxonomy/areas", {
    params: { header: { "Idempotency-Key": crypto.randomUUID() } },
    body,
  });
  if (error) throw asApiError(error, "taxonomy.areas.create_failed");
  if (!data) throw new ApiError("taxonomy.areas.empty_response", 0, "Resposta vazia ao criar área.");
  return toProcessArea(data);
}

export async function updateArea(code: string, req: UpdateAreaRequest): Promise<ProcessArea> {
  const body: AreaUpsertRequest = {
    code,
    name: req.name ?? "",
    description: req.description ?? "",
    parent_code: req.parentCode ?? null,
    default_approver_role: req.defaultApproverRole ?? null,
  };
  const { data, error } = await api.PUT("/taxonomy/areas/{code}", {
    params: { path: { code } },
    body,
  });
  if (error) throw asApiError(error, "taxonomy.areas.update_failed");
  if (!data) throw new ApiError("taxonomy.areas.empty_response", 0, "Resposta vazia ao atualizar área.");
  return toProcessArea(data);
}

export async function archiveArea(code: string): Promise<void> {
  const { error } = await api.DELETE("/taxonomy/areas/{code}", {
    params: { path: { code } },
  });
  if (error) throw asApiError(error, "taxonomy.areas.archive_failed");
}

export async function fetchFamilies(includeInactive?: boolean): Promise<DocumentFamily[]> {
  const { data, error } = await api.GET("/taxonomy/families", {
    params: { query: includeInactive ? { include_inactive: true } : undefined },
  });
  if (error) throw asApiError(error, "taxonomy.families.list_failed");
  return (data?.items ?? []).map(toDocumentFamily);
}

export async function createFamily(req: CreateFamilyRequest): Promise<DocumentFamily> {
  const body: FamilyUpsertRequest = {
    code: req.code,
    name: req.name,
    description: req.description ?? "",
  };
  const { data, error } = await api.POST("/taxonomy/families", {
    params: { header: { "Idempotency-Key": crypto.randomUUID() } },
    body,
  });
  if (error) throw asApiError(error, "taxonomy.families.create_failed");
  if (!data) throw new ApiError("taxonomy.families.empty_response", 0, "Resposta vazia ao criar família.");
  return toDocumentFamily(data);
}

export async function updateFamily(code: string, req: UpdateFamilyRequest): Promise<DocumentFamily> {
  const body: FamilyUpsertRequest = {
    code,
    name: req.name,
    description: req.description ?? "",
  };
  const { data, error } = await api.PATCH("/taxonomy/families/{code}", {
    params: { path: { code } },
    body,
  });
  if (error) throw asApiError(error, "taxonomy.families.update_failed");
  if (!data) throw new ApiError("taxonomy.families.empty_response", 0, "Resposta vazia ao atualizar família.");
  return toDocumentFamily(data);
}

export async function deactivateFamily(code: string): Promise<void> {
  const { error } = await api.DELETE("/taxonomy/families/{code}", {
    params: { path: { code } },
  });
  if (error) throw asApiError(error, "taxonomy.families.deactivate_failed");
}
