import { Navigate, useLocation, useParams, type RouteObject } from "react-router-dom";

// F2d.5 S3 — the approval cockpit route is retired (ADR 0080); the worklist
// now forwards into the single mode-adaptive workspace at the canonical
// artifact path, preserving `location.search` (e.g. `?decision=`) since
// `<Navigate to="...">` does not do this for us.
function RedirectApprovalToWorkspace() {
  const { documentId } = useParams<{ documentId: string }>();
  const location = useLocation();
  return <Navigate to={`/documents/${documentId}${location.search}`} replace />;
}

export const approvalRoutes: RouteObject[] = [
  {
    path: "approvals",
    handle: { workspaceView: "approvals" },
    lazy: () => import("./pages/InboxPage").then((module) => ({ Component: module.InboxPage })),
  },
  {
    path: "approvals/:documentId",
    element: <RedirectApprovalToWorkspace />,
  },
  {
    path: "approval-routes",
    handle: { workspaceView: "approval-routes", requiresCapability: "route.manage" },
    lazy: () =>
      import("./pages/route-admin/RouteAdminPage").then((module) => ({
        Component: module.RouteAdminPage,
      })),
  },
];
