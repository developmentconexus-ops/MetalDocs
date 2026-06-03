import { useState } from "react";
import { toast } from "sonner";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useExportAuditMutation } from "../mutations/useExportAuditMutation";
import styles from "./AuditExportButton.module.css";

type ExportFormat = "csv" | "jsonl";

export interface AuditExportFilter {
  actorId?: string;
  action?: string;
  resourceType?: string;
  resourceId?: string;
  occurredAfter?: string;
  occurredBefore?: string;
  q?: string;
}

interface AuditExportButtonProps {
  filter: AuditExportFilter;
}

interface ReadyJob {
  exportId: string;
  signedUrl: string;
  format: ExportFormat;
  expiresAt: string;
}

// Accept only safe download URLs: relative paths starting with "/" or
// same-origin https URLs. Reject javascript:, data:, blob:, and cross-origin
// http to neutralise any signed-URL injection in the backend response.
function safeDownloadUrl(url: string): string | undefined {
  if (!url) return undefined;
  if (url.startsWith("/") && !url.startsWith("//")) return url;
  try {
    const parsed = new URL(url, window.location.origin);
    if (parsed.protocol !== "https:") return undefined;
    if (parsed.origin !== window.location.origin) return undefined;
    return parsed.toString();
  } catch {
    return undefined;
  }
}

export default function AuditExportButton({ filter }: AuditExportButtonProps) {
  const [format, setFormat] = useState<ExportFormat>("csv");
  const [job, setJob] = useState<ReadyJob | null>(null);
  const mutation = useExportAuditMutation();

  const handleExport = () => {
    setJob(null);
    mutation.mutate(
      { format, filter },
      {
        onSuccess: (data) => {
          if (!data?.exportId || !data?.signedUrl) {
            toast.error("Resposta inesperada do servidor.");
            return;
          }
          const safeUrl = safeDownloadUrl(data.signedUrl);
          if (!safeUrl) {
            toast.error("URL de download inválida recebida do servidor.");
            return;
          }
          setJob({
            exportId: data.exportId,
            signedUrl: safeUrl,
            format,
            expiresAt: data.expiresAt,
          });
          toast.success("Exportação pronta para download.");
        },
        onError: (err) => {
          toast.error(
            resolveQueryError(err, "Falha ao solicitar exportação."),
          );
        },
      },
    );
  };

  return (
    <div className={styles.wrap} data-testid="audit-export-button">
      <select
        className={styles.format}
        aria-label="Formato da exportação"
        value={format}
        onChange={(e) => setFormat(e.target.value as ExportFormat)}
        disabled={mutation.isPending}
      >
        <option value="csv">CSV</option>
        <option value="jsonl">JSONL</option>
      </select>

      <button
        type="button"
        className={styles.button}
        onClick={handleExport}
        disabled={mutation.isPending}
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden="true"
        >
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <polyline points="7 10 12 15 17 10" />
          <line x1="12" y1="15" x2="12" y2="3" />
        </svg>
        {mutation.isPending ? "Em andamento…" : "Exportar"}
      </button>

      {mutation.isPending ? (
        <span className={styles.status} role="status" aria-live="polite">
          <span className={styles.spinner} aria-hidden="true" />
          Preparando arquivo
        </span>
      ) : null}

      {job ? (
        <a
          className={styles.download}
          href={job.signedUrl}
          download={`audit-export-${job.exportId}.${job.format}`}
          rel="noopener noreferrer"
        >
          Baixar {job.format.toUpperCase()}
        </a>
      ) : null}
    </div>
  );
}
