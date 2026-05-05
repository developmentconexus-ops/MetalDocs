import { NotificationsPanel } from "../../../components/NotificationsPanel";
import { useWorkspaceRouteContext } from "../../shell/pages/WorkspaceRoot";

export function Component() {
  const workspace = useWorkspaceRouteContext();

  return (
    <NotificationsPanel
      loadState={workspace.loadState}
      notifications={workspace.notifications}
      formatDate={workspace.formatDate}
      onRefreshWorkspace={workspace.onRefreshWorkspace}
      onMarkRead={workspace.onMarkNotificationRead}
    />
  );
}
