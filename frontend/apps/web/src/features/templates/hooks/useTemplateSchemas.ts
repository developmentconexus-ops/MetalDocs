import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  deriveTemplateSchemas,
  putTemplateSchemas,
  StaleLockVersionError,
  type TemplateSchemas,
  type VersionDTO,
} from '../api/templates';

interface UseTemplateSchemasResult {
  schemas: TemplateSchemas | null;
  loading: boolean;
  error: string | null;
  save: (s: TemplateSchemas) => Promise<void>;
  saving: boolean;
  staleConflict: boolean;
}

/**
 * Editor schema state derived from the version that `useTemplateDraft` already
 * loaded — no second fetch of the same `/versions/{n}` resource. The version GET
 * is the single source of truth; this hook only owns the write path (optimistic
 * lock CAS via putTemplateSchemas). Recovery from a stale-lock 412 is a parent
 * concern: reload the draft (draft.refetch) and the fresh version re-seeds the
 * lock token and clears the conflict here.
 */
export function useTemplateSchemas(
  templateId: string,
  versionNum: number,
  version: VersionDTO | null,
): UseTemplateSchemasResult {
  const derived = useMemo(() => (version ? deriveTemplateSchemas(version) : null), [version]);
  const [lockVersion, setLockVersion] = useState(0);
  const [saving, setSaving] = useState(false);
  const [staleConflict, setStaleConflict] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // A freshly loaded version is server truth: re-seed the lock token and clear
  // any prior save conflict/error.
  useEffect(() => {
    if (!derived) return;
    setLockVersion(derived.lockVersion);
    setStaleConflict(false);
    setError(null);
  }, [derived]);

  const save = useCallback(
    async (s: TemplateSchemas) => {
      setSaving(true);
      try {
        const { lockVersion: next } = await putTemplateSchemas(templateId, versionNum, s, lockVersion);
        setLockVersion(next);
        setStaleConflict(false);
      } catch (e) {
        if (e instanceof StaleLockVersionError) setStaleConflict(true);
        else setError(e instanceof Error ? e.message : String(e));
        throw e;
      } finally {
        setSaving(false);
      }
    },
    [templateId, versionNum, lockVersion],
  );

  return {
    schemas: derived?.schemas ?? null,
    loading: version == null,
    error,
    save,
    saving,
    staleConflict,
  };
}
