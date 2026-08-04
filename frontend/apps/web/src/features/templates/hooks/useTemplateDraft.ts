import { useCallback, useEffect, useState } from 'react';
import { getVersion, getTemplate, getDocxURL, type VersionDTO, type TemplateDTO } from '../api/templates';

type DraftState = {
  loading: boolean;
  error: string | null;
  template: TemplateDTO | null;
  version: VersionDTO | null;
  docxBytes: ArrayBuffer | null;
  // Non-fatal: the version loaded but its .docx blob could not be fetched
  // (storage unreachable, presign expired, 5xx, or 404). The editor still
  // opens — this surfaces a retry banner. Under ADR 0088 every template
  // version is materialized with content at creation time (blank creation
  // copies the system blank asset, PRE-TX), so a 404 here is never "blank
  // canvas" — it means something is genuinely wrong (missing object) and
  // must surface like any other blob failure, regardless of version status.
  docxError: string | null;
};

type TemplateDraft = DraftState & {
  refetch: () => void;
};

export function useTemplateDraft(templateId: string, versionNum: number): TemplateDraft {
  const [state, setState] = useState<DraftState>({
    loading: true,
    error: null,
    template: null,
    version: null,
    docxBytes: null,
    docxError: null,
  });
  const [tick, setTick] = useState(0);
  const refetch = useCallback(() => setTick((t) => t + 1), []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const [template, version] = await Promise.all([getTemplate(templateId), getVersion(templateId, versionNum)]);

        // Loading the .docx blob is SEPARATE from loading the version. A blob
        // failure must NOT discard the version or blank the whole editor — that
        // was the masking bug where any storage hiccup became an infinite
        // "Carregando template..." spinner. The version is the source of truth;
        // the editor opens regardless, with a retry banner if the blob failed.
        let docxBytes: ArrayBuffer | null = null;
        let docxError: string | null = null;
        if (version.docx_storage_key) {
          try {
            const url = await getDocxURL(templateId, versionNum);
            const res = await fetch(url);
            if (res.ok) {
              docxBytes = await res.arrayBuffer();
            } else {
              // ADR 0088: every version is materialized with content at creation
              // time, so a missing blob (404 included) is a real failure for any
              // version status — surface the retry banner.
              docxError = `Não foi possível carregar o conteúdo do template (HTTP ${res.status}).`;
            }
          } catch (e) {
            docxError = e instanceof Error ? e.message : String(e);
          }
        }

        if (!cancelled) {
          setState({
            loading: false,
            error: null,
            template: template.template,
            version,
            docxBytes,
            docxError,
          });
        }
      } catch (e) {
        if (!cancelled) {
          setState((s) => ({
            ...s,
            loading: false,
            error: e instanceof Error ? e.message : String(e),
          }));
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [templateId, versionNum, tick]);

  return { ...state, refetch };
}
