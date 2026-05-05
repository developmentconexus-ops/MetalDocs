import { request } from "../../../lib/api/client";
import type { DocumentProfileItem, ProcessAreaItem } from "../../../lib/types";

type TaxonomyProfileItem = Pick<DocumentProfileItem, "code" | "name" | "description" | "familyCode"> & {
  archived?: boolean;
};

type TaxonomyAreaItem = Pick<ProcessAreaItem, "code" | "name" | "description"> & {
  archived?: boolean;
};

function normalizeDocumentProfile(value: Partial<DocumentProfileItem>): DocumentProfileItem {
  const fallbackName = value.name ?? value.code ?? "";
  return {
    code: value.code ?? "",
    familyCode: value.familyCode ?? "",
    name: fallbackName,
    alias: value.alias?.trim?.() || fallbackName,
    description: value.description ?? "",
    reviewIntervalDays: Number(value.reviewIntervalDays ?? 0),
    activeSchemaVersion: Number(value.activeSchemaVersion ?? 0),
    workflowProfile: value.workflowProfile ?? "",
    approvalRequired: Boolean(value.approvalRequired),
    retentionDays: Number(value.retentionDays ?? 0),
    validityDays: Number(value.validityDays ?? 0),
  };
}

function normalizeProcessArea(value: ProcessAreaItem): ProcessAreaItem {
  return {
    code: value?.code ?? "",
    name: value?.name ?? value?.code ?? "",
    description: value?.description ?? "",
  };
}

export async function listTaxonomyProfiles(): Promise<{ items: DocumentProfileItem[] }> {
  const response = await request<{ items: TaxonomyProfileItem[] }>("/api/v2/taxonomy/profiles");
  return {
    items: Array.isArray(response.items)
      ? response.items
          .filter((item) => item.archived !== true)
          .map((item) =>
            normalizeDocumentProfile({
              code: item.code,
              familyCode: item.familyCode,
              name: item.name,
              description: item.description,
            }),
          )
      : [],
  };
}

export async function listTaxonomyAreas(): Promise<{ items: ProcessAreaItem[] }> {
  const response = await request<{ items: TaxonomyAreaItem[] }>("/api/v2/taxonomy/areas");
  return {
    items: Array.isArray(response.items)
      ? response.items.filter((item) => item.archived !== true).map(normalizeProcessArea)
      : [],
  };
}
