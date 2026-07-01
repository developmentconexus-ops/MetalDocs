/**
 * Kind-agnostic view-model boundary for the shared controlled-artifact view layer.
 *
 * Adapters (`useDocumentArtifact`, `useTemplateArtifact`) are responsible for
 * fetching kind-specific data and mapping it into these types. Components in the
 * controlled-artifact shell consume only these types — they have no knowledge of
 * document vs. template API shapes.
 *
 * Template-divergent fields are nullable so the shared shell can render a reduced
 * surface gracefully without branching on kind.
 */

import type { DocumentStatus } from "../../../components/ui/StatusPill";

// ---------------------------------------------------------------------------
// Core discriminators
// ---------------------------------------------------------------------------

export type ArtifactKind = "document" | "template";

/**
 * Unified lifecycle status vocabulary. Aliases the canonical `DocumentStatus`
 * union from StatusPill — do NOT fork a parallel union.
 */
export type LifecycleStatus = DocumentStatus;

// ---------------------------------------------------------------------------
// Hero / breadcrumb / badge supporting types
// ---------------------------------------------------------------------------

/** A single crumb in the artifact breadcrumb trail (data only — not ReactNode). */
export interface BreadcrumbItem {
  label: string;
  href?: string;
}

/** Visual classification of a badge displayed in the artifact hero area. */
export type ArtifactBadgeVariant = "code" | "status" | "type" | "neutral";

/** A single badge rendered in the hero area (data only — not ReactNode). */
export interface ArtifactBadge {
  /** Stable React key (e.g. "code", "status", "type") — avoids index-keyed reconciliation. */
  key: string;
  label: string;
  variant: ArtifactBadgeVariant;
}

/**
 * One cell in the artifact KPI strip. Adapters own the full content; the view is
 * purely presentational and renders a clickable navigation cell when `href` is set,
 * a static cell otherwise. This keeps every kind-specific KPI rule (e.g. the
 * document scheduled→published-head "current version" override, the live coverage
 * denominator) in the adapter, never in the shared view.
 */
export interface ArtifactKpiCell {
  /** Stable React key + test hook (e.g. "currentVersion", "coverage"). */
  key: string;
  /** Cell label (e.g. "Versão atual", "Cobertura"). */
  label: string;
  /** Primary value text (e.g. "REV08", "12", "—"). Always a string — adapters stringify. */
  value: string;
  /** Secondary hint line under the value. Null renders no hint. */
  hint: string | null;
  /** When set, the cell is a react-router navigation target (e.g. "/distribution"). */
  href?: string;
}

// ---------------------------------------------------------------------------
// ArtifactHeroModel
// ---------------------------------------------------------------------------

/**
 * Data for the top-of-page hero block: breadcrumb trail, status/type badges,
 * and an optional subtitle line.
 *
 * `subtitle` is nullable because templates have no regulated subtitle field.
 */
export interface ArtifactHeroModel {
  breadcrumb: BreadcrumbItem[];
  badges: ArtifactBadge[];
  /** Optional secondary descriptor rendered beneath the title. Null for templates. */
  subtitle: string | null;
}

// ---------------------------------------------------------------------------
// ArtifactMetaModel
// ---------------------------------------------------------------------------

/**
 * Profile / visibility / file-info metadata rendered in the artifact sidebar.
 *
 * All fields are nullable: documents may lack some metadata; templates have no
 * area, file size, page count, or effective / review dates.
 */
export interface ArtifactMetaModel {
  /** Human-readable label for the document profile (e.g. "Procedimento Operacional"). */
  profileLabel: string | null;
  /** Human-readable label for the process area (e.g. "Recursos Humanos"). Null for templates. */
  areaLabel: string | null;
  /** Human-readable label for the visibility scope (e.g. "Toda a organização"). */
  visibilityLabel: string | null;
  /** Raw byte size of the current revision file. Null for templates or when unknown. */
  fileSizeBytes: number | null;
  /** Page count of the current revision file. Null for templates or when unknown. */
  pageCount: number | null;
  /** ISO-8601 creation timestamp. */
  createdAt: string | null;
  /** ISO-8601 effective ("vigente desde") date. Null for templates. */
  effectiveFrom: string | null;
  /** ISO-8601 next scheduled review date. Null for templates. */
  nextReviewAt: string | null;
  /** Display name of the artifact owner/creator (resolved by the adapter, with
   *  current-user displayName fallback). Null when unknown. */
  ownerName: string | null;
  /** Governed owner-banner descriptor — DISTINCT from `hero.subtitle`. For documents
   *  this is the per-status copy (e.g. "substituído · publicado em 19 de maio de 2026").
   *  Null for kinds with no governed descriptor. */
  ownerDescriptor: string | null;
}

