// WorkspaceRoot is being replaced by AppRoot + AppShell in Foundation Group C.
// This stub keeps the build green while the new shell components are built.
// See: wiki/implementation/plan-foundation.md — Tasks C1 / C2

import { Outlet } from "react-router-dom";

export type WorkspaceRouteContext = Record<string, unknown>;

/** Temporary stub — will be removed when AppRoot takes over routing. */
export function useWorkspaceRouteContext(): WorkspaceRouteContext {
  return {};
}

export function Component() {
  return <Outlet />;
}

export default Component;
