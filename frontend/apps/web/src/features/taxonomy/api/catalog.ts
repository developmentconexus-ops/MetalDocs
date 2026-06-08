import { request } from "../../../lib/api/client";
import type { DocumentProfileItem } from "../../../lib/types";
import type { components } from "../../../lib/api-types";

// Wire shape from GET /taxonomy/profiles (snake_case per contract) + the
// runtime-only `archived` flag not yet declared in the spec.
type TaxonomyProfileItem = Pick<
  components["schemas"]["DocumentProfileItem"],
  "code" | "name" | "description" | "family_code"
> & {
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

export async function listTaxonomyProfiles(): Promise<{ items: DocumentProfileItem[] }> {
  const response = await request<{ items: TaxonomyProfileItem[] }>("/api/v1/taxonomy/profiles");
  return {
    items: Array.isArray(response.items)
      ? response.items
          .filter((item) => item.archived !== true)
          .map((item) =>
            normalizeDocumentProfile({
              code: item.code,
              familyCode: item.family_code,
              name: item.name,
              description: item.description,
            }),
          )
      : [],
  };
}