// ---------------------------------------------------------------------------
// ApprovalChainItem
// ---------------------------------------------------------------------------

/**
 * Normalized flow state for one approval-chain step, driving the shared
 * flow-viz dot rendering. Kind-agnostic — both the document instance mapper and
 * the template version mapper resolve their raw states into this union so the
 * timeline renders identically regardless of kind.
 *
 * - `approved` — step signed off / accepted (solid success dot with check).
 * - `rejected` — step returned / rejected (danger dot).
 * - `sent`     — step completed as a hand-off, not a signoff (e.g. author
 *                submitted for review — solid neutral dot).
 * - `current`  — the step awaiting the viewer's decision now (dashed brand ring).
 * - `pending`  — a future step not yet reached (hollow dot).
 */
export type ApprovalFlowState = "approved" | "rejected" | "sent" | "current" | "pending";

/**
 * A single step within the approval chain.
 *
 * Generalizes both `ApprovalInstance.stages[].signoffs[]` (documents) and the
 * inline submit/review/approve fields on a template `VersionDTO`. Templates are
 * NOT signoff-less: they carry the same submit→review→approve actors + timestamps
 * inline on the version, so both kinds populate this shape (the former
 * `approvalChain: null for templates` was an adapter shortcut, not a data gap).
 */
export interface ApprovalChainItem {
  /** Zero-based index of the parent approval stage. */
  stageIndex: number;
  /** Human-readable stage label (e.g. "Revisão", "Aprovação"). */
  label: string;
  /** Stage-level status string (mirrors the API `stage.status` values). */
  status: string;
  /** Short role descriptor shown under the actor (e.g. "Autora", "Revisão técnica"). Null when not applicable. */
  roleLabel: string | null;
  /** Normalized flow state driving the shared flow-viz dot. */
  flowState: ApprovalFlowState;
  /** User ID of the actor assigned to this signoff slot. Null when unassigned. */
  actorUserId: string | null;
  /** Display name of the actor. Null when unassigned. */
  actorDisplay: string | null;
  /** Decision string (e.g. "approved", "rejected"). Null when pending. */
  decision: string | null;
  /** ISO-8601 timestamp of the signoff action. Null when pending. */
  signedAt: string | null;
}

// ---------------------------------------------------------------------------
// VersionHistoryItem
// ---------------------------------------------------------------------------

/**
 * A single entry in the artifact revision history list.
 *
 * Generalizes document `revision-history` API items. `lineage` is always an
 * array (never null); use an empty array when no history is available.
 */
export interface VersionHistoryItem {
  versionNumber: number;
  /** Revision counter within the version. Null when not applicable. */
  revisionNumber: number | null;
  /** Formatted revision label (output of `formatRevisionCode`, e.g. "REV01"). Null when not applicable. */
  revisionLabel: string | null;
  status: LifecycleStatus;
  /** Title of this revision snapshot. Null when not recorded. */
  title: string | null;
  /** ISO-8601 creation timestamp of this revision. */
  createdAt: string | null;
  /** True when this entry represents the current active revision. */
  isCurrent: boolean;
}

// ---------------------------------------------------------------------------
// ArtifactTab
// ---------------------------------------------------------------------------

/**
 * A single tab in the artifact detail navigation bar.
 *
 * Documents drive tabs via router NavLinks (`href`); templates have a single
 * embedded tab with no navigation. `href` is optional to cover both cases.
 */
