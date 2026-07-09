// F2d.3 — the single FE workspace-mode selector. Pure, side-effect-free.
//
// The screen (F2d.5) and DecisionFooter render exactly ONE mode; they call
// `deriveWorkspaceMode` and switch on the result. The frontend never derives
// eligibility itself — it renders server-derived facts (ADR 0078). Eligibility
// truth is `viewer.eligible_for_active_stage`; enforcement stays at tier-2
// authz.Require + the write-path CheckEligibility/CheckSoD.
//
// Two sub-derivations:
//  - deriveStageMode — SUBJECT-AGNOSTIC (only stage_kind × the eligibility bool);
//    survives the M3 kernel extraction / templates reuse, knows nothing of documents.
//  - the lifecycle/author branch — subject-specific, keyed on document status plus
//    instance.status (the changes_requested split: a request_changes verdict reverts
//    the document to `draft` while the instance stays `changes_requested`).

import type { components } from '../../../lib/api-types';

type Instance = components['schemas']['ApprovalInstanceByDocumentResponse'];
type Viewer = Instance['viewer'];

export type WorkspaceMode =
  | 'author-editing'
  | 'author-waiting'
  | 'author-changes-requested'
  | 'reviewing'
  | 'approving'
  | 'observing'
  | 'lifecycle';

// Document statuses at/after approval where the lifecycle surface (publish /
// schedule / supersede) is the relevant action. The publish CTA itself stays
// capability-gated (canPublish in useDocumentArtifact) — the mode only decides
// which panel the screen shows.
const TERMINAL_LIFECYCLE_STATUSES = new Set([
  'approved',
  'scheduled',
  'published',
  'superseded',
  'obsolete',
]);

/**
 * Subject-agnostic stage-mode: what a caller sees on the active stage, from the
 * stage kind and the server eligibility bool alone. No document/subject input.
 */
export function deriveStageMode(
  stageKind: 'review' | 'approval' | undefined,
  eligibleForActiveStage: boolean,
): 'reviewing' | 'approving' | 'observing' {
  if (!eligibleForActiveStage) return 'observing';
  if (stageKind === 'review') return 'reviewing';
  if (stageKind === 'approval') return 'approving';
  return 'observing';
}

/**
 * The single workspace-mode selector. Precedence:
 *  1. is_author → the lifecycle/author lens (SoD guarantees an author is never
 *     stage-eligible, so the author never sees reviewing/approving).
 *  2. non-author with an active stage → deriveStageMode.
 *  3. otherwise → lifecycle for terminal document statuses, else observing.
 */
export function deriveWorkspaceMode(
  doc: { status: string } | null,
  instance: Instance | null,
  viewer: Viewer | null,
): WorkspaceMode {
  const docStatus = doc?.status ?? 'draft';

  if (viewer?.is_author) {
    if (instance?.status === 'changes_requested') return 'author-changes-requested';
    if (docStatus === 'under_review') return 'author-waiting';
    if (TERMINAL_LIFECYCLE_STATUSES.has(docStatus)) return 'lifecycle';
    return 'author-editing'; // draft, or no instance yet
  }

  const activeStage = instance?.stages?.find((s) => s.status === 'active');
  if (activeStage) {
    return deriveStageMode(activeStage.stage_kind, viewer?.eligible_for_active_stage ?? false);
  }

  if (TERMINAL_LIFECYCLE_STATUSES.has(docStatus)) return 'lifecycle';
  return 'observing';
}
