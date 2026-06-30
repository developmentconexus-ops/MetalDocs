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
 * A single signoff slot within one stage of the approval chain.
 *
 * Generalizes `ApprovalInstance.stages[].signoffs[]`. `approvalChain` on the
 * root view-model is `null` for templates, which have no instance/signoff model.
 */
export interface ApprovalChainItem {
  /** Zero-based index of the parent approval stage. */
  stageIndex: number;
  /** Human-readable stage label (e.g. "Revisão", "Aprovação"). */
  label: string;
  /** Stage-level status string (mirrors the API `stage.status` values). */
  status: string;
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
// ArtifactActionSet
// ---------------------------------------------------------------------------

/**
 * The complete set of workflow actions available on an artifact. Every key is
 * always present; availability is expressed via `available` + optional `reason`,
 * not by omitting keys. This keeps the decision sidebar dumb — it renders all
 * known actions and shows a disabled reason when `available` is false.
 *
 * The adapter owns backend specifics (ETag/If-Match for documents, reviewer /
 * approver role for templates). These types do not leak those details.
 */
export type ArtifactActionKind =
  | "submit"
  | "review"
  | "approve"
  | "reject"
  | "publish"
  | "createVersion";

export interface ArtifactAction {
  available: boolean;
  /** Human-readable reason shown when `available` is false (e.g. "Documento já aprovado"). */
  reason?: string;
  run: () => Promise<void>;
}

export type ArtifactActionSet = Record<ArtifactActionKind, ArtifactAction>;

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
   * Ordered list of approval chain signoff slots. Null for templates (which have
   * no approval instance / signoff model — only pending reviewer/approver roles).
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
  /** Full action set; every key always present, gated by `available`. */
  actions: ArtifactActionSet;
}