export interface ArtifactTab {
  key: string;
  label: string;
  href?: string;
  /** Optional badge count displayed on the tab (e.g. distribution recipients). */
  count?: number;
}

// ---------------------------------------------------------------------------
// ArtifactAction
// ---------------------------------------------------------------------------

/** Visual emphasis for a decision-sidebar action button. */
export type ArtifactActionVariant = "primary" | "secondary" | "danger";

/**
 * One contextual workflow action rendered in the approval decision sidebar.
 * Adapters emit only the actions that apply in the current state, in display
 * order. The shared sidebar renders each as a button, disables it when
 * `available` is false (surfacing `reason` as the disabled tooltip), and calls
 * `run` on click. `run` typically opens a route-owned dialog or fires a
 * mutation — the shell owns none of that behavior.
 */
export interface ArtifactAction {
  /** Stable key + test hook (e.g. "submit", "signoff", "publish", "createVersion"). */
  key: string;
  label: string;
  variant: ArtifactActionVariant;
  available: boolean;
  /** Shown as the disabled-button tooltip / inline reason when `available` is false. */
  reason?: string;
  run: () => void | Promise<void>;
}

// ---------------------------------------------------------------------------
// Author / submission meta (approval cockpit header row)
// ---------------------------------------------------------------------------

/**
 * The author / submission descriptor row rendered in the approval cockpit header
 * (avatar + author, area, submitted-ago, prior-approval). Every field is nullable
 * so the header degrades honestly when a fact is unavailable (Grade-A: omit, never
 * fabricate — e.g. documents expose no due-date, so there is no due-date field).
 */
export interface ArtifactAuthorMeta {
  /** Display name of the author/owner. Null when unknown. */
  authorName: string | null;
  /** Role descriptor for the author (e.g. "Autora"). Null when not applicable. */
  authorRole: string | null;
  /** Process-area label (e.g. "Qualidade"). Null for templates / when unknown. */
  areaLabel: string | null;
  /** Free-text submission descriptor (e.g. "Enviado para revisão há 3h"). Null when unknown. */
  submittedLabel: string | null;
  /** Prior-decision descriptor (e.g. "1ª aprovação por Bruno Said · 11:08"). Null when none. */
  priorDecisionLabel: string | null;
}

// ---------------------------------------------------------------------------
// Decision model (approval cockpit sidebar)
// ---------------------------------------------------------------------------

/** Visual tone of a decision option — approve (success) or return/reject (danger). */
export type ArtifactDecisionTone = "approve" | "reject";

/**
 * One selectable decision in the cockpit decision panel (e.g. "Aprovar",
 * "Devolver com ajuste"). Kind-agnostic: documents map these to signoff
 * approve/reject, templates to review/approve accept/reject.
 */
export interface ArtifactDecisionOption {
  /** Stable key + test hook (e.g. "approve", "reject"). */
  key: string;
  /** Radio-card title (e.g. "Aprovar"). */
  label: string;
  /** Radio-card description line (e.g. "Congela a versão · dispara fanout PDF · publica"). */
  description: string;
  tone: ArtifactDecisionTone;
  /** Submit-button label when this option is selected (e.g. "Assinar e aprovar"). */
  submitLabel: string;
  /** When true, the justificativa is required before submit is enabled. */
  requiresReason: boolean;
}

/**
 * Signer identity block. Shows the REAL current-user identity; the cryptographic
 * artifacts (hash/timestamp/IP) are produced server-side at signing and are not
 * fabricated here — `note` carries an honest "gerado no ato" statement instead.
 */
export interface ArtifactDecisionSigner {
  /** Real signer display name. */
  name: string;
  /** Secondary identity line (e.g. "Gerente de Qualidade · email"). Null when unknown. */
  detail: string | null;
  /** Honest note replacing fabricated signature metadata. Null to omit. */
  note: string | null;
}

