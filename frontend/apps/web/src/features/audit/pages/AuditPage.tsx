import { OperationsCenter } from "../../../components/OperationsCenter";
import { useWorkspaceRouteContext } from "../../shell/pages/WorkspaceRoot";

export function Component() {
  const workspace = useWorkspaceRouteContext();

  return (
    <OperationsCenter
      loadState={workspace.loadState}
      documents={workspace.documents}
      notifications={workspace.notifications}
      documentProfiles={workspace.documentProfiles}
      processAreas={workspace.processAreas}
      formatDate={workspace.formatDate}
      onRefreshWorkspace={workspace.onRefreshWorkspace}
      onOpenDocument={workspace.onOpenDocument}
    />
  );
}
