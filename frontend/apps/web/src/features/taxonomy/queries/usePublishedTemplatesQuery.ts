import { useQuery } from '@tanstack/react-query';
import { listTemplates, type TemplateDTO, type VersionRef } from '../../templates/api/templates';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './constants';

/**
 * A TemplateDTO known to have a published version — narrowed by the type
 * guard below so downstream consumers (e.g. ProfileEditDialog) can read
 * `published_version.id` without a non-null assertion.
 */
export type PublishedTemplate = Omit<TemplateDTO, 'published_version'> & {
  published_version: VersionRef;
};

function hasPublishedVersion(t: TemplateDTO): t is PublishedTemplate {
  return t.published_version != null;
}

/**
 * Published templates, for the profile default-template picker. Filters to
 * published_version != null client-side — the templates list endpoint has
 * no "published only" server filter. ADR 0065: gates on the whole nested
 * VersionRef object, never an inner field.
 */
export function usePublishedTemplatesQuery(enabled: boolean) {
  return useQuery({
    queryKey: QK.templates.list(),
    queryFn: async () => {
      const { templates } = await listTemplates();
      return templates.filter(hasPublishedVersion);
    },
    enabled,
    staleTime: STALE_FIVE_MINUTES,
  });
}
