import { ContentBuilderView } from "../../../components/content-builder/ContentBuilderView";
import { useWorkspaceRouteContext } from "../../shell/pages/WorkspaceRoot";

export function Component() {
  const workspace = useWorkspaceRouteContext();

  return (
    <ContentBuilderView
      document={workspace.selectedDocument}
      onBack={workspace.onBackToCreate}
      currentUserId={workspace.currentUserId}
    />
  );
}
