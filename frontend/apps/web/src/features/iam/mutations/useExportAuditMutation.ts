import { useMutation } from "@tanstack/react-query";
import { api } from "../../../lib/api/client";

type AuditExportFormat = "csv" | "jsonl";

type AuditFilter = {
  actorId?: string;
  action?: string;
  resourceType?: string;
  resourceId?: string;
  occurredAfter?: string;
  occurredBefore?: string;
  q?: string;
};

type ExportAuditBody = {
  format: AuditExportFormat;
  filter: AuditFilter;
};

export function useExportAuditMutation() {
  return useMutation({
    mutationFn: async (body: ExportAuditBody) => {
      const { data, error } = await api.POST("/audit/events/export", { body });
      if (error) throw error;
      return data;
    },
  });
}
