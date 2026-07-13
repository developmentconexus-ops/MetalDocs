import { useState } from "react";
import { useParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { ArtifactApprovalScreen } from "../../shared/controlled-artifact/ArtifactApprovalScreen";
import {
  useTemplateApprovalArtifact,
  type TemplateApprovalHandlers,
} from "../adapters/useTemplateApprovalArtifact";
import { useTemplateDetailQuery } from "../queries/useTemplateDetailQuery";
import { publishVersion, signoffVersion, type VersionStatus } from "../api/templates";
import { TemplateReviewCanvas } from "../components/TemplateReviewCanvas";
import { TemplateApprovalExtras } from "../components/TemplateApprovalExtras";
import { QK } from "../../../lib/queryKeys";
import { resolveQueryError } from "../../../lib/api";
import type { ArtifactViewModel } from "../../shared/controlled-artifact/types";
import styles from "./TemplateApprovalRoute.module.css";

/**
 * Template-specific route wrapper for the shared ArtifactApprovalScreen.
 *
 * Kernel model: the DecisionPanel owns the under_review signoff (accept/reject
 * + motivo + senha, POST .../signoff). Publish (approved -> published) has no
 * signature ceremony and runs through the plain `actions[]` band instead
 * (Publicar, POST .../publish). Draft submission lives solely in the template
 * editor (R1) — this cockpit yields no actions for draft versions. Reads the
 * working version_number from the detail query.
 *
 * The main slot renders the template content read-only (TemplateReviewCanvas).
 * When a decision is offered the DecisionPanel owns accept/reject + the motivo
 * + senha; otherwise the sidebar extras surface status + the fire-and-forget
 * result message.
 */
export function TemplateApprovalRoute() {
  const { templateId = "" } = useParams<{ templateId: string }>();
  const queryClient = useQueryClient();

  const detailQuery = useTemplateDetailQuery(templateId);
  const versionNum = detailQuery.data?.latest_version.version_number ?? null;
  const status: VersionStatus | null = detailQuery.data?.latest_version.status ?? null;

  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<{ kind: "success" | "error"; text: string } | null>(null);

  // Fire-and-forget runner for the plain fallback actions (publish) — surfaces
  // its outcome as a sidebar message.
  const runAction = async (fn: () => Promise<unknown>, successText: string) => {
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

  // Awaitable runner for the DecisionPanel: the panel owns the in-flight/error UI
  // and re-throws a resolved message string for its error banner. The decision
  // always disappears after a signoff either way (status leaves under_review on
  // both approve and reject), so this also sets the sidebar message the fallback
  // extras then surfaces once the panel unmounts.
  const runDecision = async (fn: () => Promise<unknown>, successText: string) => {
    try {
      await fn();
    } catch (err) {
      throw new Error(resolveQueryError(err, "Falha ao registrar a decisão."));
    }
    await queryClient.invalidateQueries({ queryKey: QK.templates.detail(templateId) });
    void queryClient.invalidateQueries({ queryKey: QK.templates.list() });
    setMessage({ kind: "success", text: successText });
  };

  const handlers: TemplateApprovalHandlers = {
    // Unreachable in practice: whenever canSignoff allows this pair, the adapter
    // offers `model.decision` instead and zeroes `actions[]` (see
    // useTemplateApprovalArtifact). This only exists so the disabled fallback
    // button (gate denied, no decision offered either) has somewhere to point.
    runSignoff: (_accept) => {},
    runPublish: () => {
      if (versionNum == null) return;
      void runAction(() => publishVersion(templateId, versionNum, crypto.randomUUID()), "Publicado.");
    },
  };

  const decisionSubmit = async (accept: boolean, reason: string, password: string) => {
    if (versionNum == null) return;
    await runDecision(
      () =>
        signoffVersion(
          templateId,
          versionNum,
          { decision: accept ? "approve" : "reject", reason: reason || undefined, password },
          crypto.randomUUID(),
        ),
      accept ? "Assinatura registrada." : "Rejeitado — volta para rascunho.",
    );
  };

  const { model, isLoading, isError } = useTemplateApprovalArtifact(templateId, handlers, decisionSubmit);

  if (isLoading) {
    return <div className={styles.state} role="status" aria-live="polite">Carregando aprovação…</div>;
  }
  if (isError || versionNum == null) {
    return <div className={styles.state} role="alert">Não foi possível carregar esta versão do modelo.</div>;
  }

  const decision = model.decision;

  // The adapter already emits actions=[] whenever a decision is offered (single
  // owner of that invariant); here we only disable the plain actions while busy.
  const actions = busy ? model.actions.map((a) => ({ ...a, available: false })) : model.actions;
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
