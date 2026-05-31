import { useCallback, useEffect, useState } from 'react';
import {
  getTemplateSchemas,
  putTemplateSchemas,
  StaleLockVersionError,
  type TemplateSchemas,
} from '../api/templates';

interface UseTemplateSchemasResult {
  schemas: TemplateSchemas | null;
  loading: boolean;
  error: string | null;
  save: (s: TemplateSchemas) => Promise<void>;
  saving: boolean;
  staleConflict: boolean;
  refetch: () => Promise<void>;
}

export function useTemplateSchemas(templateId: string, versionNum: number): UseTemplateSchemasResult {
  const [schemas, setSchemas] = useState<TemplateSchemas | null>(null);
  const [lockVersion, setLockVersion] = useState<number>(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [staleConflict, setStaleConflict] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const { schemas: s, lockVersion: lv } = await getTemplateSchemas(templateId, versionNum);
      setSchemas(s);
      setLockVersion(lv);
      setStaleConflict(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  }, [templateId, versionNum]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    getTemplateSchemas(templateId, versionNum)
      .then(({ schemas: s, lockVersion: lv }) => {
        if (cancelled) return;
        setSchemas(s);
        setLockVersion(lv);
        setStaleConflict(false);
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [templateId, versionNum]);

  const save = useCallback(
    async (s: TemplateSchemas) => {
      setSaving(true);
      try {
        const { lockVersion: next } = await putTemplateSchemas(templateId, versionNum, s, lockVersion);
        setSchemas(s);
        setLockVersion(next);
        setStaleConflict(false);
      } catch (e) {
        if (e instanceof StaleLockVersionError) {
          setStaleConflict(true);
        }
        throw e;
      } finally {
        setSaving(false);
      }
    },
    [templateId, versionNum, lockVersion],
  );

  return { schemas, loading, error, save, saving, staleConflict, refetch: load };
}
