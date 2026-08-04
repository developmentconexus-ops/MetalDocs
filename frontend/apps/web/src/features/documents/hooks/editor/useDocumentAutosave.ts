import { useCallback, useEffect, useRef, useState } from 'react';
import { presignAutosave, commitAutosave } from '../../api/documents';
import { deletePending, putPending, getAllPending } from './useIndexedDBRestore';

export type AutosaveStatus = 'idle' | 'dirty' | 'saving' | 'saved' | 'stale' | 'session_lost' | 'error';

export interface AutosaveArgs {
  documentID: string;
  sessionID: string;
  baseRevisionID: string;
  onAdvanceBase: (newRevisionID: string) => void;
  onArtifactMetadata?: (metadata: { fileSizeBytes?: number | null; pageCount?: number | null }) => void;
  onSessionLost: (reason: 'conflict.concurrent_modification' | 'conflict.generic' | 'force_released') => void;
}

const SYNC_DEBOUNCE_MS = 3_000;

function asAutosaveFormSnapshot(value: unknown): Record<string, unknown> | undefined {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? value as Record<string, unknown>
    : undefined;
}

async function sha256Hex(buf: ArrayBuffer): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', buf);
  return Array.from(new Uint8Array(digest)).map((b) => b.toString(16).padStart(2, '0')).join('');
}

export function useDocumentAutosave(args: AutosaveArgs) {
  const { documentID, sessionID, baseRevisionID, onAdvanceBase, onArtifactMetadata, onSessionLost } = args;
  const pending = useRef<ArrayBuffer | null>(null);
  const pendingHash = useRef<string>('');
  const pendingPageCount = useRef<number | null>(null);
  const formSnapshot = useRef<unknown>(null);
  const timer = useRef<number | null>(null);
  const [status, setStatus] = useState<AutosaveStatus>('idle');

  const flush = useCallback(async (): Promise<boolean> => {
    if (!pending.current) return true;
    setStatus('saving');
    const buf = pending.current;
    const hash = pendingHash.current;
    try {
      // Persist to IndexedDB BEFORE hitting network -- crash recovery.
      await putPending({
        document_id: documentID,
        session_id: sessionID,
        base_revision_id: baseRevisionID,
        content_hash: hash,
        buffer: buf,
        created_at: Date.now(),
      });
      const presigned = await presignAutosave(documentID, {
        session_id: sessionID,
        base_revision_id: baseRevisionID,
        content_hash: hash,
      });
      const uploadRes = await fetch(presigned.upload_url, {
        method: 'PUT',
        headers: { 'content-type': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' },
        body: buf,
      });
      if (!uploadRes.ok) {
        throw Object.assign(new Error('autosave upload failed'), {
          status: uploadRes.status,
          body: await uploadRes.text(),
        });
      }
      // Server re-computes content_hash from S3; client does NOT send a hash.
      const commit = await commitAutosave(documentID, {
        session_id: sessionID,
        pending_upload_id: presigned.pending_upload_id,
        form_data_snapshot: asAutosaveFormSnapshot(formSnapshot.current),
        page_count: pendingPageCount.current ?? undefined,
      });
      await deletePending(documentID, hash);
      pending.current = null; pendingHash.current = ''; pendingPageCount.current = null;
      onAdvanceBase(commit.revision_id);
      onArtifactMetadata?.({
        fileSizeBytes: commit.file_size_bytes ?? null,
        pageCount: commit.page_count ?? null,
      });
      setStatus('saved');
      return true;
    } catch (e: any) {
      if (e?.status === 409) {
        // presign/commit emit RFC 9457 Problem; ApiError.code carries the
        // canonical code. CONCURRENT_MODIFICATION = stale base; CONFLICT_ERROR =
        // session inactive / not holder.
        if (e?.code === 'conflict.concurrent_modification') { onSessionLost('conflict.concurrent_modification'); setStatus('stale'); return false; }
        if (e?.code === 'conflict.generic') {
          onSessionLost('conflict.generic'); setStatus('session_lost'); return false;
        }
      }
      if (e?.status === 410) {
        // UPLOAD_MISSING or UPLOAD_EXPIRED: the S3 object is gone.
        try { await deletePending(documentID, hash); } catch { /* ignore */ }
        pending.current = null; pendingHash.current = ''; pendingPageCount.current = null;
        setStatus('error'); return false;
      }
      if (e?.status === 422) {
        // VALIDATION_ERROR (server-side content-hash mismatch): discard local pending.
        try { await deletePending(documentID, hash); } catch { /* ignore */ }
        pending.current = null; pendingHash.current = ''; pendingPageCount.current = null;
        setStatus('error'); return false;
      }
      setStatus('error');
      return false;
    }
  }, [documentID, sessionID, baseRevisionID, onAdvanceBase, onArtifactMetadata, onSessionLost]);

  const schedule = useCallback(() => {
    if (timer.current) window.clearTimeout(timer.current);
    timer.current = window.setTimeout(flush, SYNC_DEBOUNCE_MS);
  }, [flush]);

  const queue = useCallback(async (buf: ArrayBuffer, snapshot: unknown, pageCount?: number | null) => {
    pending.current = buf;
    formSnapshot.current = snapshot;
    pendingPageCount.current = pageCount ?? null;
    pendingHash.current = await sha256Hex(buf);
    setStatus('dirty');
    schedule();
  }, [schedule]);

  // Recovery on mount: if IndexedDB has a pending blob not yet committed,
  // replay it (if session matches base we still hold).
  useEffect(() => {
    (async () => {
      const leftovers = await getAllPending(documentID);
      for (const p of leftovers) {
        if (p.session_id !== sessionID) continue;
        pending.current = p.buffer;
        pendingHash.current = p.content_hash;
        await flush();
      }
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [documentID, sessionID]);

  useEffect(() => () => { if (timer.current) window.clearTimeout(timer.current); }, []);

  return { status, queue, flush };
}
