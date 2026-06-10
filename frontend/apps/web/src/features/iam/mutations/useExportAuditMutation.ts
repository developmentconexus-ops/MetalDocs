import { useMutation } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";
import type { components } from "../../../lib/api-types";

// Derived from the generated contract so the SERIALIZED body keys are exactly
// what the handler reads (filter.actor_id/resource_type/occurred_after/…). A
// hand-rolled camelCase shape here would launder past tsc and be silently
// dropped by the snake-reading Go decoder — the bug class this rename fixed.
export type AuditExportFilter = components["schemas"]["AuditFilter"];
export type ExportAuditBody = components["schemas"]["AuditExportRequest"];

export function useExportAuditMutation() {
  return useMutation({
    mutationFn: async (body: ExportAuditBody) => {
      const { data, error } = await api.POST("/audit/events/export", { body });
      if (error) throw error;
      return data;
    },
  });
}
