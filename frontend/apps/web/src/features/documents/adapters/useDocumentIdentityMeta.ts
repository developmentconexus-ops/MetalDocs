import { useAreasQuery } from '../queries/useAreasQuery';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import { useControlledDocumentDetailQuery } from '../../controlled-documents/queries/useControlledDocumentDetailQuery';
import { buildVisibilityLabel, resolveAreaLabel, resolveProfileLabel } from '../lib/documentDetailMeta';
import type { ArtifactMetaAbsenceReasons } from '../../shared/controlled-artifact/types';
import type { DocumentDetail } from '../api/documents';

/**
 * F-QA4-4 — the single place the document IDENTIFICAÇÃO facts (tipo / área /
 * visibilidade / páginas) are resolved from live query data.
 *
 * Sources, all generated DTOs (ADR 0035 — no hand-written `body.data.X`):
 * - `DocumentDetailResponse.profile_code_snapshot`      → taxonomy profile name
 * - `DocumentDetailResponse.process_area_code_snapshot` → taxonomy area name
 * - `ControlledDocument.visibility`                     → visibility label
 *   (GET /controlled-documents/{id}; same `document.view` capability the
 *   workspace already holds — see apps/api/cmd/metaldocs-api/permissions.go:228)
 * - `DocumentDetailResponse.current_revision_page_count` / `_file_size_bytes`
 *
 * Nothing here fabricates a value: when a fact genuinely does not exist for the
 * document's current state the label stays `null` and `absenceReasons` carries
 * the human explanation the panel renders as the em-dash tooltip.
 */

const REASON_PROFILE_MISSING =
  'Este documento não tem perfil registrado no snapshot da revisão.';
const REASON_AREA_MISSING =
  'Este documento não tem área responsável registrada no snapshot da revisão.';
const REASON_VISIBILITY_UNLINKED =
  'A visibilidade é registrada no documento controlado; esta revisão ainda não está vinculada a um.';
const REASON_VISIBILITY_LOADING =
  'Carregando a visibilidade registrada no documento controlado.';
const REASON_VISIBILITY_ERROR =
  'Não foi possível carregar a visibilidade do documento controlado. Recarregue a página.';
const REASON_PAGES_MISSING =
  'A contagem de páginas só existe depois que o PDF oficial da revisão é renderizado.';

export interface DocumentIdentityMeta {
  /** Taxonomy profile name; falls back to the raw snapshot code until the taxonomy list resolves. */
  profileLabel: string | null;
  /** Taxonomy area name; falls back to the raw snapshot code until the taxonomy list resolves. */
  areaLabel: string | null;
  /** Visibility scope label built from the controlled document. Null while loading / on error / when unlinked. */
  visibilityLabel: string | null;
  /** Per-field explanations for the values that are genuinely absent. */
  absenceReasons: ArtifactMetaAbsenceReasons;
}

export function useDocumentIdentityMeta(doc: DocumentDetail | null | undefined): DocumentIdentityMeta {
  const areasQuery = useAreasQuery();
  const profilesQuery = useProfilesQuery();

  const controlledDocumentId = doc?.controlled_document_id ?? '';
  // `enabled: Boolean(id)` inside the hook makes the empty-id case a no-op.
  const controlledDocumentQuery = useControlledDocumentDetailQuery(controlledDocumentId);

  const areas = areasQuery.data ?? [];
  const profiles = profilesQuery.data ?? [];

  const profileCode = doc?.profile_code_snapshot ?? '';
  const areaCode = doc?.process_area_code_snapshot ?? '';

  // resolveProfileLabel/resolveAreaLabel degrade to the raw code when the
  // taxonomy list has not resolved (or the code was archived out of it). The
  // code IS the real identifier, so this is a truthful degradation — identical
  // to the detail screen's behavior (useDocumentArtifact.ts).
  const profileLabel = profileCode ? resolveProfileLabel(profileCode, profiles) : null;
  const areaLabel = areaCode ? resolveAreaLabel(areaCode, areas) : null;

  const visibility = controlledDocumentQuery.data?.visibility ?? null;
  const visibilityLabel = visibility ? buildVisibilityLabel(visibility, areas) : null;

  const absenceReasons: ArtifactMetaAbsenceReasons = {};
  if (!profileLabel) absenceReasons.profileLabel = REASON_PROFILE_MISSING;
  if (!areaLabel) absenceReasons.areaLabel = REASON_AREA_MISSING;
  if (!visibilityLabel) {
    absenceReasons.visibilityLabel = !controlledDocumentId
      ? REASON_VISIBILITY_UNLINKED
      : controlledDocumentQuery.isError
        ? REASON_VISIBILITY_ERROR
        : REASON_VISIBILITY_LOADING;
  }
  if (doc && doc.current_revision_page_count == null) {
    absenceReasons.pageCount = REASON_PAGES_MISSING;
  }

  return { profileLabel, areaLabel, visibilityLabel, absenceReasons };
}
