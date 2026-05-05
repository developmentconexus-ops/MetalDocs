import { RegistryExplorerView } from "../RegistryExplorerView";
import { useWorkspaceRouteContext } from "../../shell/pages/WorkspaceRoot";

export function Component() {
  const workspace = useWorkspaceRouteContext();

  return (
    <RegistryExplorerView
      loadState={workspace.loadState}
      documentProfiles={workspace.documentProfiles}
      processAreas={workspace.processAreas}
      subjects={workspace.subjects}
      selectedProfileCode={workspace.documentForm.documentProfile}
      selectedProfileSchema={workspace.selectedProfileSchema}
      selectedProfileSchemas={workspace.selectedProfileSchemas}
      selectedProfileGovernance={workspace.selectedProfileGovernance}
      showAdmin={workspace.isAdmin}
      onRefreshWorkspace={workspace.onRefreshWorkspace}
      onSelectProfile={(profileCode) => workspace.onApplyDocumentProfile(profileCode, workspace.documentForm.processArea)}
      onCreateProcessArea={workspace.onCreateProcessArea}
      onUpdateProcessArea={workspace.onUpdateProcessArea}
      onDeleteProcessArea={workspace.onDeleteProcessArea}
      onCreateSubject={workspace.onCreateSubject}
      onUpdateSubject={workspace.onUpdateSubject}
      onDeleteSubject={workspace.onDeleteSubject}
      onCreateDocumentProfile={workspace.onCreateDocumentProfile}
      onUpdateDocumentProfile={workspace.onUpdateDocumentProfile}
      onDeleteDocumentProfile={workspace.onDeleteDocumentProfile}
      onUpdateDocumentProfileGovernance={workspace.onUpdateDocumentProfileGovernance}
      onUpsertDocumentProfileSchema={workspace.onUpsertDocumentProfileSchema}
      onActivateDocumentProfileSchema={workspace.onActivateDocumentProfileSchema}
    />
  );
}
