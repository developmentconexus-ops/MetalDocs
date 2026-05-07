// WorkspaceRoot is being replaced by AppRoot + AppShell in Foundation Group C.
// This stub keeps the build green while the new shell components are built.
// See: wiki/implementation/plan-foundation.md — Tasks C1 / C2

import { Outlet } from "react-router-dom";
import type { DocumentProfileItem, NotificationItem, ProcessAreaItem, SearchDocumentItem } from "../../../lib/types";

type LoadState = "idle" | "loading" | "ready" | "error";

export type WorkspaceRouteContext = Record<string, any>;

type WorkspaceRouteDefaults = {
  loadState: LoadState;
  documents: SearchDocumentItem[];
  notifications: NotificationItem[];
  documentProfiles: DocumentProfileItem[];
  processAreas: ProcessAreaItem[];
  formatDate: (value?: string) => string;
  onRefreshWorkspace: () => void | Promise<void>;
  onOpenDocument: (documentId: string) => void | Promise<void>;
};

const noop = () => {};
const fmtDate = (value?: string) => (value ? new Date(value).toLocaleDateString("pt-BR") : "");

/**
 * Temporary stub — returns empty defaults so consumer pages render an empty
 * state instead of crashing on `.length` / `.map` of undefined. Real data
 * wiring lands when each feature owns its own TanStack Query hooks per the
 * canonical structure (wiki/architecture/frontend-structure.md).
 */
export function useWorkspaceRouteContext(): WorkspaceRouteContext {
  const defaults: WorkspaceRouteDefaults = {
    loadState: "ready",
    documents: [],
    notifications: [],
    documentProfiles: [],
    processAreas: [],
    formatDate: fmtDate,
    onRefreshWorkspace: noop,
    onOpenDocument: noop,
  };
  return defaults;
}

export function Component() {
  return <Outlet />;
}

export default Component;
