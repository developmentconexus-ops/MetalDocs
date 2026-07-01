import { useState } from "react";
import { useParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { ArtifactApprovalScreen } from "../../shared/controlled-artifact/ArtifactApprovalScreen";
import {
  useTemplateApprovalArtifact,
  type TemplateApprovalHandlers,
} from "../adapters/useTemplateApprovalArtifact";
import { useTemplateDetailQuery } from "../queries/useTemplateDetailQuery";
import { submitForReview, reviewVersion, approveVersion, type VersionDTO } from "../api/templates";
import { TemplateReviewCanvas } from "../components/TemplateReviewCanvas";
import { TemplateApprovalExtras } from "../components/TemplateApprovalExtras";
import { QK } from "../../../lib/queryKeys";
import { resolveQueryError } from "../../../lib/api";
import type { DocumentStatus } from "../../../components/ui";
import styles from "./TemplateApprovalRoute.module.css";

/**
 * Template-specific route wrapper for the shared ArtifactApprovalScreen. Owns the
 * reason composer state + the busy/result message, and implements the three
 * TemplateApprovalHandlers (submit / review / approve) that the adapter binds to
 * the gated action buttons. Reads the working version_number from the detail query.
 *
 * The main slot renders the template content read-only (TemplateReviewCanvas); the
 * decision sidebar extras render the reason composer + status + result message.
 * No dialogs — template lifecycle decisions take an optional inline reason only.
 */
export function TemplateApprovalRoute() {
  const { templateId = "" } = useParams<{ templateId: string }>();
  const queryClient = useQueryClient();

  const detailQuery = useTemplateDetailQuery(templateId);
  const versionNum = detailQuery.data?.latest_version.version_number ?? null;
  const status = (detailQuery.data?.latest_version.status ?? null) as DocumentStatus | null;

  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<{ kind: "success" | "error"; text: string } | null>(null);

  const runAction = async (fn: () => Promise<VersionDTO>, successText: string) => {
    if (busy) return;
    setBusy(true);
    setMessage(null);
    try {
      await fn();
      await queryClient.invalidateQueries({ queryKey: QK.templates.detail(templateId) });
      void queryClient.invalidateQueries({ queryKey: QK.templates.list() });
      setMessage({ kind: "success", text: successText });
    } catch (err) {
      setMessage({ kind: "error", text: resolveQueryError(err, "Falha ao processar a ação.") });
    } finally {
      setBusy(false);
    }
  };

  const handlers: TemplateApprovalHandlers = {
    runSubmit: () => {
      if (versionNum == null) return;
      void runAction(() => submitForReview(templateId, versionNum, crypto.randomUUID()), "Enviado para revisão.");
    },
    runReview: (accept) => {
      if (versionNum == null) return;
      void runAction(
        () => reviewVersion(templateId, versionNum, accept, crypto.randomUUID(), reason),
        accept ? "Revisão aprovada." : "Revisão rejeitada — volta para rascunho.",
      );
    },
    runApprove: (accept) => {
      if (versionNum == null) return;
      void runAction(
        () => approveVersion(templateId, versionNum, accept, crypto.randomUUID(), reason),
        accept ? "Publicado." : "Rejeitado — volta para rascunho.",
      );
    },
  };

  const { model, isLoading, isError } = useTemplateApprovalArtifact(templateId, handlers);

  if (isLoading) {
    return <div className={styles.state} role="status" aria-live="polite">Carregando aprovação…</div>;
  }
  if (isError || versionNum == null) {
    return <div className={styles.state} role="alert">Não foi possível carregar esta versão do modelo.</div>;
  }

  // Disable all action buttons while an action is in flight.
  const screenModel = busy
    ? { ...model, actions: model.actions.map((a) => ({ ...a, available: false })) }
    : model;

  const showReason = status === "under_review" || status === "approved";

  return (
    <ArtifactApprovalScreen
      model={screenModel}
      main={<TemplateReviewCanvas templateId={templateId} versionNum={versionNum} />}
      decisionExtras={
        <TemplateApprovalExtras
          status={status}
          showReason={showReason}
          reason={reason}
          onReasonChange={setReason}
          busy={busy}
          message={message}
        />
      }
    />
  );
}

export { TemplateApprovalRoute as Component };
