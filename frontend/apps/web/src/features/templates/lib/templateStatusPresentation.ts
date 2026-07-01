import type { VersionStatus } from '../api/templates';

const TEMPLATE_STATUS_LABELS: Record<VersionStatus, string> = {
  draft: 'rascunho',
  under_review: 'em revisão',
  approved: 'aprovado',
  published: 'publicado',
  obsolete: 'obsoleto',
};

/** pt-BR label for a template version status; echoes the raw value for unknown inputs. */
export function templateStatusLabel(status: string): string {
  return TEMPLATE_STATUS_LABELS[status as VersionStatus] ?? status;
}
