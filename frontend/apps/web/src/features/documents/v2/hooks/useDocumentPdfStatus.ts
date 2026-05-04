import { useEffect, useRef, useState } from 'react';

export type PDFStatus = 'pending' | 'ready' | 'failed';
type ViewResponse = { pdf_status: PDFStatus; pdf_url?: string };

export type DocumentPdfStatus = {
  status: PDFStatus;
  url?: string;
  retry: () => void;
};

const POLL_INTERVAL_MS = 3_000;
const TIMEOUT_MS = 60_000;

export function useDocumentPdfStatus(documentID: string, enabled: boolean): DocumentPdfStatus {
  const [data, setData] = useState<{ status: PDFStatus; url?: string }>({ status: 'pending' });
  const [tick, setTick] = useState(0);
  const startedAt = useRef(0);

  useEffect(() => {
    if (!enabled || !documentID) return;
    let cancelled = false;
    let timer = 0;
    startedAt.current = Date.now();
    setData({ status: 'pending' });

    const poll = async () => {
      try {
        const res = await fetch(`/api/v2/documents/${encodeURIComponent(documentID)}/view`);
        if (cancelled) return;
        if (res.ok) {
          const v = (await res.json()) as ViewResponse;
          if (cancelled) return;
          setData({ status: v.pdf_status, url: v.pdf_url });
          if (v.pdf_status === 'ready' || v.pdf_status === 'failed') return;
        }
        if (Date.now() - startedAt.current > TIMEOUT_MS) {
          if (!cancelled) setData({ status: 'failed' });
          return;
        }
      } catch {
        // network glitch — retry next tick
        if (Date.now() - startedAt.current > TIMEOUT_MS) {
          if (!cancelled) setData({ status: 'failed' });
          return;
        }
      }
      if (!cancelled) {
        timer = window.setTimeout(poll, POLL_INTERVAL_MS);
      }
    };
    void poll();

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [documentID, enabled, tick]);

  return { ...data, retry: () => setTick((n) => n + 1) };
}
