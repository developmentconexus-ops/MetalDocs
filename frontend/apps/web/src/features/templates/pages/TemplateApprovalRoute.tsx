import { useState } from "react";
import { useParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { ArtifactApprovalScreen } from "../../shared/controlled-artifact/ArtifactApprovalScreen";
import {
  useTemplateApprovalArtifact,
  type TemplateApprovalHandlers,
} from "../adapters/useTemplateApprovalArtifact";
import { useTemplateDetailQuery } from "../queries/useTemplateDetailQuery";
import {
  submitForReview,
  reviewVersion,
  approveVersion,
  type VersionDTO,
  type VersionStatus,
} from "../api/templates";
import { TemplateReviewCanvas } from "../components/TemplateReviewCanvas";
import { TemplateApprovalExtras } from "../components/TemplateApprovalExtras";
import { QK } from "../../../lib/queryKeys";
import { resolveQueryError } from "../../../lib/api";
import type { ArtifactViewModel } from "../../shared/controlled-artifact/types";
import styles from "./TemplateApprovalRoute.module.css";

/**
 * Template-specific route wrapper for the shared ArtifactApprovalScreen. Composes
 * the shared DecisionModel for the review/approve states (accept/reject radio + an
 * inline motivo, no password/legal — templates carry no legal e-signature) and
 * feeds the plain fallback actions (draft submit) for every other state. Reads the
 * working version_number from the detail query.
 *
 * The main slot renders the template content read-only (TemplateReviewCanvas). When
 * a decision is offered the DecisionPanel owns accept/reject + the motivo; otherwise
 * the sidebar extras surface status + the fire-and-forget result message.
 */
export function TemplateApprovalRoute() {
  const { templateId = "" } = useParams<{ templateId: string }>();
  const queryClient = useQueryClient();

  const detailQuery = useTemplateDetailQuery(templateId);
  const versionNum = detailQuery.data?.latest_version.version_number ?? null;
  const status: VersionStatus | null = detailQuery.data?.latest_version.status ?? null;

  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<{ kind: "success" | "error"; text: string } | null>(null);

  // Fire-and-forget runner for the plain fallback actions (draft submit) — surfaces
  // its outcome as a sidebar message.
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

  // Awaitable runner for the DecisionPanel: the panel owns the in-flight/error UI,
  // so this re-throws a resolved message string for the panel's error banner.
  const runDecision = async (fn: () => Promise<VersionDTO>) => {
    try {
      await fn();
    } catch (err) {
      throw new Error(resolveQueryError(err, "Falha ao registrar a decisão."));
    }
    await queryClient.invalidateQueries({ queryKey: QK.templates.detail(templateId) });
    void queryClient.invalidateQueries({ queryKey: QK.templates.list() });
  };

  const handlers: TemplateApprovalHandlers = {
    runSubmit: () => {
      if (versionNum == null) return;
      void runAction(() => submitForReview(templateId, versionNum, crypto.randomUUID()), "Enviado para revisão.");
    },
    runReview: (accept) => {
      if (versionNum == null) return;
      void runAction(
        () => reviewVersion(templateId, versionNum, accept, crypto.randomUUID()),
        accept ? "Revisão aprovada." : "Revisão rejeitada — volta para rascunho.",
      );
    },
    runApprove: (accept) => {
      if (versionNum == null) return;
      void runAction(
        () => approveVersion(templateId, versionNum, accept, crypto.randomUUID()),
        accept ? "Publicado." : "Rejeitado — volta para rascunho.",
      );
    },
  };

  // A decision is offered only when the actor can actually act on a review/approve
  // state — the adapter's `buildTemplateApprovalDecision` reuses the gate it already
  // computed (accept.available) rather than re-deriving capability logic here. This
  // route only supplies the raw submit call — it knows `versionNum` + which of
  // reviewVersion/approveVersion applies (FE-02: single decision-model construction
  // path, owned by the adapter/lib, not duplicated in the route).
  const underReview = status === "under_review";
  const hasReviewer = detailQuery.data?.latest_version.pending_reviewer_role != null;
  const isReview = underReview && hasReviewer;

  const decisionSubmit = async (accept: boolean, reason: string) => {
    if (versionNum == null) return;
    const call = isReview
      ? () => reviewVersion(templateId, versionNum, accept, crypto.randomUUID(), reason)
      : () => approveVersion(templateId, versionNum, accept, crypto.randomUUID(), reason);
    await runDecision(call);
  };

  const { model, isLoading, isError } = useTemplateApprovalArtifact(templateId, handlers, decisionSubmit);

  if (isLoading) {
    return <div className={styles.state} role="status" aria-live="polite">Carregando aprovação…</div>;
  }
  if (isError || versionNum == null) {
    return <div className={styles.state} role="alert">Não foi possível carregar esta versão do modelo.</div>;
  }

  const decision = model.decision;

  // When a decision is offered the panel owns accept/reject, so strip those from the
  // plain actions band. Otherwise disable the fallback buttons while busy.
  const actions = decision != null ? [] : busy ? model.actions.map((a) => ({ ...a, available: false })) : model.actions;
  const screenModel: ArtifactViewModel = { ...model, actions, decision };

  return (
    <ArtifactApprovalScreen
      model={screenModel}
      main={<TemplateReviewCanvas templateId={templateId} versionNum={versionNum} />}
      decisionExtras={
        decision != null ? undefined : (
          <TemplateApprovalExtras status={status} message={message} />
        )
      }
    />
  );
}

export { TemplateApprovalRoute as Component };
