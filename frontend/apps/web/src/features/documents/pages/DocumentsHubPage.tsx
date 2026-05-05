import type { WorkspaceView } from "../../../components/DocumentWorkspaceShell";
import { DocumentsHubView } from "../DocumentsHubView";
import { useWorkspaceRouteContext } from "../../shell/pages/WorkspaceRoot";

type DocumentsHubPageProps = {
  view: Extract<WorkspaceView, "library" | "my-docs" | "recent">;
};

function DocumentsHubPage({ view }: DocumentsHubPageProps) {
  const workspace = useWorkspaceRouteContext();

  return (
    <DocumentsHubView
      view={view}
      loadState={workspace.loadState}
      documentProfiles={workspace.documentProfiles}
      processAreas={workspace.processAreas}
      documents={workspace.visibleDocuments}
      managedUsers={workspace.managedUsers}
      selectedDocument={workspace.selectedDocument}
      selectedProfileGovernance={workspace.selectedProfileGovernance}
      searchQuery={workspace.searchQuery}
      currentUserId={workspace.currentUserId}
      formatDate={workspace.formatDate}
      onSearchQueryChange={workspace.onSearchQueryChange}
      onCreateDocument={workspace.onCreateDocument}
      onRefreshDocuments={workspace.onRefreshWorkspace}
      onOpenDocument={workspace.onOpenDocument}
      onOpenDocumentForHub={workspace.onOpenDocumentForHub}
    />
  );
}

export function LibraryDocumentsPage() {
  return <DocumentsHubPage view="library" />;
}

export function MyDocumentsPage() {
  return <DocumentsHubPage view="my-docs" />;
}

export function RecentDocumentsPage() {
  return <DocumentsHubPage view="recent" />;
}
