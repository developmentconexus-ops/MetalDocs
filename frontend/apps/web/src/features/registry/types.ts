import type { components } from "../../lib/api-types";

export type ControlledDocument = components["schemas"]["ControlledDocument"];

export interface CreateControlledDocumentRequest {
  profileCode: string;
  processAreaCode: string;
  title: string;
  ownerUserId: string;
  departmentCode?: string;
  overrideTemplateVersionId?: string;
  overrideTemplateReason?: string;
  manualCode?: string;
  manualCodeReason?: string;
}