/**
 * Normalized decision surface for the cockpit sidebar. Present only in states
 * that require a multi-option decision (document under_review sign; template
 * under_review review / approved approve). Absent states fall back to the plain
 * `actions[]` + a route-supplied extras slot.
 *
 * The shared `ArtifactDecisionPanel` owns all local UI state (selected option,
 * reason, password, legal-checkbox, in-flight/error) and calls `submit`. The
 * adapter/route owns what `submit` does (lifted signoff hook, or template
 * review/approve mutation) and surfaces user-facing failures by rejecting with
 * an Error whose message the panel renders.
 */
export interface ArtifactDecisionModel {
  /** Sidebar kicker (e.g. "§ Decisão requerida"). */
  kicker: string;
  /** Sidebar heading (e.g. "Sua aprovação"). */
  heading: string;
  /** Context copy under the heading. */
  description: string;
  /** Ordered decision options (>= 1). */
  options: ArtifactDecisionOption[];
  reasonLabel: string;
  reasonPlaceholder: string;
  /** Legal e-signature password field (documents). Null = no password (templates). */
  password: { label: string } | null;
  /** Legal-effect confirmation checkbox (documents). Null = none (templates). */
  legal: { text: string } | null;
  /** Signer identity block. Null to omit. */
  signer: ArtifactDecisionSigner | null;
  /** Option preselected on mount (e.g. from a `?decision=` deep-link). Null = none. */
  defaultOptionKey?: string | null;
  /**
   * Perform the decision. Resolves on success (the panel then relies on query
   * invalidation to re-render the next lifecycle state); rejects with an Error
   * whose `.message` the panel shows inline.
   */
  submit: (input: { optionKey: string; reason: string; password: string }) => Promise<void>;
}

// ---------------------------------------------------------------------------
// ArtifactViewModel — root normalized view-model
// ---------------------------------------------------------------------------

/**
 * The single, kind-agnostic view-model fed to every shared controlled-artifact
 * component. Adapters produce this; the shared shell consumes it.
 */
export interface ArtifactViewModel {
  kind: ArtifactKind;
  id: string;
  /**
   * Regulated document code (e.g. "DC-RH-001"). Null for templates, which have
   * no regulated code — only a slug key.
   */
  code: string | null;
  title: string;
  status: LifecycleStatus;
  versionNumber: number;
  /**
   * Formatted revision label from `formatRevisionCode` (e.g. "REV02"). Null for
   * templates or first-version documents with no prior revision.
   */
  revisionLabel: string | null;
  /** Hero block data: breadcrumb, badges, optional subtitle. */
  hero: ArtifactHeroModel;
  /** Profile / visibility / file metadata for the sidebar. */
  meta: ArtifactMetaModel;
  /** KPI strip cells, fully composed by the adapter. The shared view renders these
   *  in order with zero kind awareness. */
  kpis: ArtifactKpiCell[];
  /**
   * Ordered list of approval chain signoff slots. Both kinds populate this — the
   * document adapter from `ApprovalInstance.stages[]`, the template adapter from
   * the inline submit→review→approve fields (`buildTemplateApprovalChain`). Null
   * only when no version/instance is loaded yet (the base template detail adapter).
   */
  approvalChain: ApprovalChainItem[] | null;
  /**
   * Revision history entries. Never null; use an empty array when no history is
   * available or the kind does not expose a version-list endpoint.
   */
  lineage: VersionHistoryItem[];
  /**
   * Navigation tabs for the detail view. Documents: "Documento" + "Distribuição".
   * Templates: "Documento" only.
   */
  tabs: ArtifactTab[];
  /**
   * Ordered list of contextual workflow actions for the approval decision sidebar.
   * Empty when no actions apply (e.g. read-only states) — the detail view does not
   * render these. In the cockpit these render as the fallback when `decision` is
   * absent (e.g. draft submit-for-review, approved publish/schedule).
   */
  actions: ArtifactAction[];
  /**
   * Author / submission descriptor for the approval cockpit header. Optional —
   * the detail view ignores it; the approval screen renders it when present.
   */
  authorMeta?: ArtifactAuthorMeta;
  /**
   * Multi-option decision surface for the approval cockpit sidebar. Present only
   * in states that require a review/approve decision. When absent, the cockpit
   * renders the plain `actions[]` + the route's decision-extras slot instead.
   */
  decision?: ArtifactDecisionModel;
}
